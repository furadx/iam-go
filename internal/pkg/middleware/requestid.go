package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	// XRequestIDKey 是 Request-ID 的 HTTP 头键名。
	XRequestIDKey = "X-Request-ID"
)

// RequestID 中间件为每个请求生成或传递 Request-ID。
// 优先使用客户端传递的 Request-ID，如果没有则生成新的。
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 优先使用客户端传递的 Request-ID
		rid := c.GetHeader(XRequestIDKey)

		if rid == "" {
			// 生成新的 UUID
			rid = uuid.New().String()
		}

		// 设置到请求头（供后续处理使用）
		c.Request.Header.Set(XRequestIDKey, rid)

		// 设置到上下文（供日志等使用）
		c.Set(XRequestIDKey, rid)

		// 设置响应头
		c.Writer.Header().Set(XRequestIDKey, rid)

		c.Next()
	}
}
