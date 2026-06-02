package ws

import (
	"errors"
	"log"
	"runtime/debug"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// writeWait 单次写帧的超时
	writeWait = 10 * time.Second
	// defaultPingPeriod 默认 ping 间隔，需小于客户端读超时（建议 60s）
	defaultPingPeriod = 30 * time.Second
)

type Connection struct {
	wsConnect *websocket.Conn
	inChan    chan []byte // 连接读取到的消息会发送到 inChan
	outChan   chan []byte // 需要发送到连接的消息会发送到 outChan
	closeChan chan struct{}

	pingPeriod time.Duration // 每连接的 ping 间隔，构造时注入（便于测试）

	mutex    sync.Mutex
	isClosed bool
}

func New(conn *websocket.Conn, pingPeriod time.Duration) *Connection {
	if pingPeriod <= 0 {
		pingPeriod = defaultPingPeriod
	}
	wsConn := &Connection{
		wsConnect:  conn,
		inChan:     make(chan []byte, 256),
		outChan:    make(chan []byte, 256),
		closeChan:  make(chan struct{}),
		pingPeriod: pingPeriod,
	}
	// 启动读协程
	go wsConn.readLoop()
	// 启动写协程
	go wsConn.writeLoop()

	return wsConn
}

// ReadMessage 从 inChan 读取消息，阻塞直到有消息或连接关闭
func (conn *Connection) ReadMessage() (data []byte, err error) {
	select {
	case data = <-conn.inChan:
	case <-conn.closeChan:
		err = errors.New("connection is closed")
	}
	return
}

// WriteMessage 向 outChan 写入应用层文本消息，阻塞直到有空闲位置或连接关闭。
// 实际写帧由唯一写者 writeLoop 完成，保证每条连接只有一个 goroutine 写。
func (conn *Connection) WriteMessage(data []byte) (err error) {
	select {
	case conn.outChan <- data:
	case <-conn.closeChan:
		err = errors.New("connection is closed")
	}
	return
}

// readLoop 从 WebSocket 连接读取消息并发送到 inChan，直到连接关闭
func (conn *Connection) readLoop() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("ws readLoop panic recovered: %v\n%s", r, debug.Stack())
		}
		conn.Close()
	}()

	var (
		data []byte
		err  error
	)
	for {
		if _, data, err = conn.wsConnect.ReadMessage(); err != nil {
			return
		}
		// 阻塞在这里，等待 inChan 有空闲位置
		select {
		case conn.inChan <- data:
		case <-conn.closeChan: // closeChan 感知 conn 断开
			return
		}
	}
}

// writeLoop 是连接的唯一写者：发送 outChan 的应用消息，并定期发送 ping 控制帧。
// 所有对 wsConnect 的写操作只在此函数内发生，避免 gorilla/websocket 的并发写。
func (conn *Connection) writeLoop() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("ws writeLoop panic recovered: %v\n%s", r, debug.Stack())
		}
		conn.Close()
	}()

	ticker := time.NewTicker(conn.pingPeriod)
	defer ticker.Stop()

	for {
		select {
		case data := <-conn.outChan:
			_ = conn.wsConnect.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.wsConnect.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}
		case <-ticker.C:
			_ = conn.wsConnect.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.wsConnect.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case <-conn.closeChan:
			return
		}
	}
}

// Close 关闭连接，安全地关闭所有相关资源
func (conn *Connection) Close() {
	// 线程安全，可多次调用
	conn.wsConnect.Close()
	// 利用标记，让 closeChan 只关闭一次
	conn.mutex.Lock()
	defer conn.mutex.Unlock()
	if !conn.isClosed {
		close(conn.closeChan)
		conn.isClosed = true
	}
}
