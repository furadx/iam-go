# iam-go 快速开始指南

## 5 分钟快速上手

### 1. 环境准备

确保已安装：
- Go 1.25+
- PostgreSQL 12+

### 2. 进入项目目录

```bash
cd iam-go
```

### 3. 初始化数据库

```bash
# 创建数据库
createdb iam

# 或使用 psql
psql -U postgres
CREATE DATABASE iam;
\q

# 初始化表结构
psql -U postgres -d iam -f scripts/init.sql
```

### 4. 启动服务

```bash
# 安装依赖
go mod tidy

# 编译
go build -o bin/apiserver cmd/apiserver/main.go

# 运行
./bin/apiserver
```

服务启动成功后会输出：
```
服务启动在 :8080
```

### 5. 测试 API

**创建用户：**

```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "name": "alice",
    "nickname": "Alice",
    "password": "IamGo2026!Aa",
    "email": "alice@example.com",
    "phone": "13800138001"
  }'
```

**响应：**
```json
{
  "code": 0,
  "message": "OK",
  "data": {
    "id": 3,
    "name": "alice",
    "nickname": "Alice",
    "email": "alice@example.com",
    "phone": "13800138001",
    "status": 1,
    "createdAt": "2026-06-02T17:30:00Z"
  }
}
```

**获取用户列表：**

```bash
curl http://localhost:8080/api/v1/users?limit=10
```

**获取用户详情：**

```bash
curl http://localhost:8080/api/v1/users/alice
```

## 配置修改

编辑 `configs/config.yaml`，修改 database 配置：

```yaml
database:
  host: localhost
  port: 5432
  user: postgres
  password: YOUR_PASSWORD
  dbname: iam
  sslmode: disable
```

## 常见问题

### Q: 编译失败？
A: 运行 `go mod tidy` 确保所有依赖已安装。

### Q: 连接数据库失败？
A: 检查 PostgreSQL 是否运行，数据库是否已创建，连接字符串是否正确。

### Q: 端口被占用？
A: 修改 `configs/config.yaml` 中的 `server.addr` 为其他端口。

## 项目结构说明

```
cmd/apiserver/          # 主程序入口
internal/apiserver/     # API 服务器代码
  ├── controller/       # 控制器（处理 HTTP 请求）
  ├── service/          # 业务逻辑
  ├── store/            # 数据访问
  └── model/            # 数据模型
pkg/                    # 公共库
scripts/                # 脚本文件
```

## 下一步

- 查看完整 API 文档：`README.md`
- 了解架构设计：参考 [marmotedu/iam](https://github.com/marmotedu/iam)
- 扩展功能：添加更多业务逻辑到 Service 层

## 停止服务

按 `Ctrl+C` 停止服务，会自动触发优雅关闭。
