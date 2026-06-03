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

security:
  rate_limit:
    enabled: true
    api_limit: 300
    api_window_seconds: 60
    login_ip_limit: 10
    login_window_seconds: 60
    fail_open: true
  login_lock:
    enabled: true
    user_max_failures: 5
    ip_max_failures: 20
    failure_window_minutes: 15
    lock_minutes: 15
  password_policy:
    min_length: 12
    max_length: 64
    min_classes: 3
    reject_username: true
    reject_common_passwords: true
  cors:
    allowed_origins:
      - http://localhost:3000
      - http://localhost:5173
    allow_credentials: true
    max_age_seconds: 43200
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

### Security 配置

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| rate_limit.enabled | 是否启用 Redis 限流 | true |
| rate_limit.api_limit | 每 IP 每窗口 API 请求上限 | 300 |
| rate_limit.api_window_seconds | API 限流窗口秒数 | 60 |
| rate_limit.login_ip_limit | 每 IP 每窗口登录请求上限 | 10 |
| rate_limit.login_window_seconds | 登录限流窗口秒数 | 60 |
| rate_limit.fail_open | Redis 限流异常时是否放行 | true |
| login_lock.enabled | 是否启用登录失败锁定 | true |
| login_lock.user_max_failures | 同一用户名失败多少次后锁定 | 5 |
| login_lock.ip_max_failures | 同一 IP 失败多少次后锁定 | 20 |
| login_lock.failure_window_minutes | 失败计数窗口分钟数 | 15 |
| login_lock.lock_minutes | 锁定分钟数 | 15 |
| password_policy.min_length | 密码最小长度 | 12 |
| password_policy.max_length | 密码最大长度 | 64 |
| password_policy.min_classes | 最少字符类型数 | 3 |
| password_policy.reject_username | 禁止密码包含用户名 | true |
| password_policy.reject_common_passwords | 禁止常见弱密码 | true |
| cors.allowed_origins | CORS 允许的前端域名列表 | localhost dev origins |
| cors.allow_credentials | CORS 是否允许 credentials | true |
| cors.max_age_seconds | CORS preflight 缓存秒数 | 43200 |

## 优先级

配置加载优先级（从高到低）：

1. 环境变量（最高优先级）
2. 配置文件
3. 默认值（最低优先级）
