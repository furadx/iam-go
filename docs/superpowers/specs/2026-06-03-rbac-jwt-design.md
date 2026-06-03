# RBAC 权限控制 + JWT 优化 设计文档

- 日期：2026-06-03
- 状态：已通过设计评审，待 spec 复核
- 适用项目：iam-go（基于 Gin 的分层 Web API 脚手架）

## 1. 背景与目标

当前项目已有用户管理、JWT（单一 HS256 access token）、WebSocket。缺少**授权（Authorization）**能力，且 JWT 无刷新、无吊销。本设计补齐两块：

1. **RBAC 权限控制**：基于 Casbin 的角色-权限模型，API 级（路径+方法）鉴权，策略落 Postgres、运行时可改。
2. **JWT 优化**：Access + Refresh 双令牌，基于 Redis 黑名单的即时吊销（登出 / 改密踢全部会话）。

### 关键决策（来自设计问答）

| 决策点 | 选择 |
|--------|------|
| RBAC 实现 | Casbin + GORM 适配器 |
| 鉴权粒度 | API 级：路径 + HTTP 方法 |
| 多租户 | 否，单域标准 RBAC（sub, obj, act） |
| JWT | Access + Refresh 双令牌 |
| 吊销策略 | Redis 黑名单（即时吊销） |
| 签名算法 | 保持 HS256 |
| RBAC 管理 API | 保留整组，仅 admin 可访问 |

### 新增依赖

- `github.com/casbin/casbin/v2`
- `github.com/casbin/gorm-adapter/v3`
- `github.com/redis/go-redis/v9`
- 测试：`github.com/alicebob/miniredis/v2`

部署上新增 **Redis** 依赖（仅用于令牌吊销）。

## 2. RBAC 设计

### 2.1 Casbin 模型

`configs/rbac_model.conf`：

```ini
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && keyMatch2(r.obj, p.obj) && (r.act == p.act || p.act == "*")
```

- `sub` = 用户名（username），因此 JWT claims 必须携带 username。
- `obj` = 请求路径，`keyMatch2` 支持 `:param` 与 `*` 通配。
- `act` = HTTP 方法，`p.act == "*"` 表示该资源全方法放行。

策略示例：
```
p, admin, /api/v1/*, *
p, user, /api/v1/users/:name, GET
g, alice, admin
```

### 2.2 策略存储

GORM adapter 使用现有 Postgres 连接，自动建 `casbin_rule` 表。策略增删改后调用 `enforcer.LoadPolicy()` 热加载（gorm-adapter 的写操作会同步内存，无需重启）。

### 2.3 鉴权中间件 `Authz`

`internal/pkg/middleware/authz.go`：

```go
func Authz(e *casbin.Enforcer) gin.HandlerFunc
```

- 从上下文取 `username`（由 Auth 中间件写入）。
- 以 `(username, c.Request.URL.Path, c.Request.Method)` 调 `e.Enforce`。
- 通过则 `c.Next()`；否则 `403` + `ErrPermissionDenied`。
- 仅挂在需要授权的路由组上（公开接口如 login/refresh 不挂）。

中间件顺序：`Auth`（认证，确定你是谁）→ `Authz`（授权，确定你能不能）。

### 2.4 Enforcer 初始化

`internal/pkg/authz/authz.go`：封装 `NewEnforcer(db *gorm.DB, modelPath string) (*casbin.Enforcer, error)`，并提供策略/角色操作封装：

```go
type Manager struct { e *casbin.Enforcer }
func (m *Manager) AddPolicy(role, obj, act string) (bool, error)
func (m *Manager) RemovePolicy(role, obj, act string) (bool, error)
func (m *Manager) AssignRole(user, role string) (bool, error)
func (m *Manager) RevokeRole(user, role string) (bool, error)
func (m *Manager) RolesForUser(user string) ([]string, error)
func (m *Manager) Enforcer() *casbin.Enforcer
```

### 2.5 种子与默认角色

- 启动时若策略为空，seed：`p, admin, /api/v1/*, *`。首个管理员的 `g, <name>, admin` 通过运维脚本或首次手工设置（init.sql 注释说明）。
- `userService.Create` 成功后追加 `g, <newUserName>, user`，新用户默认 `user` 角色。
- `users.is_admin` 字段保留但不再作为权威授权依据（避免破坏现有模型）。

### 2.6 RBAC 管理 API（仅 admin）

