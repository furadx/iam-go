package user

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/furadx/iam-go/internal/apiserver/model"
	"github.com/furadx/iam-go/internal/pkg/code"
	"github.com/furadx/iam-go/internal/pkg/util"
	"github.com/furadx/iam-go/pkg/auth"
	"github.com/furadx/iam-go/pkg/log"
)

// Create 创建用户。
func (u *UserController) Create(c *gin.Context) {
	log.Info("user create function called.")

	var r model.User
	if err := c.ShouldBindJSON(&r); err != nil {
		util.WriteResponse(c, code.WithCode(code.ErrBind, err), nil)
		return
	}

	// 加密密码
	hashedPassword, err := auth.Encrypt(r.Password)
	if err != nil {
		util.WriteResponse(c, code.WithCode(code.ErrEncrypt, err), nil)
		return
	}
	r.Password = hashedPassword
	r.Status = 1
	r.LoginedAt = time.Now()

	// 创建用户
	if err := u.srv.Users().Create(c, &r, model.CreateOptions{}); err != nil {
		util.WriteResponse(c, err, nil)
		return
	}

	// 清除密码字段后返回
	r.Password = ""
	util.WriteResponse(c, nil, r)
}
