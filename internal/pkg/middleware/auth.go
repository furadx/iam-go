package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/furadx/iam-go/internal/pkg/code"
	"github.com/furadx/iam-go/internal/pkg/revoke"
	"github.com/furadx/iam-go/pkg/log"
	"github.com/furadx/iam-go/pkg/token"
)

// 认证后写入上下文的键。
const (
	ContextUserIDKey   = "user_id"
	ContextUsernameKey = "username"
	ContextClaimsKey   = "claims"
)

// Auth 是基于 JWT 的认证中间件：校验 access token 与吊销状态，
// 通过后把用户信息与 claims 写入上下文。
// failOpen 为 true 时，吊销存储（Redis）故障会放行并记日志。
func Auth(tm *token.Manager, rv revoke.Revoker, failOpen bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			abortWithCode(c, http.StatusUnauthorized, code.ErrUnauthorized)
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
			abortWithCode(c, http.StatusUnauthorized, code.ErrUnauthorized)
			return
		}

		claims, err := tm.ParseTyped(parts[1], token.TypeAccess)
		if err != nil {
			switch {
			case errors.Is(err, token.ErrTokenExpired):
				abortWithCode(c, http.StatusUnauthorized, code.ErrTokenExpired)
			case errors.Is(err, token.ErrWrongTokenType):
				abortWithCode(c, http.StatusUnauthorized, code.ErrTokenTypeInvalid)
			default:
				abortWithCode(c, http.StatusUnauthorized, code.ErrTokenInvalid)
			}
			return
		}

		allowed, rerr := rv.Allowed(c.Request.Context(), claims)
		if rerr != nil {
			if !failOpen {
				abortWithCode(c, http.StatusUnauthorized, code.ErrUnauthorized)
				return
			}
			log.Warnf("revoke store unavailable, fail-open allows request: %v", rerr)
		} else if !allowed {
			abortWithCode(c, http.StatusUnauthorized, code.ErrUnauthorized)
			return
		}

		c.Set(ContextUserIDKey, claims.UserID)
		c.Set(ContextUsernameKey, claims.Username)
		c.Set(ContextClaimsKey, claims)
		c.Next()
	}
}

func abortWithCode(c *gin.Context, status, bizCode int) {
	c.AbortWithStatusJSON(status, gin.H{
		"code":    bizCode,
		"message": code.Text(bizCode),
	})
}
