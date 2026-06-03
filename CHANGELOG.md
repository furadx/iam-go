# Changelog

本项目所有重要变更都记录在本文件中。

格式遵循 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/)，
版本号遵循 [语义化版本 SemVer](https://semver.org/lang/zh-CN/)。版本策略见 [docs/VERSIONING.md](docs/VERSIONING.md)。

## [Unreleased]

（尚未发布的变更写在这里）

## [1.0.0] - 2026-06-03

首个稳定版本：可作为后端项目二次开发的认证授权底座。

### Added
- **用户管理**：注册、查询、列表（分页）、删除。
- **JWT 双令牌认证**：access + refresh 令牌；refresh 刷新时轮转（旧令牌作废）；登出吊销当前令牌。
- **Casbin RBAC 授权**：接口级（路径 + 方法）鉴权；策略存 PostgreSQL（`casbin_rule`），运行时可增删；新用户默认 `user` 角色，`admin` 默认拥有 `/api/v1/*` 全权限。
- **角色 / 策略管理接口**：`/api/v1/users/:name/roles`、`/api/v1/authz/policies`（仅授权用户）。
- **Redis 令牌吊销**：单点黑名单（按 jti）+ 用户级全局吊销（改密后踢掉旧令牌）。
- **安全加固**：登录失败锁定（按用户 / IP）、接口限流（全局 + 登录）、密码强度策略、CORS 白名单。
- **配置体系**：所有配置支持配置文件 + 命令行 flag，并带 `Validate` 校验。
- **CI 质量门禁**：`.github/workflows/ci.yml` 执行 gofmt 检查 / `go vet` / `go build` / `go test`；`make ci` 本地复现。
- **工程文档**：开发规范 `docs/CONVENTIONS.md`、JWT 与权限流转笔记、设计文档。

### Security
- **生产密钥防护**：`release` 模式下拒绝默认密钥与 <32 字节的弱 JWT 密钥，避免使用公开密钥签发可被伪造的令牌。
- **注册批量赋值防护**：注册改用专用 DTO，客户端无法再通过请求体设置 `is_admin` / `status` 等特权字段。
- **密码响应脱敏**：用户模型 `Password` 字段标记 `json:"-"`，任何响应都不再泄漏密码哈希。

### Changed
- 全仓库行尾统一为 LF，并通过 `.gitattributes` 固化，杜绝跨平台 CRLF 差异。

[Unreleased]: https://github.com/furadx/iam-go/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/furadx/iam-go/releases/tag/v1.0.0
