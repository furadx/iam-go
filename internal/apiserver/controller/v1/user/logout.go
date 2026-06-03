package user

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/furadx/iam-go/internal/pkg/middleware"
	"github.com/furadx/iam-go/internal/pkg/util"
	"github.com/furadx/iam-go/pkg/token"
)

type logoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// Logout 吊销当前 access token；若 body 带 refresh_token，一并吊销。
// 需经 Auth 中间件（依赖上下文里的 claims）。
func (u *UserController) Logout(c *gin.Context) {
	if v, ok := c.Get(middleware.ContextClaimsKey); ok {
		if claims, ok := v.(*token.Claims); ok && claims.ExpiresAt != nil {
			_ = u.rv.Revoke(c.Request.Context(), claims.ID, time.Until(claims.ExpiresAt.Time))
		}
	}

	var r logoutRequest
	if err := c.ShouldBindJSON(&r); err == nil && r.RefreshToken != "" {
		if rc, perr := u.tm.ParseTyped(r.RefreshToken, token.TypeRefresh); perr == nil && rc.ExpiresAt != nil {
			_ = u.rv.Revoke(c.Request.Context(), rc.ID, time.Until(rc.ExpiresAt.Time))
		}
	}

	util.WriteResponse(c, nil, gin.H{"message": "logged out"})
}
