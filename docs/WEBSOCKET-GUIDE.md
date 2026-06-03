# WebSocket 功能使用指南

## 📡 功能概述

iam-go 框架现在支持完整的 WebSocket 功能，用于实时双向通信。

### 核心特性

- ✅ **单写者模式** - 避免并发写冲突
- ✅ **自动心跳检测** - 30 秒 ping/pong 保持连接
- ✅ **优雅关闭** - 正确处理连接清理
- ✅ **用户连接管理** - Hub 模式管理多连接
- ✅ **广播消息** - 支持单用户和全员广播
- ✅ **Panic 恢复** - 读写协程自动恢复

---

## 🏗️ 架构设计

### 组件结构

```
pkg/ws/
├── connection.go    # WebSocket 连接封装
└── hub.go          # 连接管理中心

internal/apiserver/controller/v1/ws/
└── ws.go           # WebSocket 控制器
```

### 核心组件

#### 1. Connection（连接封装）
- **单写者模式**：writeLoop 是唯一写者，避免并发写
- **双向通道**：inChan 接收消息，outChan 发送消息
- **自动心跳**：定期发送 ping 帧保持连接
- **优雅关闭**：通过 closeChan 信号通知所有协程

#### 2. Hub（连接管理器）
- **用户连接映射**：`map[int64][]*Connection`
- **线程安全**：读写锁保护并发访问
- **广播功能**：向指定用户或所有用户发送消息
- **连接统计**：查询在线用户和连接数

---

## 🚀 快速开始

### 1. 连接 WebSocket

**客户端连接**：
```javascript
// 需要先获取 JWT token 并设置 user_id 到上下文
const ws = new WebSocket('ws://localhost:8080/api/v1/ws');

ws.onopen = () => {
    console.log('WebSocket 连接已建立');
};

ws.onmessage = (event) => {
    const data = JSON.parse(event.data);
    console.log('收到消息:', data);
};

ws.onerror = (error) => {
    console.error('WebSocket 错误:', error);
};

ws.onclose = () => {
    console.log('WebSocket 连接已关闭');
};
```

### 2. 发送消息到客户端

**从服务器端发送消息**：

```go
package main

import (
    "github.com/furadx/iam-go/internal/apiserver"
)

// 向指定用户发送消息
func notifyUser(userID int64, message interface{}) {
    hub := apiserver.GetHub()
    hub.BroadcastToUser(userID, message)
}

// 向所有在线用户发送消息
func notifyAllUsers(message interface{}) {
    hub := apiserver.GetHub()
    hub.BroadcastToAll(message)
}

// 示例：发送通知
type Notification struct {
    Type    string `json:"type"`
    Title   string `json:"title"`
    Message string `json:"message"`
    Time    int64  `json:"time"`
}

func sendNotification(userID int64) {
    notifyUser(userID, Notification{
        Type:    "notification",
        Title:   "新消息",
        Message: "您有一条新消息",
        Time:    time.Now().Unix(),
    })
}
```

### 3. 获取连接信息

```go
hub := apiserver.GetHub()

// 获取所有在线用户
onlineUsers := hub.GetOnlineUsers()

// 获取指定用户的连接数
count := hub.GetConnectionCount(userID)

// 获取总连接数
totalCount := hub.GetTotalConnectionCount()
```

---

## 📝 使用示例

### 示例 1：实时通知系统

```go
package service

import (
    "github.com/furadx/iam-go/internal/apiserver"
)

// NotificationService 通知服务
type NotificationService struct {}

// SendUserNotification 发送用户通知
func (s *NotificationService) SendUserNotification(userID int64, title, message string) {
    hub := apiserver.GetHub()
    
    notification := map[string]interface{}{
        "type":    "notification",
        "title":   title,
        "message": message,
        "time":    time.Now().Unix(),
    }
    
    hub.BroadcastToUser(userID, notification)
}

// SendSystemBroadcast 发送系统广播
func (s *NotificationService) SendSystemBroadcast(message string) {
    hub := apiserver.GetHub()
    
    broadcast := map[string]interface{}{
        "type":    "system_broadcast",
        "message": message,
        "time":    time.Now().Unix(),
    }
    
    hub.BroadcastToAll(broadcast)
}
```

### 示例 2：任务状态推送

```go
// TaskStatusUpdate 任务状态更新
type TaskStatusUpdate struct {
    Type     string `json:"type"`
    TaskID   int64  `json:"task_id"`
    Status   string `json:"status"`  // pending, processing, completed, failed
    Progress int    `json:"progress"` // 0-100
    Message  string `json:"message"`
}

func notifyTaskStatus(userID int64, taskID int64, status string, progress int) {
    hub := apiserver.GetHub()
    
    update := TaskStatusUpdate{
        Type:     "task_status_update",
        TaskID:   taskID,
        Status:   status,
        Progress: progress,
        Message:  fmt.Sprintf("任务 %d 状态: %s", taskID, status),
    }
    
    hub.BroadcastToUser(userID, update)
}
```

