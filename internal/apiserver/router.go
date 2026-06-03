package apiserver

import (
	"github.com/gin-gonic/gin"

	"github.com/furadx/iam-go/internal/apiserver/controller/v1/user"
	wsctrl "github.com/furadx/iam-go/internal/apiserver/controller/v1/ws"
	"github.com/furadx/iam-go/internal/apiserver/store"
	"github.com/furadx/iam-go/pkg/ws"
)

var (
	// Hub 全局 WebSocket Hub
	Hub *ws.Hub
)

func init() {
	// 初始化 WebSocket Hub
	Hub = ws.NewHub()
}

// InitRouter 初始化路由。
func InitRouter(store store.Factory) *gin.Engine {
	r := gin.New()

	// 全局中间件
	installMiddleware(r)

	// 健康检查
	r.GET("/healthz", healthCheck)

	// 注册路由
	installRoutes(r, store)

	return r
}

// healthCheck 健康检查处理器。
func healthCheck(c *gin.Context) {
	c.JSON(200, gin.H{"status": "ok"})
}

// installRoutes 注册 API 路由。
func installRoutes(r *gin.Engine, store store.Factory) {
	// 初始化控制器
	userController := user.NewUserController(store)
	wsController := wsctrl.NewController(Hub)

	// v1 API 路由组
	v1 := r.Group("/api/v1")
	{
		// 用户相关接口
		users := v1.Group("/users")
		{
			users.POST("", userController.Create)
			users.GET("", userController.List)
			users.GET("/:name", userController.Get)
		}

		// WebSocket 连接（需要认证）
		v1.GET("/ws", wsController.HandleWebSocket)
	}
}

// GetHub 返回全局 WebSocket Hub。
func GetHub() *ws.Hub {
	return Hub
}
