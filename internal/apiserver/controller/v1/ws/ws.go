package ws

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	"github.com/furadx/iam-go/internal/pkg/code"
	"github.com/furadx/iam-go/internal/pkg/util"
	"github.com/furadx/iam-go/pkg/log"
	pkgws "github.com/furadx/iam-go/pkg/ws"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// 生产环境应该检查 Origin
		return true
	},
	HandshakeTimeout: 10 * time.Second,
}

// Controller WebSocket 控制器。
type Controller struct {
	hub *pkgws.Hub
}

// NewController 创建新的 WebSocket 控制器。
func NewController(hub *pkgws.Hub) *Controller {
	return &Controller{hub: hub}
}

// HandleWebSocket 处理 WebSocket 连接。
// 需要用户认证，从上下文获取 user_id。
func (c *Controller) HandleWebSocket(ctx *gin.Context) {
	// 从上下文获取 user_id（需要认证中间件设置）
	userIDVal, exists := ctx.Get("user_id")
	if !exists {
		util.WriteResponse(ctx, code.New(code.ErrUnauthorized), nil)
		return
	}

	userID, ok := userIDVal.(int64)
	if !ok {
		util.WriteResponse(ctx, code.New(code.ErrUnauthorized), nil)
		return
	}

	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		log.Error("WebSocket 升级失败",
			zap.Int64("user_id", userID),
			zap.Error(err))
		return
	}

	// 设置读取超时
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	// 注册连接
	wsConn := c.hub.Register(userID, conn)

	// 启动读取循环（保持连接活跃，处理客户端断开）
	go c.readPump(userID, wsConn)
}

// readPump 读取客户端消息（主要用于检测连接断开）。
func (c *Controller) readPump(userID int64, conn *pkgws.Connection) {
	defer func() {
		c.hub.Unregister(userID, conn)
		conn.Close()
	}()

	for {
		_, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Warn("WebSocket 读取错误",
					zap.Int64("user_id", userID),
					zap.Error(err))
			}
			break
		}
		// 客户端发送的消息暂时忽略，只用于保持连接
	}
}
