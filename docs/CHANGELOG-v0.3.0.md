# iam-go v0.3.0 功能增强说明

## 📦 本次更新内容

### ✅ 已完成的功能

#### 1. RequestID 中间件
**文件**: `internal/pkg/middleware/requestid.go`

**功能**:
- 为每个请求生成或传递唯一的 Request-ID
- 支持客户端传递的 Request-ID
- 自动添加到响应头和上下文
- 便于请求追踪和日志关联

**使用**:
```go
r.Use(middleware.RequestID())
```

#### 2. 结构化日志系统（zap）
**文件**: `pkg/log/log.go`

**功能**:
- 升级到 uber/zap 高性能日志库
- 支持 JSON 和 Console 两种格式
- 支持多种日志级别（debug/info/warn/error/fatal）
- 支持彩色输出（Console 模式）
- 支持多输出路径
- 性能优于标准库 log

**使用**:
```go
// 使用默认 logger
log.Info("message", zap.String("key", "value"))
log.Infof("formatted %s", "message")

// 自定义 logger
opts := log.NewOptions()
opts.Level = "debug"
opts.Format = "json"
logger := log.New(opts)
```

#### 3. 日志中间件
**文件**: `internal/pkg/middleware/logger.go`

**功能**:
- 记录每个 HTTP 请求的详细信息
- 包含 Request-ID、方法、路径、状态码、延迟、IP、User-Agent
- 结构化日志输出
- 便于日志分析和问题追踪

#### 4. Options 模式配置管理
**文件**: `internal/apiserver/options/`

**包含**:
- `options.go` - 主配置组合
- `server.go` - 服务器选项
- `database.go` - 数据库选项
- `log.go` - 日志选项

**功能**:
- 层次化配置命名（`--server.mode`, `--db.host`, `--log.level`）
- 支持 pflag 命令行参数
- 配置验证（Validate 方法）
- Complete 模式（填充默认值）
- 更好的可扩展性

**示例**:
```go
opts := options.NewOptions()

// 添加命令行标志
fs := pflag.NewFlagSet("apiserver", pflag.ExitOnError)
opts.AddFlags(fs)
fs.Parse(os.Args[1:])

// 验证配置
if errs := opts.Validate(); len(errs) > 0 {
    // 处理错误
}

// 完成配置
completed := opts.Complete()
```

#### 5. Makefile 自动化
**文件**: `Makefile`

**包含的目标**:
```bash
make build          # 构建二进制文件
make run            # 运行应用
make test           # 运行测试
make test.cover     # 测试覆盖率
make lint           # 代码检查
make fmt            # 格式化代码
make clean          # 清理构建产物
make install        # 安装依赖
make install.tools  # 安装开发工具
make tidy           # 整理依赖
make vet            # 运行 go vet
make docker-build   # 构建 Docker 镜像

# 跨平台编译
make build.linux    # Linux 版本
make build.darwin   # macOS 版本
make build.windows  # Windows 版本
make build.all      # 所有平台
```

**版本信息注入**:
```bash
# 编译时自动注入版本信息
VERSION=$(git describe --tags --always --dirty)
BUILD_DATE=$(date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(git rev-parse --short HEAD)
```

#### 6. 响应格式增强
**文件**: `internal/pkg/util/response.go`

**更新**:
- 添加 `request_id` 字段
- 统一响应格式
- 更好的错误追踪

