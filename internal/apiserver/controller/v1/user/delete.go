package user

import (
	"github.com/gin-gonic/gin"

	"github.com/furadx/iam-go/internal/apiserver/model"
	"github.com/furadx/iam-go/internal/pkg/util"
)

// Delete removes a user. Service/store decide whether deletion is soft or hard.
func (u *UserController) Delete(c *gin.Context) {
	username := c.Param("name")

	if err := u.srv.Users().Delete(c, username, model.DeleteOptions{}); err != nil {
		util.WriteResponse(c, err, nil)
		return
	}

	util.WriteResponse(c, nil, nil)
}
