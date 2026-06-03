package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/furadx/iam-go/internal/pkg/code"
	"github.com/furadx/iam-go/pkg/token"
)

// ContextUserIDKey 是认证后写入上下文的用户 ID 键名。
const ContextUserIDKey = "user_id"

// Auth 是基于 JWT 的认证中间件。
// 从 Authorization: Bearer <token> 头解析并校验 Token，
// 成功后把用户 ID 写入上下文（ContextUserIDKey），供后续处理使用。
func Auth(tm *token.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			abortWithCode(c, code.ErrUnauthorized)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
			abortWithCode(c, code.ErrUnauthorized)
			return
		}

		userID, err := tm.Parse(parts[1])
		if err != nil {
			switch {
			case errors.Is(err, token.ErrTokenExpired):
				abortWithCode(c, code.ErrTokenExpired)
			default:
				abortWithCode(c, code.ErrTokenInvalid)
			}
			return
		}

		c.Set(ContextUserIDKey, userID)
		c.Next()
	}
}

func abortWithCode(c *gin.Context, bizCode int) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
		"code":    bizCode,
		"message": code.Text(bizCode),
	})
}
