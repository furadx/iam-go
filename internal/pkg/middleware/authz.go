package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/furadx/iam-go/internal/pkg/code"
)

// Enforcer 是 Authz 中间件依赖的鉴权判定接口（authz.Manager 实现之）。
type Enforcer interface {
	Enforce(sub, obj, act string) (bool, error)
}

// Authz 基于 Casbin 做 API 级鉴权（subject=username, object=path, action=method）。
// 必须在 Auth 之后使用（依赖上下文里的 username）。
func Authz(e Enforcer) gin.HandlerFunc {
	return func(c *gin.Context) {
		username := c.GetString(ContextUsernameKey)
		if username == "" {
			abortWithCode(c, http.StatusUnauthorized, code.ErrUnauthorized)
			return
		}
		ok, err := e.Enforce(username, c.Request.URL.Path, c.Request.Method)
		if err != nil {
			abortWithCode(c, http.StatusInternalServerError, code.ErrInternal)
			return
		}
		if !ok {
			abortWithCode(c, http.StatusForbidden, code.ErrPermissionDenied)
			return
		}
		c.Next()
	}
}