**响应示例**:
```json
{
  "code": 0,
  "message": "OK",
  "data": {...},
  "request_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

#### 7. 中间件集成更新
**文件**: `internal/apiserver/middleware.go`

**中间件顺序**:
1. RequestID - 生成请求 ID
2. Logger - 记录请求日志
3. Recovery - 捕获 panic
4. CORS - 跨域支持

### 📦 新增依赖

```
github.com/google/uuid v1.6.0       # UUID 生成
go.uber.org/zap v1.28.0             # 结构化日志
go.uber.org/multierr v1.10.0        # 多错误处理
github.com/spf13/pflag v1.0.5       # 命令行参数
```

### 📁 项目结构更新

```
iam-go/
├── api/                    # API 定义（新增）
├── build/                  # CI/CD 配置（新增）
├── cmd/
│   └── apiserver/
├── configs/
├── deployments/            # 部署配置（新增）
├── docs/                   # 文档（新增）
├── init/                   # 系统初始化（新增）
├── internal/
│   ├── apiserver/
│   │   ├── config/
│   │   ├── controller/
│   │   ├── middleware.go
│   │   ├── model/
│   │   ├── options/        # Options 模式（新增）
│   │   │   ├── options.go
│   │   │   ├── server.go
│   │   │   ├── database.go
│   │   │   └── log.go
│   │   ├── router.go
│   │   ├── server.go
│   │   ├── service/
│   │   └── store/
│   └── pkg/
│       ├── code/
│       ├── middleware/
│       │   ├── cors.go
│       │   ├── logger.go      # 新增
│       │   ├── recovery.go
│       │   └── requestid.go   # 新增
│       └── util/
├── pkg/
│   ├── auth/
│   └── log/               # 升级到 zap
├── scripts/
├── test/                  # 测试数据（新增）
├── tools/                 # 项目工具（新增）
├── Makefile              # 新增
├── go.mod
└── go.sum
```

### 🚀 使用示例

#### 启动应用

```bash
# 使用默认配置
make run

# 使用配置文件
./bin/apiserver -config configs/config.yaml

# 使用命令行参数
./bin/apiserver --server.mode=debug --server.addr=:8080 --db.host=localhost --log.level=debug
```

#### 查看日志

**Console 格式（彩色输出）**:
```
2026-06-02T10:00:57.123Z    INFO    HTTP Request    {"request_id": "550e8400-e29b-41d4-a716-446655440000", "method": "POST", "path": "/v1/users", "status": 200, "latency": "15ms"}
```

**JSON 格式**:
```json
{
  "time": "2026-06-02T10:00:57.123Z",
  "level": "INFO",
  "msg": "HTTP Request",
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "method": "POST",
  "path": "/v1/users",
  "status": 200,
  "latency": 15000000
}
```

### 📋 后续可选功能

以下功能已准备好框架，可按需添加：

1. **认证中间件** - JWT/Basic Auth
2. **限流中间件** - Rate Limiter
3. **单元测试** - 测试框架和用例
4. **API 文档** - Swagger 集成
5. **Docker 支持** - Dockerfile 和 docker-compose
6. **CI/CD** - GitHub Actions
7. **健康检查增强** - 数据库连接检查
8. **性能监控** - Prometheus metrics

### 🔄 升级建议

如果从 v0.2.0 升级：

1. 安装新依赖：
   ```bash
   go mod tidy
   ```

2. 更新配置文件（可选）：
   - 使用新的层次化配置
   - 配置日志选项

3. 编译测试：
   ```bash
   make build
   ```

4. 运行测试（如果有）：
   ```bash
   make test
   ```

### ⚙️ 配置示例

**使用环境变量**:
```bash
export IAM_SERVER_MODE=release
export IAM_SERVER_ADDR=:8080
export IAM_DB_HOST=localhost
export IAM_DB_USER=postgres
export IAM_DB_PASSWORD=your_password
export IAM_LOG_LEVEL=info
export IAM_LOG_FORMAT=json
```

**使用命令行参数**:
```bash
./bin/apiserver \
  --server.mode=release \
  --server.addr=:8080 \
  --db.host=localhost \
  --db.user=postgres \
  --db.password=your_password \
  --log.level=info \
  --log.format=json
```

### 🎯 性能改进

- **日志性能**: zap 比标准库快 4-10 倍
- **请求追踪**: Request-ID 零性能损耗
- **中间件优化**: 最小化内存分配

### 📊 版本对比

| 功能 | v0.2.0 | v0.3.0 |
|------|--------|--------|
| 请求追踪 | ❌ | ✅ RequestID |
| 结构化日志 | ❌ | ✅ zap |
| 日志中间件 | ❌ | ✅ |
| Options 模式 | ❌ | ✅ |
| 层次化配置 | ❌ | ✅ |
| Makefile | ❌ | ✅ |
| 响应 RequestID | ❌ | ✅ |
| 跨平台编译 | ❌ | ✅ |

---

**作者**: Claude Opus 4.8  
**日期**: 2026-06-02  
**版本**: v0.3.0
