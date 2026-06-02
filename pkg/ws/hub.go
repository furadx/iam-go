package ws

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Hub 管理所有 WebSocket 连接
type Hub struct {
	connections map[int64][]*Connection
	pingPeriod  time.Duration // 注入到每条新连接的 ping 间隔
	mu          sync.RWMutex
}

// NewHub 创建新的 Hub
func NewHub() *Hub {
	return &Hub{
		connections: make(map[int64][]*Connection),
		pingPeriod:  defaultPingPeriod,
	}
}

// SetPingPeriod 设置 ping 间隔（用于测试或自定义配置）
func (h *Hub) SetPingPeriod(period time.Duration) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.pingPeriod = period
}

// Register 注册用户连接
func (h *Hub) Register(userID int64, conn *websocket.Conn) *Connection {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 创建 Connection 包装 WebSocket 连接
	Conn := New(conn, h.pingPeriod)
	h.connections[userID] = append(h.connections[userID], Conn)
	log.Printf("WebSocket 连接注册 userID=%d, 当前连接数=%d", userID, len(h.connections[userID]))

	// 返回 Connection 以供后续使用（如注销）
	return Conn
}

// Unregister 注销用户连接
func (h *Hub) Unregister(userID int64, conn *Connection) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 创建 Connection 包装 WebSocket 连接
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

	log.Printf("WebSocket 连接注销 userID=%d, 剩余连接数=%d", userID, len(h.connections[userID]))
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
	log.Printf("WebSocket hub 已关闭所有连接")
}

// BroadcastToUser 向指定用户的所有连接广播消息
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
		log.Printf("序列化 WebSocket 消息失败: %v", err)
		return
	}

	log.Printf("向用户 %d 广播消息: %s", userID, string(data))

	// 发送时不持锁，慢连接不阻塞 Register/Unregister 与其它广播。
	for _, conn := range connsCopy {
		if err := conn.WriteMessage(data); err != nil {
			log.Printf("WebSocket 发送消息失败: %v", err)
			// 发送失败的连接会在下次心跳时被清理
		}
	}
}

// GetConnectionCount 获取指定用户的连接数（用于监控或调试）
func (h *Hub) GetConnectionCount(userID int64) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.connections[userID])
}

// GetTotalConnections 获取所有用户的总连接数
func (h *Hub) GetTotalConnections() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	total := 0
	for _, conns := range h.connections {
		total += len(conns)
	}
	return total
}
