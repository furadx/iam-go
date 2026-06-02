package apiserver

import (
	"github.com/gin-gonic/gin"

	"github.com/furadx/iam-go/internal/pkg/middleware"
)

// installMiddleware 安装全局中间件。
func installMiddleware(r *gin.Engine) {
	// Recovery 中间件 - 捕获 panic
	r.Use(middleware.Recovery())

	// CORS 中间件 - 跨域支持
	r.Use(middleware.CORS())

	// 可以在这里添加更多中间件：
	// r.Use(middleware.RequestID())
	// r.Use(middleware.Logger())
	// r.Use(middleware.RateLimit())
}
