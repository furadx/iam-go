# Makefile for iam-go project
#
# Usage:
#   make build              构建二进制文件
#   make run                运行应用
#   make test               运行测试
#   make lint               代码检查
#   make fmt                格式化代码
#   make clean              清理构建产物
#   make install            安装依赖
#   make docker-build       构建 Docker 镜像
#   make help               显示帮助信息

# 变量定义
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOLINT=golangci-lint

# 项目信息
BINARY_NAME=apiserver
BINARY_DIR=bin
CMD_DIR=cmd/apiserver
MAIN_FILE=$(CMD_DIR)/main.go

# 版本信息
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "v0.3.0")
BUILD_DATE=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# 构建标志
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildDate=$(BUILD_DATE) -X main.GitCommit=$(GIT_COMMIT)"

.PHONY: all build run test lint fmt fmt.check vet ci clean install docker-build help

## all: 默认目标 - 格式化、检查、测试、构建
all: fmt lint test build

## build: 构建二进制文件
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BINARY_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME) $(MAIN_FILE)
	@echo "Build complete: $(BINARY_DIR)/$(BINARY_NAME)"

## run: 运行应用
run:
	@echo "Running $(BINARY_NAME)..."
	@$(GOBUILD) -o $(BINARY_DIR)/$(BINARY_NAME) $(MAIN_FILE)
	@$(BINARY_DIR)/$(BINARY_NAME)

## test: 运行所有测试
test:
	@echo "Running tests..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	@echo "Tests complete"

## test.cover: 运行测试并显示覆盖率
test.cover: test
	@echo "Generating coverage report..."
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

## lint: 代码检查
lint:
	@echo "Running linter..."
	@which $(GOLINT) > /dev/null || (echo "golangci-lint not installed, run: make install.tools" && exit 1)
	$(GOLINT) run ./...
	@echo "Lint complete"

## fmt: 格式化代码
fmt:
	@echo "Formatting code..."
	$(GOFMT) -s -w .
	@echo "Format complete"

## fmt.check: 只检查格式不改写（CI 门禁用）
fmt.check:
	@echo "Checking format..."
	@unformatted=$$($(GOFMT) -s -l .); \
	if [ -n "$$unformatted" ]; then \
		echo "以下文件未通过 gofmt -s，请运行 'make fmt'："; \
		echo "$$unformatted"; \
		exit 1; \
	fi
	@echo "Format check passed"

## ci: 本地复现 CI 门禁（fmt 检查 / vet / build / test，与 .github/workflows/ci.yml 一致）
ci: fmt.check vet build
	@echo "Running tests..."
	$(GOTEST) ./...
	@echo "CI gate passed"

## clean: 清理构建产物
clean:
	@echo "Cleaning..."
	@rm -rf $(BINARY_DIR)
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

## install: 安装项目依赖
install:
	@echo "Installing dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "Dependencies installed"

## install.tools: 安装开发工具
install.tools:
	@echo "Installing development tools..."
	@which golangci-lint > /dev/null || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin
	@echo "Tools installed"

## docker-build: 构建 Docker 镜像
docker-build:
	@echo "Building Docker image..."
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		-t iam-go:$(VERSION) -f build/Dockerfile .
	@echo "Docker image built: iam-go:$(VERSION)"

## tidy: 整理依赖
tidy:
	@echo "Tidying dependencies..."
	$(GOMOD) tidy
	@echo "Tidy complete"

## vet: 运行 go vet
vet:
	@echo "Running go vet..."
	$(GOCMD) vet ./...
	@echo "Vet complete"

## build.linux: 交叉编译 Linux 版本
build.linux:
	@echo "Building for Linux..."
	@mkdir -p $(BINARY_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_FILE)
	@echo "Linux build complete: $(BINARY_DIR)/$(BINARY_NAME)-linux-amd64"

## build.darwin: 交叉编译 macOS 版本
build.darwin:
	@echo "Building for macOS..."
	@mkdir -p $(BINARY_DIR)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_FILE)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_FILE)
	@echo "macOS build complete"

## build.windows: 交叉编译 Windows 版本
build.windows:
	@echo "Building for Windows..."
	@mkdir -p $(BINARY_DIR)
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_FILE)
	@echo "Windows build complete: $(BINARY_DIR)/$(BINARY_NAME)-windows-amd64.exe"

## build.all: 构建所有平台
build.all: build.linux build.darwin build.windows

## help: 显示帮助信息
help:
	@echo "Available targets:"
	@awk '/^##/ { \
		sub(/^## /, "", $$0); \
		printf "  %-20s %s\n", $$1, substr($$0, index($$0, $$2)); \
	}' $(MAKEFILE_LIST)