挂在 `Auth` + `Authz` 之后（admin 策略覆盖）：

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/authz/policies` | 增加策略 `{role, path, method}` |
| DELETE | `/api/v1/authz/policies` | 删除策略 |
| GET | `/api/v1/authz/policies` | 列出全部策略 |
| POST | `/api/v1/users/:name/roles` | 给用户分配角色 `{role}` |
| DELETE | `/api/v1/users/:name/roles/:role` | 撤销角色 |
| GET | `/api/v1/users/:name/roles` | 查询用户角色 |

控制器：`internal/apiserver/controller/v1/authz/`。

## 3. JWT 优化设计

### 3.1 `pkg/token` 重构（保持不依赖 internal）

Claims：
```go
type Claims struct {
    UserID   int64  `json:"uid"`
    Username string `json:"username"`
    Type     string `json:"typ"` // "access" | "refresh"
    jwt.RegisteredClaims          // jti(ID), iat, exp, iss
}
```

Manager：
```go
func NewManager(secret string, accessExpire, refreshExpire time.Duration) *Manager
func (m *Manager) SignAccess(uid int64, username string) (token string, claims *Claims, err error)
func (m *Manager) SignRefresh(uid int64, username string) (token string, claims *Claims, err error)
func (m *Manager) Parse(tokenStr string) (*Claims, error)        // 校验签名/过期，返回 claims
func (m *Manager) ParseTyped(tokenStr, want string) (*Claims, error) // 额外校验 typ
```

- 每个令牌带唯一 `jti`（uuid）。
- 哨兵错误：`ErrInvalidToken`、`ErrTokenExpired`、`ErrWrongTokenType`。
- 防算法混淆：仅接受 HMAC 签名。

### 3.2 吊销组件（Redis）

接口（供 Auth 中间件依赖，便于替换/测试）：
```go
type Revoker interface {
    // Revoke 把单个令牌的 jti 拉黑，ttl 为该令牌剩余寿命。
    Revoke(ctx context.Context, jti string, ttl time.Duration) error
    // RevokeAllBefore 使某用户 iat 早于 t 的全部令牌失效。
    RevokeAllBefore(ctx context.Context, uid int64, t time.Time) error
    // Allowed 校验 claims 是否未被吊销（jti 未拉黑 且 iat >= revokeBefore）。
    Allowed(ctx context.Context, c *Claims) (bool, error)
}
```

Redis 实现 `internal/pkg/revoke/redis.go`：
- `jwt:bl:{jti} = "1"`，`SET ... EX ttl` → 单点吊销。
- `jwt:revoke_before:{uid} = unixSeconds` → 全局吊销，`claims.iat < value` 即失效。

### 3.3 Auth 中间件升级

`internal/pkg/middleware/auth.go`：
- `Auth(tm *token.Manager, rv revoke.Revoker)`。
- 解析 `Authorization: Bearer <access>` → `ParseTyped(.., "access")` → `rv.Allowed(claims)`。
- 通过则 `c.Set("user_id", uid)`、`c.Set("username", username)`。
- 错误映射（`pkg/token` 哨兵错误 → `code` 业务码）：`ErrTokenExpired`→`code.ErrTokenExpired`，`ErrWrongTokenType`→`code.ErrTokenTypeInvalid`，`ErrInvalidToken`→`code.ErrTokenInvalid`，被吊销→`code.ErrUnauthorized`。

### 3.4 接口与流程

| 方法 | 路径 | 鉴权 | 行为 |
|------|------|------|------|
| POST | `/api/v1/login` | 公开 | 校验密码 → 签发 `{access_token, refresh_token, expires_in, user}` |
| POST | `/api/v1/refresh` | 公开（带 refresh） | 校验 refresh 未吊销 → 签发新 access **并轮转 refresh**（旧 refresh jti 入黑名单） |
| POST | `/api/v1/logout` | 需登录 | 当前 access 的 jti 入黑名单；body 可选带 refresh_token 一并吊销 |
| —（改密） | `ChangePassword` 成功后 | — | `RevokeAllBefore(uid, now)` 踢掉该用户全部旧令牌 |

### 3.5 装配（main.go）

1. 初始化 Redis 客户端（`internal/pkg/redis`）。
2. 构建 `revoke.NewRedisRevoker(client)`。
3. 构建 `token.NewManager(secret, accessExpire, refreshExpire)`。
4. 构建 casbin enforcer（复用 gorm DB）。
5. `InitRouter(store, tm, revoker, enforcer)`。

## 4. 配置 / Options

```yaml
jwt:
  secret: change-me-in-production
  access_expire: 900       # 秒，15 分钟
  refresh_expire: 604800   # 秒，7 天
