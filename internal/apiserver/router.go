package apiserver

import (
	"github.com/gin-gonic/gin"

	authzctrl "github.com/furadx/iam-go/internal/apiserver/controller/v1/authz"
	"github.com/furadx/iam-go/internal/apiserver/controller/v1/user"
	wsctrl "github.com/furadx/iam-go/internal/apiserver/controller/v1/ws"
	srvv1 "github.com/furadx/iam-go/internal/apiserver/service/v1"
	"github.com/furadx/iam-go/internal/apiserver/store"
	authzpkg "github.com/furadx/iam-go/internal/pkg/authz"
	"github.com/furadx/iam-go/internal/pkg/middleware"
	"github.com/furadx/iam-go/internal/pkg/password"
	"github.com/furadx/iam-go/internal/pkg/revoke"
	"github.com/furadx/iam-go/pkg/token"
	"github.com/furadx/iam-go/pkg/ws"
)

var (
	// Hub 全局 WebSocket Hub
	Hub *ws.Hub
)

func init() {
	Hub = ws.NewHub()
}

// RouterDeps 路由装配所需依赖。
type RouterDeps struct {
	Store    store.Factory
	Token    *token.Manager
	Revoker  revoke.Revoker
	Authz    *authzpkg.Manager
	Security SecurityDeps
}

// SecurityDeps 路由安全能力依赖。
type SecurityDeps struct {
	LoginGuard     srvv1.LoginGuard
	PasswordPolicy password.Policy
	RateLimiter    middleware.RateLimiter
	CORS           middleware.CORSConfig
	APIRateLimit   middleware.RateLimitConfig
	LoginRateLimit middleware.RateLimitConfig
	RevokeFailOpen bool
}

// InitRouter 初始化路由。
func InitRouter(deps RouterDeps) *gin.Engine {
	r := gin.New()
	installMiddleware(r, deps.Security.CORS)
	r.GET("/healthz", healthCheck)
	installRoutes(r, deps)
	return r
}

func healthCheck(c *gin.Context) {
	c.JSON(200, gin.H{"status": "ok"})
}

func installRoutes(r *gin.Engine, deps RouterDeps) {
	userController := user.NewUserController(user.Deps{
		Store:          deps.Store,
		Token:          deps.Token,
		Revoker:        deps.Revoker,
		Authz:          deps.Authz,
		LoginGuard:     deps.Security.LoginGuard,
		PasswordPolicy: deps.Security.PasswordPolicy,
	})
	wsController := wsctrl.NewController(Hub)
	authzController := authzctrl.NewController(deps.Authz)

	auth := middleware.Auth(deps.Token, deps.Revoker, deps.Security.RevokeFailOpen)
	authz := middleware.Authz(deps.Authz)

	v1 := r.Group("/api/v1")
	v1.Use(middleware.RateLimit(deps.Security.RateLimiter, deps.Security.APIRateLimit))
	{
		// 公开接口
		v1.POST("/login", middleware.RateLimit(deps.Security.RateLimiter, deps.Security.LoginRateLimit), userController.Login)
		v1.POST("/refresh", userController.Refresh)
		v1.POST("/users", userController.Create) // 注册（公开）

		// 需登录（仅认证）
		authed := v1.Group("")
		authed.Use(auth)
		{
			authed.POST("/logout", userController.Logout)
			authed.GET("/ws", wsController.HandleWebSocket)

			// 需授权（Casbin；默认仅 admin 有策略）
			protected := authed.Group("")
			protected.Use(authz)
			{
				protected.GET("/users", userController.List)
				protected.GET("/users/:name", userController.Get)
				protected.DELETE("/users/:name", userController.Delete)
				protected.POST("/users/:name/roles", authzController.AssignRole)
				protected.DELETE("/users/:name/roles/:role", authzController.RevokeRole)
				protected.GET("/users/:name/roles", authzController.ListRoles)

				protected.POST("/authz/policies", authzController.AddPolicy)
				protected.DELETE("/authz/policies", authzController.RemovePolicy)
				protected.GET("/authz/policies", authzController.ListPolicies)
			}
		}
	}
}

// GetHub 返回全局 WebSocket Hub。
func GetHub() *ws.Hub {
	return Hub
}
