# IAM-Go

一个基于 Go 的 IAM（Identity and Access Management）框架，参考 [marmotedu/iam](https://github.com/marmotedu/iam) 项目架构设计。

## 特性

- 🏗️ **标准项目结构** - 遵循 Go 项目最佳实践
- 🔐 **用户管理** - 完整的用户 CRUD 操作
- 📦 **分层架构** - Controller -> Service -> Store 清晰分层
- 🗄️ **数据存储** - 基于 GORM 的 PostgreSQL 存储
- 🛡️ **密码加密** - bcrypt 加密存储
- 🚀 **Gin 框架** - 高性能 HTTP 路由

## 项目结构

```
iam-go/
├── cmd/
│   └── apiserver/        # API 服务器入口
│       └── main.go
├── internal/
│   ├── apiserver/
│   │   ├── controller/   # 控制器层
│   │   │   └── v1/
│   │   │       └── user/ # 用户控制器
│   │   ├── service/      # 业务逻辑层
│   │   │   └── v1/
│   │   ├── store/        # 存储层接口
│   │   │   └── postgres/ # PostgreSQL 实现
│   │   └── model/        # 数据模型
│   └── pkg/              # 内部公共库
│       ├── code/         # 错误码
│       ├── middleware/   # 中间件
│       └── util/         # 工具函数
├── pkg/                  # 可导出的公共库
│   ├── auth/            # 认证工具
│   └── log/             # 日志工具
└── scripts/             # 脚本
    └── init.sql         # 数据库初始化
```

## 快速开始

### 1. 环境准备

确保已安装：
- Go 1.25+
- PostgreSQL 12+

### 2. 数据库初始化

```bash
# 创建数据库
createdb iam

# 初始化表结构
psql -U postgres -d iam -f scripts/init.sql
```

### 3. 运行服务

```bash
# 安装依赖
go mod tidy

# 运行服务
go run cmd/apiserver/main.go
```

服务将启动在 `http://localhost:8080`

### 4. 测试 API

**创建用户：**

```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "name": "testuser",
    "nickname": "测试用户",
    "password": "123456",
    "email": "test@example.com",
    "phone": "13900139000"
  }'
```

**获取用户列表：**

```bash
curl http://localhost:8080/api/v1/users
```

**获取用户详情：**

```bash
curl http://localhost:8080/api/v1/users/testuser
```

## API 文档

### 用户接口

#### POST /api/v1/users - 创建用户

**请求体：**
```json
{
  "name": "username",      // 必填，唯一
  "nickname": "昵称",
  "password": "password",  // 必填
  "email": "email@example.com",  // 必填
  "phone": "13800138000"
}
```

**响应：**
```json
{
  "code": 0,
  "message": "OK",
  "data": {
    "id": 1,
    "name": "username",
    "nickname": "昵称",
    "email": "email@example.com",
    "createdAt": "2026-06-02T10:00:00Z"
  }
}
```

#### GET /api/v1/users - 获取用户列表

**查询参数：**
- `offset` - 偏移量
- `limit` - 每页数量（默认 20）
- `name` - 用户名模糊搜索

**响应：**
```json
{
  "code": 0,
  "message": "OK",
  "data": {
    "totalCount": 10,
    "items": [...]
  }
}
```

#### GET /api/v1/users/:name - 获取用户详情

**响应：**
```json
{
  "code": 0,
  "message": "OK",
  "data": {
    "id": 1,
    "name": "username",
    ...
  }
}
```

## 认证与授权

### JWT 双令牌

**登录**（公开）：
```bash
curl -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{"name":"alice","password":"your-password"}'
# 返回 { access_token, refresh_token, expires_in, user }
```

**访问受保护接口**：请求头带 `Authorization: Bearer <access_token>`。

**刷新**（access 过期后用 refresh 换新，旧 refresh 失效）：
```bash
curl -X POST http://localhost:8080/api/v1/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"<refresh_token>"}'
```

**登出**（吊销当前令牌，需登录）：
```bash
curl -X POST http://localhost:8080/api/v1/logout \
  -H "Authorization: Bearer <access_token>" \
  -d '{"refresh_token":"<refresh_token>"}'
```

吊销基于 Redis（`redis` 配置段）。`jwt.revoke_fail_open=true` 时 Redis 故障会放行并记日志。
Token 校验见 `pkg/token`，认证中间件见 `internal/pkg/middleware/auth.go`。

### RBAC（Casbin）

API 级鉴权：策略 `(角色, 路径, 方法)`，分组 `(用户, 角色)`，存于 Postgres 的 `casbin_rule` 表。
启动时默认 seed 一条 `admin → /api/v1/* → *`，新用户自动获得 `user` 角色。

**设首个管理员**（注册后执行一次）：
```sql
INSERT INTO casbin_rule (ptype, v0, v1) VALUES ('g', 'alice', 'admin');
```

**管理接口**（仅 admin）：
```bash
# 增加策略
curl -X POST http://localhost:8080/api/v1/authz/policies \
  -H "Authorization: Bearer <admin-access>" \
  -d '{"role":"editor","path":"/api/v1/users","method":"GET"}'

# 给用户分配角色
curl -X POST http://localhost:8080/api/v1/users/bob/roles \
  -H "Authorization: Bearer <admin-access>" \
  -d '{"role":"editor"}'
```

## 错误码

| 错误码 | 说明 |
|--------|------|
| 0 | 成功 |
| 100001 | 参数绑定失败 |
| 100002 | 参数验证失败 |
| 100003 | 服务器内部错误 |
| 100101 | 数据库错误 |
| 100202 | 签发 Token 失败 |
| 100203 | Token 无效 |
| 100204 | Token 已过期 |
| 100205 | 未授权 |
| 100206 | Token 类型错误 |
| 100207 | RefreshToken 无效 |
| 110001 | 用户不存在 |
| 110002 | 用户已存在 |
| 110003 | 密码错误 |
| 110006 | 用户已被禁用 |
| 110007 | 权限不足 |

## 配置

配置通过 `configs/config.yaml` 提供，启动时用 `-c` 指定：

```bash
go run cmd/apiserver/main.go -c configs/config.yaml
```

也支持命令行标志覆盖（如 `--db.host`、`--server.addr`、`--log.level`），
以及 `IAM_` 前缀的环境变量。完整选项见 `internal/apiserver/options/`。

```yaml
database:
  host: localhost
  port: 5432
  user: postgres
  password: postgres
  dbname: iam
  sslmode: disable
```

## 架构设计

### 分层架构

```
Controller (处理 HTTP 请求)
    ↓
Service (业务逻辑)
    ↓
Store (数据访问)
    ↓
Database (PostgreSQL)
```

### 设计模式

- **工厂模式** - Store Factory 管理存储层
- **依赖注入** - Controller 依赖 Service，Service 依赖 Store
- **接口抽象** - 所有层都定义接口，便于测试和扩展

## 参考项目

- [marmotedu/iam](https://github.com/marmotedu/iam) - IAM 系统设计参考

## 许可证

MIT License