redis:
  addr: localhost:6379
  password: ""
  db: 0
```

- `options/jwt.go` 扩展（access/refresh 两个有效期 + validate）。
- 新增 `options/redis.go`（addr/password/db + validate + flags）。

## 5. 数据模型 / 迁移

- 新表 `casbin_rule`：由 gorm-adapter 自动迁移，`scripts/init.sql` 加注释说明（不手写该表）。
- `users` 表不变。
- 应用启动时执行 enforcer 初始化（含 AutoMigrate）+ 策略 seed。

## 6. 文件布局

```
configs/rbac_model.conf                         # 新增：casbin 模型
pkg/token/token.go                              # 重构：双令牌 + jti + typ
pkg/token/token_test.go                         # 扩充
internal/pkg/redis/redis.go                     # 新增：go-redis 客户端
internal/pkg/revoke/revoke.go                   # 新增：Revoker 接口
internal/pkg/revoke/redis.go                    # 新增：Redis 实现
internal/pkg/revoke/redis_test.go               # 新增：miniredis 单测
internal/pkg/authz/authz.go                     # 新增：enforcer + Manager 封装
internal/pkg/authz/authz_test.go                # 新增：内存 adapter 判定单测
internal/pkg/middleware/auth.go                 # 升级：吊销校验 + username 注入
internal/pkg/middleware/authz.go                # 新增：casbin 鉴权中间件
internal/apiserver/controller/v1/user/login.go  # 改：返回双令牌
internal/apiserver/controller/v1/user/refresh.go# 新增
internal/apiserver/controller/v1/user/logout.go # 新增
internal/apiserver/controller/v1/authz/*.go     # 新增：策略/角色管理
internal/apiserver/options/jwt.go               # 扩展
internal/apiserver/options/redis.go             # 新增
internal/apiserver/router.go                    # 改：装配新中间件与路由
internal/apiserver/service/v1/user.go           # 改：改密后 RevokeAllBefore；建用户 seed user 角色
cmd/apiserver/main.go                           # 改：装配 redis/enforcer/revoker
```

## 7. 错误码新增（`internal/pkg/code`）

| 常量 | 码 | 含义 |
|------|----|------|
| `ErrTokenTypeInvalid` | 100206 | Token 类型错误 |
| `ErrRefreshTokenInvalid` | 100207 | RefreshToken 无效 |
| `ErrPermissionDenied` | 110007 | 权限不足（HTTP 403） |

`WriteResponse` 目前统一返回 HTTP 200 + 业务码；鉴权类失败（401/403）改为返回对应 HTTP 状态码（Auth/Authz 中间件用 `AbortWithStatusJSON`，与现有 recovery 一致）。

## 8. 测试策略

- `pkg/token`：access/refresh 签发与解析、typ 误用拒绝、过期、错密钥、畸形。
- `internal/pkg/authz`：用 casbin 内存 adapter 加载典型策略，断言允许/拒绝（admin 通配、user 限定方法、未授权拒绝）。
- `internal/pkg/revoke`：miniredis 验证单点拉黑、全局 revoke_before、TTL 过期后放行。
- 中间件：Auth（无/错/过期/被吊销 token）与 Authz（有/无权限）用 httptest 走通。

## 9. 非目标（YAGNI）

- 不做多租户 / domains。
- 不做 RS256 / 密钥轮转。
- 不做 OAuth2 / 第三方登录。
- 不做权限的前端管理界面。
- 不做审计日志（可作为后续独立 spec）。

## 10. 风险与权衡

- **引入 Redis**：增加部署组件；若 Redis 不可用，吊销校验如何降级？设计取**fail-closed 可选**：默认 Redis 故障时放行（fail-open）以保可用性，并记 error 日志；是否改为 fail-closed 由配置项 `jwt.revoke_fail_open`（默认 true）控制。
- **casbin sub 用 username**：用户改名会使其策略失效。当前用户模型改名需走专用接口（现状未开放改名），风险可控；如未来开放改名，需同步迁移 casbin 策略。
- **HTTP 状态码变更**：鉴权失败从 200+码 改为 401/403，属于对现有"统一 200"约定的局部破例，仅限认证/授权中间件，业务接口仍保持 200+码。
