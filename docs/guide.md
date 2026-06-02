# IAM-Go 使用指南

## 目录
- [快速开始](#快速开始)
- [核心模块](#核心模块)
- [示例](#示例)
- [最佳实践](#最佳实践)

## 快速开始

### 安装

```bash
go get github.com/furadx/iam-go
```

### 基础示例

```go
package main

import (
    "github.com/furadx/iam-go/middleware"
    "github.com/furadx/iam-go/pkg/jwt"
    "github.com/furadx/iam-go/pkg/response"
    "github.com/gin-gonic/gin"
    "net/http"
)

func main() {
    r := gin.New()
    
    // 必备中间件
    r.Use(middleware.Recovery())  // Panic 恢复
    r.Use(middleware.CORS())      // 跨域支持
    
    // 创建 JWT 管理器
    jwtManager := jwt.NewManager("your-secret", "your-app", 3600)
    
    // 公开接口
    r.POST("/login", loginHandler(jwtManager))
    
    // 受保护接口
    r.GET("/protected", middleware.Auth(jwtManager), protectedHandler)
    
    r.Run(":8080")
}

func loginHandler(jm *jwt.Manager) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 验证用户凭证...
        userID := int64(12345)
        
        token, ttl, err := jm.Sign(userID)
        if err != nil {
            response.Fail(c, http.StatusInternalServerError, 100003)
            return
        }
        
        response.OK(c, gin.H{
            "token": token,
            "expires_in": ttl,
        })
    }
}

func protectedHandler(c *gin.Context) {
    userID, _ := middleware.GetUserID(c)
    response.OK(c, gin.H{"user_id": userID})
}
```

## 核心模块

### 1. pkg/code - 错误码系统

统一的业务错误码管理。

**预定义错误码：**
- `OK = 0` - 成功
- `ErrInvalidParam = 100001` - 参数错误
- `ErrUnauthorized = 100002` - 未认证
- `ErrInternal = 100003` - 服务器内部错误
- `ErrTooManyReq = 100004` - 请求过于频繁
- `ErrInvalidToken = 300001` - 无效 token
- `ErrExpiredToken = 300002` - token 已过期

**自定义错误码：**

```go
import "github.com/furadx/iam-go/pkg/code"

// 注册自定义错误码
func init() {
    code.RegisterMessage(400001, "自定义错误")
}

// 创建错误
err := code.New(400001)

// 包装错误
err := code.Wrap(400001, originalErr)

// 从错误获取错误码
bizCode := code.FromError(err)
```

### 2. pkg/response - 统一响应

标准化的 HTTP 响应格式。

**成功响应：**

```go
response.OK(c, gin.H{
    "id": 123,
    "name": "test",
})
```

响应格式：
```json
{
  "code": 0,
  "message": "OK",
  "data": {
    "id": 123,
    "name": "test"
  }
}
```

**错误响应：**

```go
// 使用错误码
response.Fail(c, http.StatusBadRequest, code.ErrInvalidParam)

// 使用自定义消息
response.FailWithMessage(c, http.StatusBadRequest, code.ErrInvalidParam, "用户名不能为空")

// 使用 error 对象
response.FailError(c, http.StatusInternalServerError, err)
```

响应格式：
```json
{
  "code": 100001,
  "message": "参数错误"
}
```

### 3. pkg/jwt - JWT 认证

完整的 JWT token 生成和验证。

**创建 JWT 管理器：**

```go
import "github.com/furadx/iam-go/pkg/jwt"

jwtManager := jwt.NewManager(
    "your-secret-key",  // 密钥
    "your-app-name",    // 签发者
    3600,               // 过期时间（秒）
)
```

**生成 Token：**

```go
token, ttl, err := jwtManager.Sign(userID)
if err != nil {
    // 处理错误
}
// token: JWT 字符串
// ttl: 过期时间（秒）
```

**解析 Token：**

```go
claims, err := jwtManager.Parse(tokenString)
if err != nil {
    // token 无效或已过期
}
userID := claims.UserID
```

### 4. middleware - HTTP 中间件

#### Auth - JWT 认证中间件

```go
// 应用到路由
r.GET("/protected", middleware.Auth(jwtManager), handler)

// 在处理函数中获取用户 ID
func handler(c *gin.Context) {
    userID, exists := middleware.GetUserID(c)
    if !exists {
        // 不应该发生，因为中间件已验证
    }
}
```

Token 可以通过两种方式传递：
1. **Authorization Header**: `Bearer <token>`
2. **Query 参数**: `?token=<token>` (用于 WebSocket)

#### CORS - 跨域中间件

```go
r.Use(middleware.CORS())
```

支持：
- 动态 Origin
- 凭证携带
- 预检请求
- 常用 HTTP 方法和头部

#### Recovery - Panic 恢复中间件

```go
r.Use(middleware.Recovery())
```

特性：
- 捕获 panic
- 记录完整堆栈
- 返回统一 500 错误
- 防止连接重置

### 5. pkg/ws - WebSocket

生产级别的 WebSocket 连接管理。

**创建 Hub：**

```go
import "github.com/furadx/iam-go/pkg/ws"

hub := ws.NewHub()

// 可选：自定义 ping 间隔
hub.SetPingPeriod(30 * time.Second)
```

**WebSocket 处理函数：**

```go
import "github.com/gorilla/websocket"

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool {
        return true // 生产环境需验证
    },
}

r.GET("/ws", middleware.Auth(jwtManager), func(c *gin.Context) {
    userID, _ := middleware.GetUserID(c)
    
    // 升级连接
    conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        return
    }
    
    // 注册连接
    wsConn := hub.Register(userID, conn)
    defer hub.Unregister(userID, wsConn)
    
    // 保持连接
    for {
        if _, err := wsConn.ReadMessage(); err != nil {
            break
        }
    }
})
```

**广播消息：**

```go
// 向指定用户的所有连接广播
hub.BroadcastToUser(userID, map[string]interface{}{
    "type": "notification",
    "message": "新消息",
    "data": someData,
})
```

**监控连接：**

```go
// 获取用户连接数
count := hub.GetConnectionCount(userID)

// 获取总连接数
total := hub.GetTotalConnections()
```

**优雅关闭：**

```go
// 关闭所有连接
hub.CloseAll()
```

## 示例

完整示例代码位于 `examples/` 目录：

- `examples/basic/` - 基础 HTTP API 示例
- `examples/websocket/` - WebSocket 集成示例

运行示例：

```bash
cd examples/basic
go run main.go

# 测试
curl -X POST http://localhost:8080/login
curl -H "Authorization: Bearer <token>" http://localhost:8080/protected
```

## 最佳实践

### 1. JWT 密钥管理

```go
// ❌ 不要硬编码
jwtManager := jwt.NewManager("hardcoded-secret", ...)

// ✅ 从环境变量或配置文件读取
secret := os.Getenv("JWT_SECRET")
jwtManager := jwt.NewManager(secret, ...)
```

### 2. 错误处理

```go
// ❌ 返回内部错误细节
response.FailWithMessage(c, 500, code.ErrInternal, err.Error())

// ✅ 使用标准错误码，内部错误记录日志
log.Printf("数据库错误: %v", err)
response.Fail(c, 500, code.ErrInternal)
```

### 3. WebSocket 连接管理

```go
// ✅ 始终注销连接
wsConn := hub.Register(userID, conn)
defer hub.Unregister(userID, wsConn)

// ✅ 优雅关闭时清理所有连接
defer hub.CloseAll()
```

### 4. CORS 配置

```go
// 开发环境：允许所有
r.Use(middleware.CORS())

// 生产环境：自定义 CheckOrigin
upgrader.CheckOrigin = func(r *http.Request) bool {
    origin := r.Header.Get("Origin")
    return origin == "https://yourdomain.com"
}
```

### 5. 中间件顺序

```go
// ✅ 推荐顺序
r.Use(middleware.Recovery())  // 1. Panic 恢复
r.Use(middleware.CORS())      // 2. CORS
// ... 其他全局中间件
r.GET("/api", middleware.Auth(jm), handler)  // 3. 认证（路由级）
```

## 扩展

### 自定义错误码

```go
// 在你的项目中
package myapp

import "github.com/furadx/iam-go/pkg/code"

const (
    ErrUserNotFound = 500001
    ErrDuplicateEmail = 500002
)

func init() {
    code.RegisterMessage(ErrUserNotFound, "用户不存在")
    code.RegisterMessage(ErrDuplicateEmail, "邮箱已被注册")
}
```

### 自定义响应格式

如果需要不同的响应格式，可以基于 `pkg/response` 创建自己的包装函数：

```go
func CustomOK(c *gin.Context, data interface{}) {
    c.JSON(200, gin.H{
        "success": true,
        "result": data,
        "timestamp": time.Now().Unix(),
    })
}
```

## 参考

- [Gin 框架文档](https://gin-gonic.com/)
- [JWT 规范](https://jwt.io/)
- [Gorilla WebSocket](https://github.com/gorilla/websocket)
- [源项目 ShiYu](https://github.com/furadx/ShiYu)
