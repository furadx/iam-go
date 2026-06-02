package user

import (
	"github.com/gin-gonic/gin"

	"github.com/furadx/iam-go/internal/apiserver/model"
	"github.com/furadx/iam-go/internal/pkg/util"
)

// Get 获取用户详情。
func (u *UserController) Get(c *gin.Context) {
	username := c.Param("name")

	user, err := u.srv.Users().Get(c, username, model.GetOptions{})
	if err != nil {
		util.WriteResponse(c, err, nil)
		return
	}

	// 清除密码字段
	user.Password = ""
	util.WriteResponse(c, nil, user)
}
