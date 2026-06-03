package user

import (
	"github.com/gin-gonic/gin"

	"github.com/furadx/iam-go/internal/pkg/code"
	"github.com/furadx/iam-go/internal/pkg/util"
)

// loginRequest 登录请求体。
type loginRequest struct {
	Name     string `json:"name" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login 校验用户名/密码，成功后签发 JWT。
func (u *UserController) Login(c *gin.Context) {
	var r loginRequest
	if err := c.ShouldBindJSON(&r); err != nil {
		util.WriteResponse(c, code.WithCode(code.ErrBind, err), nil)
		return
	}

	user, err := u.srv.Users().Login(c, r.Name, r.Password)
	if err != nil {
		util.WriteResponse(c, err, nil)
		return
	}

	tokenStr, err := u.tm.Sign(int64(user.ID))
	if err != nil {
		util.WriteResponse(c, code.WithCode(code.ErrSignToken, err), nil)
		return
	}

	util.WriteResponse(c, nil, gin.H{
		"token": tokenStr,
		"user":  user,
	})
}
