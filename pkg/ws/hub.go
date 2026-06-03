package ws

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	"github.com/furadx/iam-go/pkg/log"
)

// Hub 管理所有 WebSocket 连接。
type Hub struct {
	connections map[int64][]*Connection
	pingPeriod  time.Duration // 注入到每条新连接的 ping 间隔
	mu          sync.RWMutex
}

// NewHub 创建新的 Hub。
func NewHub() *Hub {
	return &Hub{
		connections: make(map[int64][]*Connection),
		pingPeriod:  defaultPingPeriod,
	}
}

// Register 注册用户连接。
func (h *Hub) Register(userID int64, conn *websocket.Conn) *Connection {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 创建 Connection 包装 WebSocket 连接
	wsConn := New(conn, h.pingPeriod)
	h.connections[userID] = append(h.connections[userID], wsConn)
	log.Info("WebSocket 连接注册",
		zap.Int64("user_id", userID),
		zap.Int("connection_count", len(h.connections[userID])))

	// 返回 Connection 以供后续使用（如注销）
	return wsConn
}

// Unregister 注销用户连接。
func (h *Hub) Unregister(userID int64, conn *Connection) {
	h.mu.Lock()
	defer h.mu.Unlock()

	conns := h.connections[userID]
	for i, c := range conns {
		if c == conn {
			h.connections[userID] = append(conns[:i], conns[i+1:]...)
			break
		}
	}

	if len(h.connections[userID]) == 0 {
		delete(h.connections, userID)
	}

	log.Info("WebSocket 连接注销",
		zap.Int64("user_id", userID),
		zap.Int("remaining_connections", len(h.connections[userID])))
}

// CloseAll 关闭所有连接并清空 map，供优雅关闭时调用。
// 每个连接的 Close 是幂等的，readLoop/writeLoop 会感知 closeChan 后自行收尾。
func (h *Hub) CloseAll() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for userID, conns := range h.connections {
		for _, conn := range conns {
			conn.Close()
		}
		delete(h.connections, userID)
	}
	log.Info("WebSocket hub 已关闭所有连接")
}

// BroadcastToUser 向指定用户的所有连接广播消息。
func (h *Hub) BroadcastToUser(userID int64, message interface{}) {
	// 在锁内复制连接列表：conns 与底层数组被 Register/Unregister 的 append
	// 原地改写，复制必须在持锁期间完成，否则与那两者发生数据竞争。
	h.mu.RLock()
	conns := h.connections[userID]
	connsCopy := make([]*Connection, len(conns))
	copy(connsCopy, conns)
	h.mu.RUnlock()

	if len(connsCopy) == 0 {
		return
	}

	data, err := json.Marshal(message)
	if err != nil {
		log.Error("序列化 WebSocket 消息失败",
			zap.Int64("user_id", userID),
			zap.Error(err))
		return
	}

	log.Info("向用户广播消息",
		zap.Int64("user_id", userID),
		zap.String("message", string(data)))

	// 发送时不持锁，慢连接不阻塞 Register/Unregister 与其它广播。
	for _, conn := range connsCopy {
		if err := conn.WriteMessage(data); err != nil {
			log.Warn("WebSocket 发送消息失败",
				zap.Int64("user_id", userID),
				zap.Error(err))
			// 发送失败的连接会在下次心跳时被清理
		}
	}
}

// BroadcastToAll 向所有在线用户广播消息。
func (h *Hub) BroadcastToAll(message interface{}) {
	h.mu.RLock()
	userIDs := make([]int64, 0, len(h.connections))
	for userID := range h.connections {
		userIDs = append(userIDs, userID)
	}
	h.mu.RUnlock()

	for _, userID := range userIDs {
		h.BroadcastToUser(userID, message)
	}
}

// GetOnlineUsers 获取所有在线用户 ID 列表。
func (h *Hub) GetOnlineUsers() []int64 {
	h.mu.RLock()
	defer h.mu.RUnlock()

	users := make([]int64, 0, len(h.connections))
	for userID := range h.connections {
		users = append(users, userID)
	}
	return users
}

// GetConnectionCount 获取指定用户的连接数。
func (h *Hub) GetConnectionCount(userID int64) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return len(h.connections[userID])
}

// GetTotalConnectionCount 获取总连接数。
func (h *Hub) GetTotalConnectionCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	count := 0
	for _, conns := range h.connections {
		count += len(conns)
	}
	return count
}
