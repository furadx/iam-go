package user

import (
	"errors"
	"io"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/furadx/iam-go/internal/pkg/code"
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
	var r logoutRequest
	var refreshClaims *token.Claims
	if err := c.ShouldBindJSON(&r); err != nil && !errors.Is(err, io.EOF) {
		util.WriteResponse(c, code.WithCode(code.ErrBind, err), nil)
		return
	}
	if r.RefreshToken != "" {
		rc, err := u.tm.ParseTyped(r.RefreshToken, token.TypeRefresh)
		if err != nil {
			util.WriteResponse(c, code.New(code.ErrRefreshTokenInvalid), nil)
			return
		}
		refreshClaims = rc
	}

	if v, ok := c.Get(middleware.ContextClaimsKey); ok {
		if claims, ok := v.(*token.Claims); ok && claims.ExpiresAt != nil {
			if err := u.rv.Revoke(c.Request.Context(), claims.ID, time.Until(claims.ExpiresAt.Time)); err != nil {
				util.WriteResponse(c, code.WithCode(code.ErrInternal, err), nil)
				return
			}
		}
	}

	if refreshClaims != nil && refreshClaims.ExpiresAt != nil {
		if err := u.rv.Revoke(c.Request.Context(), refreshClaims.ID, time.Until(refreshClaims.ExpiresAt.Time)); err != nil {
			util.WriteResponse(c, code.WithCode(code.ErrInternal, err), nil)
			return
		}
	}

	util.WriteResponse(c, nil, gin.H{"message": "logged out"})
}
