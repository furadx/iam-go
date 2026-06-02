package user

import (
	"github.com/gin-gonic/gin"

	"github.com/furadx/iam-go/internal/apiserver/model"
	"github.com/furadx/iam-go/internal/pkg/util"
)

// List 获取用户列表。
func (u *UserController) List(c *gin.Context) {
	var opts model.ListOptions
	if err := c.ShouldBindQuery(&opts); err != nil {
		util.WriteResponse(c, err, nil)
		return
	}

	users, err := u.srv.Users().List(c, opts)
	if err != nil {
		util.WriteResponse(c, err, nil)
		return
	}

	util.WriteResponse(c, nil, users)
}
