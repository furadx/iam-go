package middleware

import (
	"net/http"
	"strings"

	"github.com/furadx/iam-go/pkg/code"
	"github.com/furadx/iam-go/pkg/jwt"
	"github.com/furadx/iam-go/pkg/response"
	"github.com/gin-gonic/gin"
)

const ContextUserIDKey = "user_id"

// Auth JWT 认证中间件
// 从 Authorization header 或 query 参数（用于 WebSocket）中提取 token 并验证
func Auth(jwtManager *jwt.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		var token string

		// 优先从 Authorization header 获取
		authorization := c.GetHeader("Authorization")
		if authorization != "" {
			parts := strings.SplitN(authorization, " ", 2)
			if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") && strings.TrimSpace(parts[1]) != "" {
				token = strings.TrimSpace(parts[1])
			}
		}

		// 如果 header 中没有，尝试从 query 参数获取（用于 WebSocket）
		if token == "" {
			token = c.Query("token")
		}

		if token == "" {
			response.Fail(c, http.StatusUnauthorized, code.ErrUnauthorized)
			c.Abort()
			return
		}

		claims, err := jwtManager.Parse(token)
		if err != nil {
			response.FailError(c, http.StatusUnauthorized, err)
			c.Abort()
			return
		}

		c.Set(ContextUserIDKey, claims.UserID)
		c.Next()
	}
}

// GetUserID 从 gin.Context 中获取当前用户 ID
func GetUserID(c *gin.Context) (int64, bool) {
	value, exists := c.Get(ContextUserIDKey)
	if !exists {
		return 0, false
	}

	uid, ok := value.(int64)
	if !ok || uid <= 0 {
		return 0, false
	}

	return uid, true
}