### 示例 3：聊天室功能

```go
// ChatMessage 聊天消息
type ChatMessage struct {
    Type      string `json:"type"`
    RoomID    int64  `json:"room_id"`
    FromUser  int64  `json:"from_user"`
    FromName  string `json:"from_name"`
    Message   string `json:"message"`
    Timestamp int64  `json:"timestamp"`
}

func sendChatMessage(roomID int64, fromUser int64, fromName string, message string, members []int64) {
    hub := apiserver.GetHub()
    
    chatMsg := ChatMessage{
        Type:      "chat_message",
        RoomID:    roomID,
        FromUser:  fromUser,
        FromName:  fromName,
        Message:   message,
        Timestamp: time.Now().Unix(),
    }
    
    // 向房间内所有成员发送消息
    for _, userID := range members {
        hub.BroadcastToUser(userID, chatMsg)
    }
}
```

---

## 🔒 认证要求

WebSocket 连接需要用户认证。需要在连接前通过认证中间件设置 `user_id` 到上下文。

### 添加认证中间件（示例）

```go
// 简单的认证中间件示例
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        
        // 验证 token 并获取 user_id
        userID, err := validateToken(token)
        if err != nil {
            c.JSON(401, gin.H{"error": "Unauthorized"})
            c.Abort()
            return
        }
        
        // 设置 user_id 到上下文
        c.Set("user_id", userID)
        c.Next()
    }
}

// 在路由中使用
v1.GET("/ws", AuthMiddleware(), wsController.HandleWebSocket)
```

---

## ⚙️ 配置参数

### 连接参数

```go
const (
    writeWait = 10 * time.Second      // 写超时
    pingPeriod = 30 * time.Second     // Ping 间隔
    pongWait = 60 * time.Second       // Pong 等待时间
    maxMessageSize = 512 * 1024       // 最大消息大小（512KB）
)
```

### Upgrader 配置

```go
var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        // 生产环境应该检查 Origin
        return true
    },
    HandshakeTimeout: 10 * time.Second,
}
```

---

## 🛡️ 最佳实践

### 1. 消息格式规范

建议使用统一的消息格式：

```json
{
  "type": "message_type",
  "data": { },
  "timestamp": 1234567890
}
```

### 2. 错误处理

```go
if err := hub.BroadcastToUser(userID, message); err != nil {
    log.Error("发送消息失败", zap.Error(err))
    // 不要 panic，记录日志即可
}
```

### 3. 连接清理

框架已自动处理连接清理，但在应用关闭时应该：

```go
// 在 main.go 中已实现
apiserver.GetHub().CloseAll()
```

### 4. 慢客户端处理

如果客户端读取慢，`outChan` 会被阻塞。当前实现：
- `outChan` 缓冲区 256 条消息
- 超过缓冲区会阻塞发送
- 心跳超时会自动断开慢客户端

---

## 🔍 调试技巧

### 查看连接状态

```go
hub := apiserver.GetHub()

// 获取在线用户列表
users := hub.GetOnlineUsers()
log.Info("在线用户", zap.Int64s("user_ids", users))

// 获取指定用户连接数
count := hub.GetConnectionCount(userID)
log.Info("用户连接数", zap.Int64("user_id", userID), zap.Int("count", count))
```

### 日志输出

所有 WebSocket 操作都会记录结构化日志：

```
INFO WebSocket 连接注册 user_id=123 connection_count=1
INFO 向用户广播消息 user_id=123 message={"type":"notification",...}
WARN WebSocket 发送消息失败 user_id=123 error=connection is closed
INFO WebSocket 连接注销 user_id=123 remaining_connections=0
```

---

## 📊 性能特点

- **单写者模式**：避免锁竞争，提高并发性能
- **缓冲通道**：256 条消息缓冲，减少阻塞
- **锁分离**：广播时不持锁，慢客户端不影响其他操作
- **自动清理**：心跳超时自动断开无响应连接

---

## 🚨 注意事项

1. **跨域配置**：生产环境需要配置 `CheckOrigin` 检查
2. **认证必需**：WebSocket 连接必须通过认证
3. **消息大小**：建议单条消息不超过 512KB
4. **连接数限制**：根据服务器资源合理限制并发连接数
5. **心跳机制**：客户端需要响应 Ping 帧，否则会被断开

---

## 📖 相关资源

- [gorilla/websocket 文档](https://github.com/gorilla/websocket)
- [WebSocket RFC 6455](https://tools.ietf.org/html/rfc6455)
- [ShiYu 项目原始实现](/Users/furad/GolandProjects/ShiYu/server/internal/pkg/ws/)

---

**版本**: v0.3.0  
**日期**: 2026-06-02
