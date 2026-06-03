# JWT 与身份权限流转笔记

本文记录项目当前的 token 流转、Redis 吊销机制、Auth 身份认证、Authz 权限认证，以及 Casbin 默认数据。

## 1. 整体关系

项目里认证和授权分成两层：

```text
Auth  = 身份认证：确认你是谁
Authz = 权限认证：确认你能访问什么
```

请求进入受保护接口时的顺序：

```text
客户端请求
  -> Auth JWT 中间件
  -> 解析 access token，得到 user_id / username / claims
  -> Authz Casbin 中间件
  -> 用 username + path + method 判断是否有权限
  -> Controller
```

## 2. JWT 签发内容

JWT 由 `pkg/token.Manager` 签发，使用 HMAC-SHA256：

```text
alg = HS256
secret = jwt.secret
issuer = iam-go
```

自定义 claims：

```text
uid      用户 ID
username 用户名
typ      token 类型：access 或 refresh
```

标准 claims：

```text
jti 唯一 token ID，用于 Redis 黑名单吊销
iat 签发时间
exp 过期时间
iss 签发方，固定为 iam-go
```

默认有效期：

```text
access token  = 900 秒，15 分钟
refresh token = 604800 秒，7 天
```

## 3. Token 流转方式

项目采用双 token：

```text
access token  短期，用于访问业务接口
refresh token 长期，用于换取新的 access + refresh
```

### 登录

用户登录：

```http
POST /api/v1/login
```

请求体：

```json
{
  "name": "alice",
  "password": "IamGo2026!Aa"
}
```

服务端校验用户名、密码、用户状态后，签发：

```json
{
  "access_token": "access-A",
  "refresh_token": "refresh-A",
  "expires_in": 900
}
```

此时 Redis 不保存 token 本体。

### 访问接口

客户端访问需要登录的接口：

```http
GET /api/v1/users
Authorization: Bearer access-A
```

服务端校验：

```text
1. Authorization header 是否存在
2. token 格式是否是 Bearer
3. JWT 签名是否正确
4. token 是否过期
5. typ 是否是 access
6. Redis 中是否已被吊销
```

通过后，把这些值写入 gin context：

```text
user_id
username
claims
```

### Access token 过期

`access-A` 过期后，客户端使用 refresh token 换新 token：

```http
POST /api/v1/refresh
```

请求体：

```json
{
  "refresh_token": "refresh-A"
}
```

服务端流程：

```text
1. 校验 refresh-A 的签名、过期时间、typ=refresh
2. 查询 Redis，确认 refresh-A 没有被吊销
3. 把 refresh-A 的 jti 写入 Redis 黑名单
4. 签发 access-B + refresh-B
```

返回：

```json
{
  "access_token": "access-B",
  "refresh_token": "refresh-B",
  "expires_in": 900
}
```

之后客户端必须改用：

```text
access-B
refresh-B
```

旧的 `refresh-A` 已经被拉黑，不能再使用。

完整轮转形态：

```text
登录:
access-A + refresh-A

access-A 过期:
refresh-A -> access-B + refresh-B
refresh-A 被拉黑

access-B 过期:
refresh-B -> access-C + refresh-C
refresh-B 被拉黑
```

## 4. Redis 吊销机制

Redis 不保存 JWT 本体，只保存吊销信息。

### 单 token 黑名单

key 格式：

```text
jwt:bl:<jti>
```

示例：

```text
jwt:bl:refresh-111 = "1"
```

TTL 是该 token 的剩余有效期。token 本来过期后，黑名单 key 也会自动过期。

典型场景：

```text
refresh 轮转时，旧 refresh token 写入黑名单
logout 时，access token 写入黑名单
logout 带 refresh_token 时，refresh token 也写入黑名单
```

### 用户级全局吊销时间

key 格式：

```text
jwt:revoke_before:<uid>
```

示例：

```text
jwt:revoke_before:7 = 1780480000
```

含义：

```text
uid=7 的用户，iat 早于该时间戳的 token 全部失效
```

典型场景：

```text
用户修改密码后，吊销该用户之前签发的全部 token
```

### Redis 检查逻辑

每次认证 access token，或 refresh token 轮转前，都会检查：

```text
1. EXISTS jwt:bl:<jti>
   存在则拒绝

2. GET jwt:revoke_before:<uid>
   如果 token.iat 早于 revoke_before，则拒绝
```

