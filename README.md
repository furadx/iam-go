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

## 错误码

| 错误码 | 说明 |
|--------|------|
| 0 | 成功 |
| 100001 | 参数绑定失败 |
| 100002 | 参数验证失败 |
| 100101 | 数据库错误 |
| 110001 | 用户不存在 |
| 110002 | 用户已存在 |

## 配置

修改 `cmd/apiserver/main.go` 中的数据库连接字符串：

```go
dsn := "host=localhost user=postgres password=postgres dbname=iam port=5432 sslmode=disable"
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
