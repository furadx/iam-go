# 项目开发规范

本文档约定 iam-go 作为二次开发底座时的工程规范。新增功能、重构和修复问题时，应优先遵守本文档；若确实需要例外，应在代码评审或设计文档中说明原因。

## 1. 分层职责

项目按入口、应用层、内部基础能力、公共包分层：

```text
cmd/apiserver                    程序入口，负责配置加载、依赖初始化、优雅关闭
internal/apiserver/router.go     路由装配，不承载业务逻辑
internal/apiserver/controller    HTTP 入参绑定、调用 service、写响应
internal/apiserver/service       业务规则、字段白名单、副作用编排
internal/apiserver/store         数据访问接口
internal/apiserver/store/postgres PostgreSQL 实现
internal/apiserver/model         数据模型与选项结构
internal/apiserver/options       配置项、默认值、校验、命令行 flags
internal/pkg                     仅本项目内部使用的通用能力
pkg                              可被项目外复用的基础包
configs                          配置文件与规则模型
scripts                          初始化脚本、维护脚本
docs                             项目文档
```

约定：

- Controller 不直接访问数据库。
- Controller 不写复杂业务规则，只负责 HTTP 语义。
- Service 承担业务校验、领域规则、调用 store、副作用编排。
- Store 只处理数据访问与数据库错误翻译。
- `internal/pkg` 可依赖 `pkg`，但 `pkg` 不依赖 `internal`。
- `cmd/apiserver` 只做依赖组装，不承载请求级业务逻辑。

## 2. API 响应规范

普通业务接口统一使用 `internal/pkg/util.WriteResponse`。

成功响应：

```json
{
  "code": 0,
  "message": "OK",
  "data": {},
  "request_id": "..."
}
```

失败响应：

```json
{
  "code": 100001,
  "message": "参数绑定失败",
  "request_id": "..."
}
```

约定：

- 业务 Controller 使用 `WriteResponse` 返回响应。
- 中间件可以直接写 HTTP status，但响应体仍应包含 `code` 和 `message`。
- 不在 Controller 内拼多套响应格式。
- `data` 字段只放业务数据，不放错误描述。
- 新增列表接口应返回结构化分页数据，不直接返回裸数组，除非该接口明确是小集合枚举。

当前项目采用 HTTP 200 + business code 的业务响应风格；Auth/Authz 中间件会使用 401/403/500 这类 HTTP status。后续若要统一为严格 REST status，需要单独成批调整。

## 3. 错误码规范

错误码集中维护在 `internal/pkg/code`。

编号区间：

```text
0      成功
100xxx 通用、数据库、认证类错误
110xxx 用户与权限类业务错误
```

约定：

- 新增业务错误必须先在 `code.go` 中定义常量和 message。
- 底层错误向上返回时使用 `code.WithCode(...)` 包装。
- 已知业务失败使用 `code.New(...)`。
- 不在 Controller 中直接返回原始数据库错误、Redis 错误、JWT 错误。
- `pkg` 包内不依赖 `internal/pkg/code`，应返回哨兵错误或普通 error，由上层翻译。

示例：

```go
if err := c.ShouldBindJSON(&r); err != nil {
    util.WriteResponse(c, code.WithCode(code.ErrBind, err), nil)
    return
}
```

## 4. 日志规范

统一使用 `pkg/log`，避免直接混用标准库 `log`，除非日志系统尚未初始化。

约定：

- 启动早期配置解析失败可以使用标准库 `log`。
- 依赖初始化完成后使用 `pkg/log`。
- Controller 避免记录无意义的进入函数日志。
- Service 可记录业务副作用失败，例如默认角色分配失败、登录时间更新失败。
- Store 层一般不打日志，只返回错误，由上层决定是否记录。
- 日志中不得输出密码、token、authorization header、数据库密码、Redis 密码。

推荐日志内容：

```text
request_id
user_id / username
operation
resource
latency
error
```

## 5. 配置规范

配置项集中放在 `internal/apiserver/options`。

每类配置应包含：

```text
Options struct
NewXxxOptions 默认值
Validate 校验
AddFlags 命令行参数
```

约定：

- 新增配置必须支持配置文件和命令行 flag。
- 必填配置必须在 `Validate` 中校验。
- 生产敏感配置不得使用危险默认值。
- Secret、密码、token 不应出现在 JSON 输出中，结构体 tag 应使用 `json:"-"`。
- 默认值应适合本地开发，生产配置必须通过环境或配置文件覆盖。

## 6. 认证与授权规范

认证使用 JWT 双 token：

```text
access token  访问业务接口
refresh token 换取新的 access + refresh
```

约定：

