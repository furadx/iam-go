package main

import (
	"log"
	"net/http"

	"github.com/furadx/iam-go/middleware"
	"github.com/furadx/iam-go/pkg/jwt"
	"github.com/furadx/iam-go/pkg/response"
	"github.com/furadx/iam-go/pkg/ws"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 生产环境需要验证 origin
	},
}

func main() {
	r := gin.New()
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS())

	// JWT 管理器
	jwtManager := jwt.NewManager("your-secret-key", "ws-example", 3600)

	// WebSocket Hub
	hub := ws.NewHub()

	// 登录接口
	r.POST("/login", func(c *gin.Context) {
		userID := int64(12345)
		token, ttl, err := jwtManager.Sign(userID)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, 100003)
			return
		}
		response.OK(c, gin.H{"token": token, "expires_in": ttl})
	})

	// WebSocket 连接（需要认证）
	r.GET("/ws", middleware.Auth(jwtManager), func(c *gin.Context) {
		userID, _ := middleware.GetUserID(c)

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("升级 WebSocket 失败: %v", err)
			return
		}

		wsConn := hub.Register(userID, conn)
		defer hub.Unregister(userID, wsConn)

		// 发送欢迎消息
		hub.BroadcastToUser(userID, map[string]interface{}{
			"type":    "welcome",
			"message": "连接成功",
		})

		// 保持连接直到断开
		for {
			if _, err := wsConn.ReadMessage(); err != nil {
				break
			}
		}
	})

	// 广播测试接口
	r.POST("/broadcast/:userID", middleware.Auth(jwtManager), func(c *gin.Context) {
		var uriParams struct {
			UserID int64 `uri:"userID" binding:"required"`
		}
		if err := c.ShouldBindUri(&uriParams); err != nil {
			response.Fail(c, http.StatusBadRequest, 100001)
			return
		}
		userID := uriParams.UserID

		var req struct {
			Message string `json:"message"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Fail(c, http.StatusBadRequest, 100001)
			return
		}

		hub.BroadcastToUser(userID, map[string]interface{}{
			"type":    "notification",
			"message": req.Message,
		})

		response.OK(c, gin.H{"status": "sent"})
	})

	log.Println("WebSocket 服务启动在 :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
