package user

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/furadx/iam-go/internal/pkg/code"
	"github.com/furadx/iam-go/internal/pkg/util"
	"github.com/furadx/iam-go/pkg/token"
)

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// Refresh 用 refresh token 换新 access，并轮转 refresh（旧 refresh 入黑名单）。
func (u *UserController) Refresh(c *gin.Context) {
	var r refreshRequest
	if err := c.ShouldBindJSON(&r); err != nil {
		util.WriteResponse(c, code.WithCode(code.ErrBind, err), nil)
		return
	}

	claims, err := u.tm.ParseTyped(r.RefreshToken, token.TypeRefresh)
	if err != nil {
		util.WriteResponse(c, code.New(code.ErrRefreshTokenInvalid), nil)
		return
	}

	if allowed, aerr := u.rv.Allowed(c.Request.Context(), claims); aerr == nil && !allowed {
		util.WriteResponse(c, code.New(code.ErrRefreshTokenInvalid), nil)
		return
	} else if aerr != nil {
		util.WriteResponse(c, code.WithCode(code.ErrInternal, aerr), nil)
		return
	}

	// 轮转：旧 refresh 拉黑（剩余寿命为 TTL）
	if claims.ExpiresAt != nil {
		if err := u.rv.Revoke(c.Request.Context(), claims.ID, time.Until(claims.ExpiresAt.Time)); err != nil {
			util.WriteResponse(c, code.WithCode(code.ErrInternal, err), nil)
			return
		}
	}

	access, _, err := u.tm.SignAccess(claims.UserID, claims.Username)
	if err != nil {
		util.WriteResponse(c, code.WithCode(code.ErrSignToken, err), nil)
		return
	}
	refresh, _, err := u.tm.SignRefresh(claims.UserID, claims.Username)
	if err != nil {
		util.WriteResponse(c, code.WithCode(code.ErrSignToken, err), nil)
		return
	}

	util.WriteResponse(c, nil, gin.H{
		"access_token":  access,
		"refresh_token": refresh,
		"expires_in":    u.tm.AccessExpireSeconds(),
	})
}
