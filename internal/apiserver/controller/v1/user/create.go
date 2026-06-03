package user

import (
	"github.com/gin-gonic/gin"

	"github.com/furadx/iam-go/internal/apiserver/model"
	"github.com/furadx/iam-go/internal/pkg/code"
	"github.com/furadx/iam-go/internal/pkg/util"
	"github.com/furadx/iam-go/pkg/log"
)

// Create 创建用户。Controller 只负责绑参与出参，业务全部在 service 层。
func (u *UserController) Create(c *gin.Context) {
	log.Info("user create function called.")

	var r model.User
	if err := c.ShouldBindJSON(&r); err != nil {
		util.WriteResponse(c, code.WithCode(code.ErrBind, err), nil)
		return
	}

	if err := u.srv.Users().Create(c, &r, model.CreateOptions{}); err != nil {
		util.WriteResponse(c, err, nil)
		return
	}

	util.WriteResponse(c, nil, r)
}
