package apiserver

import (
	"github.com/gin-gonic/gin"

	"github.com/furadx/iam-go/internal/pkg/middleware"
	"github.com/furadx/iam-go/pkg/log"
)

// installMiddleware 安装全局中间件。
func installMiddleware(r *gin.Engine) {
	// RequestID 中间件 - 请求追踪
	r.Use(middleware.RequestID())

	// Logger 中间件 - 日志记录
	r.Use(middleware.Logger(log.Default()))

	// Recovery 中间件 - 捕获 panic
	r.Use(middleware.Recovery())

	// CORS 中间件 - 跨域支持
	r.Use(middleware.CORS())

	// 可以在这里添加更多中间件：
	// r.Use(middleware.RateLimit())
	// r.Use(middleware.Auth())
}
