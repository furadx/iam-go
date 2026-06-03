package user

import (
	"github.com/gin-gonic/gin"

	"github.com/furadx/iam-go/internal/apiserver/model"
	"github.com/furadx/iam-go/internal/pkg/util"
)

// Get 获取用户详情。脱敏由 service 层统一处理。
func (u *UserController) Get(c *gin.Context) {
	username := c.Param("name")

	user, err := u.srv.Users().Get(c, username, model.GetOptions{})
	if err != nil {
		util.WriteResponse(c, err, nil)
		return
	}

	util.WriteResponse(c, nil, user)
}
