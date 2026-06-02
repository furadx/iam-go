# 配置说明

## 配置文件方式

创建 `configs/config.yaml`：

```yaml
server:
  mode: release  # debug, release, test
  addr: :8080

database:
  host: localhost
  port: 5432
  user: postgres
  password: postgres
  dbname: iam
  sslmode: disable
```

启动服务：

```bash
./bin/apiserver -config configs/config.yaml
```

## 环境变量方式

设置环境变量：

```bash
export IAM_SERVER_MODE=release
export IAM_SERVER_ADDR=:8080
export IAM_DB_HOST=localhost
export IAM_DB_USER=postgres
export IAM_DB_PASSWORD=your_password
export IAM_DB_NAME=iam
export IAM_DB_SSLMODE=disable
```

启动服务：

```bash
./bin/apiserver
```

## 配置项说明

### Server 配置

| 配置项 | 说明 | 默认值 | 环境变量 |
|--------|------|--------|----------|
| mode | Gin 运行模式 | release | IAM_SERVER_MODE |
| addr | 监听地址 | :8080 | IAM_SERVER_ADDR |

### Database 配置

| 配置项 | 说明 | 默认值 | 环境变量 |
|--------|------|--------|----------|
| host | 数据库主机 | localhost | IAM_DB_HOST |
| port | 数据库端口 | 5432 | IAM_DB_PORT |
| user | 数据库用户 | postgres | IAM_DB_USER |
| password | 数据库密码 | postgres | IAM_DB_PASSWORD |
| dbname | 数据库名称 | iam | IAM_DB_NAME |
| sslmode | SSL 模式 | disable | IAM_DB_SSLMODE |

## 优先级

配置加载优先级（从高到低）：

1. 环境变量（最高优先级）
2. 配置文件
3. 默认值（最低优先级）
