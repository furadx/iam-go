package user

import (
	"github.com/gin-gonic/gin"

	"github.com/furadx/iam-go/internal/pkg/code"
	"github.com/furadx/iam-go/internal/pkg/util"
)

type loginRequest struct {
	Name     string `json:"name" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login 校验用户名/密码，成功后签发 access + refresh 双令牌。
func (u *UserController) Login(c *gin.Context) {
	var r loginRequest
	if err := c.ShouldBindJSON(&r); err != nil {
		util.WriteResponse(c, code.WithCode(code.ErrBind, err), nil)
		return
	}

	user, err := u.srv.Users().Login(c, r.Name, r.Password, c.ClientIP())
	if err != nil {
		util.WriteResponse(c, err, nil)
		return
	}

	access, _, err := u.tm.SignAccess(int64(user.ID), user.Name)
	if err != nil {
		util.WriteResponse(c, code.WithCode(code.ErrSignToken, err), nil)
		return
	}
	refresh, _, err := u.tm.SignRefresh(int64(user.ID), user.Name)
	if err != nil {
		util.WriteResponse(c, code.WithCode(code.ErrSignToken, err), nil)
		return
	}

	util.WriteResponse(c, nil, gin.H{
		"access_token":  access,
		"refresh_token": refresh,
		"expires_in":    u.tm.AccessExpireSeconds(),
		"user":          user,
	})
}
