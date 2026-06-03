# 版本策略

本项目采用 [语义化版本 SemVer 2.0.0](https://semver.org/lang/zh-CN/)，变更记录采用 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/)。

## 版本号含义

版本号格式为 `MAJOR.MINOR.PATCH`（如 `v1.2.3`）：

| 位 | 何时递增 | 示例 |
|----|----------|------|
| **MAJOR** | 不向后兼容的变更（API 破坏、配置不兼容、行为破坏） | 删除/改名接口、错误码语义变化 |
| **MINOR** | 向后兼容的新功能 | 新增接口、新增配置项（带默认值） |
| **PATCH** | 向后兼容的缺陷修复 | 修 bug、安全补丁、文档与内部重构 |

约定：

- `v1.0.0` 起对外承诺兼容性。`0.x` 阶段视为不稳定，MINOR 也可能破坏兼容。
- 安全修复优先以 PATCH 发布；若修复必须破坏兼容，记入 MAJOR 并在 CHANGELOG 明确说明。
- 预发布版本用 `-rc.1`、`-beta.1` 等后缀（如 `v1.1.0-rc.1`）。

## 发布流程

每次发布按以下步骤操作：

1. **整理变更**：把 `CHANGELOG.md` 的 `[Unreleased]` 内容归入新版本小节，标注日期，按 `Added / Changed / Deprecated / Removed / Fixed / Security` 分类。
2. **定版本号**：依据上表确定 MAJOR / MINOR / PATCH。
3. **同步版本号**：更新 `cmd/apiserver/main.go` 中的 `Version` 常量为 `vX.Y.Z`。
4. **校验**：本地运行 `make ci`，确保门禁通过。
5. **提交**：提交信息形如 `release: vX.Y.Z`。
6. **打标签**：`git tag -a vX.Y.Z -m "vX.Y.Z"` 并推送 `git push --tags`。

构建时，`make build` 会通过 `-ldflags` 把 `Version` / `BuildDate` / `GitCommit` 注入二进制；`Version` 默认取 `git describe --tags`，因此**打了 tag 后构建即自动带上版本信息**。

## 提交信息建议（可选）

不强制 Conventional Commits，但推荐用类型前缀，便于回顾与未来自动化：

```text
feat:     新功能
fix:      缺陷修复
security: 安全修复
docs:     文档
refactor: 重构（不改行为）
test:     测试
chore:    构建 / CI / 杂项
release:  版本发布
```

## 当前版本

见 [CHANGELOG.md](../CHANGELOG.md) 与 `cmd/apiserver/main.go` 的 `Version`。当前为 **v1.0.0**。