## 5. Logout 流程

退出登录接口：

```http
POST /api/v1/logout
Authorization: Bearer access-B
```

如果请求体带 refresh token：

```json
{
  "refresh_token": "refresh-B"
}
```

服务端会吊销：

```text
access-B
refresh-B
```

写入 Redis：

```text
jwt:bl:<access-B 的 jti>
jwt:bl:<refresh-B 的 jti>
```

## 6. Auth 身份认证流程

Auth 中间件负责确认请求是谁发起的。

输入：

```http
Authorization: Bearer <access_token>
```

流程：

```text
1. 解析 Authorization header
2. 校验 JWT 签名
3. 校验 exp
4. 校验 typ=access
5. 查询 Redis 吊销状态
6. 写入 gin context
```

写入 context：

```text
ContextUserIDKey   = user_id
ContextUsernameKey = username
ContextClaimsKey   = claims
```

后续 Authz 中间件依赖 `username` 来做权限判断。

## 7. Authz 权限认证流程

Authz 中间件负责确认这个用户能不能访问当前接口。

它会取出：

```text
sub = username
obj = request path
act = request method
```

然后调用 Casbin：

```go
Enforce(username, path, method)
```

例如：

```http
GET /api/v1/users
```

实际判断：

```text
sub = alice
obj = /api/v1/users
act = GET
```

## 8. Casbin 规则模型

规则模型文件：

```text
configs/rbac_model.conf
```

请求定义：

```ini
r = sub, obj, act
```

策略定义：

```ini
p = sub, obj, act
```

角色定义：

```ini
g = _, _
```

匹配规则：

```ini
m = g(r.sub, p.sub) && keyMatch2(r.obj, p.obj) && (r.act == p.act || p.act == "*")
```

解释：

```text
用户属于某个角色
并且请求路径匹配策略路径
并且请求方法匹配策略方法，或策略方法为 *
```

## 9. Casbin 数据存储

Casbin 规则模型存在文件里：

```text
configs/rbac_model.conf
```

Casbin 权限数据存在数据库表：

```text
casbin_rule
```

该表由 Casbin GORM adapter 自动创建/迁移。

常见数据有两类：

### 权限策略

```text
ptype = p
v0    = role
v1    = path
v2    = method
```

示例：

```text
p, admin, /api/v1/*, *
```

含义：

```text
admin 角色可以访问 /api/v1/ 下所有接口和所有 HTTP 方法
```

### 角色绑定

```text
ptype = g
v0    = user
v1    = role
```

示例：

```text
g, alice, admin
```

含义：

```text
alice 拥有 admin 角色
```

## 10. 初始默认 Casbin 数据

应用启动时会执行默认种子逻辑。

新库首次启动后，`casbin_rule` 默认会有：

```text
ptype | v0    | v1        | v2
------+-------+-----------+---
p     | admin | /api/v1/* | *
g     | admin | admin     |
```

等价于：

```text
p, admin, /api/v1/*, *
g, admin, admin
```

含义：

```text
admin 角色拥有 /api/v1/* 全权限
初始化用户 admin 拥有 admin 角色
```

`init.sql` 默认创建：

```text
admin
user1
```

但默认只有 `admin` 会被绑定到 Casbin 的 `admin` 角色。

新注册用户会尝试自动绑定到 `user` 角色，但当前默认没有为 `user` 角色初始化权限策略。

## 11. 一个完整例子

假设：

```text
用户：admin
角色：admin
策略：admin 可以访问 /api/v1/* 的所有方法
```

数据库里：

```text
p, admin, /api/v1/*, *
g, admin, admin
```

请求：

```http
DELETE /api/v1/authz/policies
Authorization: Bearer <admin-access-token>
```

流程：

```text
1. Auth 校验 access token
2. 从 token 中取 username=admin
3. Auth 写入 context.username=admin
4. Authz 调用 Enforce("admin", "/api/v1/authz/policies", "DELETE")
5. Casbin 判断：
   admin 是否属于 admin 角色：是
   /api/v1/authz/policies 是否匹配 /api/v1/*：是
   DELETE 是否匹配 *：是
6. 放行
```

如果普通用户 `bob` 没有任何角色或策略：

```text
Auth 可以通过
Authz 会拒绝
```

也就是：

```text
token 有效，只代表你已登录
Casbin 放行，才代表你有权限访问这个接口
```