- 业务接口只接受 access token。
- refresh 接口只接受 refresh token。
- refresh token 使用一次后必须吊销旧 token。
- logout 应吊销当前 access token；若传入 refresh token，也应吊销 refresh token。
- 修改密码后应吊销该用户旧 token。
- Redis 吊销存储异常时，不应继续签发新的 refresh token。

授权使用 Casbin RBAC：

```text
用户 -> 角色 -> 策略
```

约定：

- Auth 中间件负责识别用户。
- Authz 中间件负责判断权限。
- Authz 必须位于 Auth 之后。
- Controller 不直接判断角色名，统一交给 Casbin。
- Casbin model 存放在 `configs/rbac_model.conf`。
- Casbin policy 存放在数据库 `casbin_rule` 表。

## 7. 数据访问规范

Store 层负责数据库访问。

约定：

- Service 不直接引用 GORM。
- Store 接口定义在 `internal/apiserver/store`。
- PostgreSQL 实现放在 `internal/apiserver/store/postgres`。
- Store 层应翻译常见数据库错误，例如唯一键冲突、记录不存在。
- 更新操作优先使用字段白名单，避免 `Save` 全量覆盖带来的意外字段变更。
- 查询接口应支持分页上限，避免无界查询。

数据库初始化：

- `scripts/init.sql` 只作为本地和演示初始化脚本。
- 生产项目应引入版本化 migration 工具。
- Casbin `casbin_rule` 表由 GORM adapter 自动创建/迁移。

## 8. 测试规范

新增核心逻辑必须补测试。

最低要求：

- `pkg` 包新增逻辑必须有单元测试。
- `internal/pkg` 中间件、token、revoke、authz 必须有单元测试。
- Controller 中涉及认证、授权、token 流转的逻辑必须有 handler 测试。
- Service 中涉及业务规则、字段过滤、副作用的逻辑应有单元测试。
- 修复 bug 时应优先补一个能复现 bug 的测试，再修复。

提交前至少运行：

```bash
go test ./...
```

高风险改动建议额外覆盖：

```text
登录
refresh 轮转
logout
修改密码吊销
RBAC 策略增删
角色绑定/解绑
权限拒绝路径
```

## 9. 接口命名与路由规范

接口统一使用 `/api/v1` 前缀。

约定：

- 公开接口放在 `v1` 直接分组下。
- 仅需登录的接口放在 `authed` 分组。
- 需要权限控制的接口放在 `protected` 分组。
- REST 资源使用复数名词，例如 `/users`。
- 操作尽量通过 HTTP method 表达，不额外制造动词路径。

示例：

```text
POST   /api/v1/login
POST   /api/v1/refresh
POST   /api/v1/logout
GET    /api/v1/users
GET    /api/v1/users/:name
POST   /api/v1/users/:name/roles
DELETE /api/v1/users/:name/roles/:role
```

## 10. 文档规范

文档放在 `docs` 下。

```text
docs/CONVENTIONS.md        项目开发规范
docs/notes                 设计笔记、流程梳理
docs/superpowers/specs     功能设计文档
docs/superpowers/plans     实施计划
```

约定：

- 新增核心模块时，应补充设计说明或笔记。
- 认证、授权、数据迁移、部署、安全策略必须有文档。
- 文档中可以包含示例 token 名称，但不得写入真实 token、密码、密钥。

## 11. 代码风格

Go 代码：

- 提交前运行 `gofmt`。
- 包名使用小写短名。
- 接口命名表达能力，不加无意义前缀。
- 错误变量使用 `ErrXxx`。
- 常量按领域分组。
- 注释解释为什么，不重复描述显而易见的代码。

Controller 风格：

```go
func (ctl *Controller) Action(c *gin.Context) {
    var r request
    if err := c.ShouldBindJSON(&r); err != nil {
        util.WriteResponse(c, code.WithCode(code.ErrBind, err), nil)
        return
    }

    data, err := ctl.srv.Do(c.Request.Context(), r)
    if err != nil {
        util.WriteResponse(c, err, nil)
        return
    }

    util.WriteResponse(c, nil, data)
}
```

Service 风格：

```go
func (s *service) Do(ctx context.Context, input Input) (*Output, error) {
    if err := validate(input); err != nil {
        return nil, err
    }
    return s.store.Resource().Create(ctx, input)
}
```

## 12. 后续规范化任务

当前项目还建议继续补齐：

- 统一中间件响应体与 `WriteResponse` 的结构。
- 梳理是否继续坚持 HTTP 200 + business code，或切换为 REST status。
- 引入 migration 工具替代纯 `init.sql`。
- 移除无意义日志，例如简单的 controller entry log。
- 为 user service 补接口级测试。
- 为 RBAC controller 补 handler 测试。
- 为配置文件提供 dev/test/prod 示例。
- 建立 CI，至少执行 `go test ./...`。
