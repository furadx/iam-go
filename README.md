# IAM-Go

一个轻量级的 Go Web 应用通用框架，提供认证、授权、统一响应、WebSocket 等基础能力。

## 特性

- 🔐 **JWT 认证** - 完整的 JWT token 生成、解析、验证
- 🌐 **HTTP 中间件** - CORS、认证、错误恢复等常用中间件
- 📡 **WebSocket 支持** - 开箱即用的 WebSocket Hub 管理
- 📦 **统一响应格式** - 标准化的 API 响应结构
- ⚠️ **错误码系统** - 可扩展的业务错误码体系
- 🚀 **基于 Gin** - 高性能的 HTTP 框架

## 安装

```bash
go get github.com/furadx/iam-go
```

## 快速开始

```go
package main

import (
    "github.com/furadx/iam-go/pkg/jwt"
    "github.com/furadx/iam-go/middleware"
    "github.com/gin-gonic/gin"
)

func main() {
    r := gin.New()
    
    // 使用中间件
    r.Use(middleware.Recovery())
    r.Use(middleware.CORS())
    
    // JWT 认证
    jwtManager := jwt.NewManager("your-secret", "your-app", 3600)
    r.GET("/protected", middleware.Auth(jwtManager), handler)
    
    r.Run(":8080")
}
```

## 模块说明

### pkg/code - 错误码系统
标准化的业务错误码定义和管理。

### pkg/response - 统一响应
HTTP API 的标准响应格式封装。

### pkg/jwt - JWT 认证
JWT token 的生成、解析和验证。

### pkg/ws - WebSocket
WebSocket 连接管理和消息广播。

### middleware - HTTP 中间件
常用的 Gin 中间件集合。

## 版本

当前版本：v0.1.0

## 许可证

MIT License

## 源项目

本框架从 [ShiYu](https://github.com/furadx/ShiYu) 项目中抽离而来。
