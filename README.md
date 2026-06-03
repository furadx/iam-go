# IAM-Go

一个基于 Go + Gin 的身份认证与权限管理（IAM）服务底座，开箱即用，适合作为后端项目的二次开发起点。

> 版本：**v1.0.0** ｜ 变更记录见 [CHANGELOG.md](CHANGELOG.md) ｜ 版本策略见 [docs/VERSIONING.md](docs/VERSIONING.md)

## 这个项目能做什么

- **登录认证**：JWT 双令牌（access + refresh），支持令牌刷新与登出吊销
- **权限控制**：基于 Casbin 的 RBAC（角色—权限），接口级（路径 + 方法）鉴权，策略存数据库、运行时可改
- **安全加固**：登录失败锁定、接口限流、密码强度策略、CORS 白名单、bcrypt 加密
- **用户管理**：注册、查询、列表、删除
- **工程规范**：分层清晰、错误码统一、配置可校验、CI 质量门禁

一句话：**认证授权这块最难啃的骨头已经做完且有测试，你可以专注写业务。**

## 技术栈

| 能力 | 选型 |
|------|------|
| Web 框架 | Gin |
| 数据库 | PostgreSQL（GORM） |
| 缓存/吊销 | Redis |
| 授权 | Casbin RBAC |
| 认证 | JWT（HS256，双令牌） |
| 密码 | bcrypt |

## 快速开始

### 1. 环境准备

- Go 1.25+
- PostgreSQL 12+
- Redis 6+

### 2. 初始化数据库

```bash
createdb iam
psql -U postgres -d iam -f scripts/init.sql
```

> `casbin_rule` 权限表由程序启动时自动创建，无需手工建表。

### 3. 配置

编辑 `configs/config.yaml`，至少确认数据库、Redis 与 JWT 密钥：

```yaml
server:
  mode: debug            # 本地开发用 debug；生产用 release
database:
  host: localhost
  port: 5432
  user: postgres
  password: postgres
  dbname: iam
redis:
  addr: localhost:6379
jwt:
  secret: change-me-in-production   # 生产(release)模式必须改成 >=32 字节的强密钥，否则拒绝启动
```

### 4. 启动

```bash
go run cmd/apiserver/main.go -c configs/config.yaml
```

服务启动在 `http://localhost:8080`。

## 认证与鉴权怎么用

```text
1. 注册 / 登录 → 拿到 access_token + refresh_token
2. 访问受保护接口时带上 Authorization: Bearer <access_token>
3. access_token 过期 → 用 refresh_token 换新的一对
4. 登出 → 令牌被吊销，立即失效
```

**注册并登录：**

```bash
# 注册（公开）
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"name":"alice","password":"IamGo2026!Aa","email":"alice@example.com"}'

# 登录，拿令牌
curl -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{"name":"alice","password":"IamGo2026!Aa"}'
```

登录响应：

```json
{
  "code": 0,
  "message": "OK",
  "data": {
    "access_token": "xxx",
    "refresh_token": "yyy",
    "expires_in": 900
  }
}
```

**访问受保护接口：**

```bash
curl http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer <access_token>"
```

> 注意：登录成功只代表"你是谁"。能不能访问某个接口由 Casbin 决定。新注册用户默认是 `user` 角色，默认没有任何接口权限；`admin` 角色默认拥有 `/api/v1/*` 全部权限。权限分配见下方接口表。

## 接口一览

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| POST | `/api/v1/users` | 公开 | 注册 |
| POST | `/api/v1/login` | 公开 | 登录，返回双令牌 |
| POST | `/api/v1/refresh` | 公开（带 refresh） | 刷新令牌（旧 refresh 作废） |
| POST | `/api/v1/logout` | 需登录 | 登出，吊销令牌 |
| GET | `/api/v1/users` | 需授权 | 用户列表 |
| GET | `/api/v1/users/:name` | 需授权 | 用户详情 |
| DELETE | `/api/v1/users/:name` | 需授权 | 删除用户 |
| GET/POST/DELETE | `/api/v1/users/:name/roles` | 需授权 | 查询/分配/撤销用户角色 |
| GET/POST/DELETE | `/api/v1/authz/policies` | 需授权 | 查询/新增/删除权限策略 |

所有业务接口统一返回 `{ code, message, data, request_id }`，`code=0` 表示成功。

## 开发

```bash
make ci          # 本地复现 CI 门禁：gofmt 检查 + go vet + build + test
make test        # 跑测试（带竞态检测）
make fmt         # 格式化代码
make build       # 编译二进制
```

提交前请确保 `make ci` 通过——CI 会在 PR 上执行同样的四道门禁。

## 项目结构

```
cmd/apiserver/              程序入口（配置加载、依赖装配、优雅关闭）
internal/apiserver/
  ├── controller/           HTTP 入参/出参
  ├── service/              业务逻辑
  ├── store/                数据访问（postgres 实现）
  ├── model/                数据模型
  └── options/              配置项与校验
internal/pkg/               项目内通用能力
  ├── authz/                Casbin 封装
  ├── middleware/           Auth / Authz / 限流 / CORS 等中间件
  ├── revoke/               JWT 吊销（Redis）
  ├── loginlock/            登录失败锁定
  ├── ratelimit/            限流
  └── password/             密码策略
pkg/                        可复用基础包（token / log / auth / ws）
configs/                    配置文件与 Casbin 模型
docs/                       开发规范、设计文档、版本策略
scripts/                    初始化脚本
```

## 文档

- [开发规范 CONVENTIONS.md](docs/CONVENTIONS.md) — 分层、错误码、日志、测试等工程约定
- [JWT 与权限流转笔记](docs/notes/jwt-authz-flow.md) — 令牌流转、吊销、Casbin 数据
- [变更记录 CHANGELOG.md](CHANGELOG.md)
- [版本策略 VERSIONING.md](docs/VERSIONING.md)

## 路线图（上生产前建议补齐）

- [ ] 反向代理可信网段（`SetTrustedProxies`），否则 IP 限流/登录锁可被 `X-Forwarded-For` 绕过
- [ ] 版本化数据库迁移工具替代 `init.sql`
- [ ] 真实就绪探针（探测 DB / Redis）
- [ ] 指标与链路追踪（Prometheus / OpenTelemetry）
- [ ] 审计日志

## 许可证

MIT License

## 参考

- [marmotedu/iam](https://github.com/marmotedu/iam) — 架构设计参考
