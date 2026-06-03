# HTTP Regression Test Checklist

这份清单用于较大改动后的真实 HTTP 回归测试。目标是启动实际 `apiserver`，用请求验证核心链路，而不是只依赖单元测试。

## 1. 准备环境

需要本机或容器中有：

- PostgreSQL，数据库名 `iam`
- Redis
- Go 1.25+

初始化数据库：

```bash
createdb iam
psql -U postgres -d iam -f scripts/init.sql
```

建议复制一份临时配置，避免污染默认配置：

```bash
cp configs/config.yaml .tmp-http-test.yaml
```

测试用种子账号：

- 用户名：`admin`
- 密码：`IamGo2026!Aa`

## 2. 启动服务

```bash
go run cmd/apiserver/main.go -c .tmp-http-test.yaml
```

如果 `:8080` 被占用，把 `.tmp-http-test.yaml` 的 `server.addr` 改成其他端口，例如 `:18080`。

以下示例默认：

```bash
BASE=http://localhost:8080
```

## 3. 健康检查

```bash
curl -i "$BASE/healthz"
```

期望：

- HTTP 200
- JSON 包含 `"status":"ok"`

## 4. CORS

允许的 Origin：

```bash
curl -i "$BASE/healthz" -H "Origin: http://localhost:5173"
```

期望：

- `Access-Control-Allow-Origin: http://localhost:5173`
- `Access-Control-Allow-Credentials: true`

未配置的 Origin：

```bash
curl -i "$BASE/healthz" -H "Origin: https://evil.example.com"
```

期望：

- 不返回 `Access-Control-Allow-Origin`

## 5. 注册与密码策略

弱密码应被拒绝：

```bash
curl -s "$BASE/api/v1/users" \
  -H "Content-Type: application/json" \
  -d '{"name":"weak-user","nickname":"Weak","password":"password123","email":"weak@example.com"}'
```

期望：

- `code` 为密码过短或密码弱相关错误码

强密码应成功：

```bash
curl -s "$BASE/api/v1/users" \
  -H "Content-Type: application/json" \
  -d '{"name":"alice","nickname":"Alice","password":"Alice2026!Good","email":"alice@example.com"}'
```

期望：

- `code: 0`
- 响应中不包含明文密码

重复创建应失败：

```bash
curl -s "$BASE/api/v1/users" \
  -H "Content-Type: application/json" \
  -d '{"name":"alice","nickname":"Alice","password":"Alice2026!Good","email":"alice@example.com"}'
```

期望：

- `code` 为用户已存在

## 6. 登录、刷新、登出

登录成功：

```bash
LOGIN=$(curl -s "$BASE/api/v1/login" \
  -H "Content-Type: application/json" \
  -d '{"name":"admin","password":"IamGo2026!Aa"}')
echo "$LOGIN"
```

期望：

- `code: 0`
- `data.access_token` 存在
- `data.refresh_token` 存在

提取 token：

```bash
ACCESS=$(echo "$LOGIN" | jq -r '.data.access_token')
REFRESH=$(echo "$LOGIN" | jq -r '.data.refresh_token')
```

刷新 token：

```bash
curl -s "$BASE/api/v1/refresh" \
  -H "Content-Type: application/json" \
  -d "{\"refresh_token\":\"$REFRESH\"}"
```

期望：

- `code: 0`
- 返回新的 access token 和 refresh token

登出：

```bash
curl -s "$BASE/api/v1/logout" \
  -H "Authorization: Bearer $ACCESS"
```

期望：

- `code: 0`

登出后旧 access token 访问受保护接口：

```bash
curl -s "$BASE/api/v1/users" \
  -H "Authorization: Bearer $ACCESS"
```

期望：

- 未授权或 token 无效相关错误

## 7. 登录失败锁定

连续 5 次错误密码：

```bash
for i in 1 2 3 4 5; do
  curl -s "$BASE/api/v1/login" \
    -H "Content-Type: application/json" \
    -d '{"name":"admin","password":"Wrong2026!Aa"}'
  echo
done
```

第 6 次即使用正确密码也应被锁定：

```bash
curl -s "$BASE/api/v1/login" \
  -H "Content-Type: application/json" \
  -d '{"name":"admin","password":"IamGo2026!Aa"}'
```

期望：

- `code` 为登录锁定
- 锁定时间为 15 分钟

## 8. 身份枚举保护

不存在用户：

```bash
curl -s "$BASE/api/v1/login" \
  -H "Content-Type: application/json" \
  -d '{"name":"not-exist","password":"Wrong2026!Aa"}'
```

禁用用户需要先在数据库中置为禁用：

```sql
UPDATE users SET status = 0 WHERE name = 'user1';
```

```bash
curl -s "$BASE/api/v1/login" \
  -H "Content-Type: application/json" \
  -d '{"name":"user1","password":"IamGo2026!Aa"}'
```

期望：

- 两者对外都返回密码错误
- 不暴露“用户不存在”或“用户禁用”的区别

## 9. RBAC

无 token 访问用户列表：

```bash
curl -s "$BASE/api/v1/users"
```

期望：

- 未授权

admin token 访问用户列表：

```bash
curl -s "$BASE/api/v1/users" \
  -H "Authorization: Bearer $ACCESS"
```

期望：

- `code: 0`

普通用户 token 访问用户列表：

```bash
USER_LOGIN=$(curl -s "$BASE/api/v1/login" \
  -H "Content-Type: application/json" \
  -d '{"name":"alice","password":"Alice2026!Good"}')
USER_ACCESS=$(echo "$USER_LOGIN" | jq -r '.data.access_token')
curl -s "$BASE/api/v1/users" \
  -H "Authorization: Bearer $USER_ACCESS"
```

期望：

- 权限不足

## 10. 软删除

用 admin token 删除用户：

```bash
curl -s -X DELETE "$BASE/api/v1/users/alice" \
  -H "Authorization: Bearer $ACCESS"
```

期望：

- `code: 0`
- 数据库中 `users.deleted_at` 已写入
- 普通查询不再返回该用户
- `Unscoped` 物理删除只在服务层显式传入 `DeleteOptions{Unscoped:true}` 时发生

## 11. 回归通过标准

一次完整 HTTP 回归至少覆盖：

- 健康检查
- CORS allowlist 正反例
- 密码策略正反例
- 登录成功、登录失败、登录锁定
- refresh/logout/token 吊销
- RBAC admin/普通用户差异
- 软删除行为

任何一步失败，都先记录请求、响应、服务日志和相关配置，再进入修复。
