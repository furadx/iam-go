package user

import (
	"github.com/gin-gonic/gin"

	"github.com/furadx/iam-go/internal/apiserver/model"
	"github.com/furadx/iam-go/internal/pkg/code"
	"github.com/furadx/iam-go/internal/pkg/util"
	"github.com/furadx/iam-go/pkg/log"
)

// createRequest 是注册入参。只暴露可由客户端设置的字段，
// 防止通过请求体批量赋值 is_admin / status / logined_at 等特权/内部字段。
type createRequest struct {
	Name     string `json:"name" binding:"required,min=1,max=32"`
	Nickname string `json:"nickname"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Phone    string `json:"phone"`
}

// Create 创建用户。Controller 只负责绑参与出参，业务全部在 service 层。
func (u *UserController) Create(c *gin.Context) {
	log.Info("user create function called.")

	var r createRequest
	if err := c.ShouldBindJSON(&r); err != nil {
		util.WriteResponse(c, code.WithCode(code.ErrBind, err), nil)
		return
	}

	user := &model.User{
		Name:     r.Name,
		Nickname: r.Nickname,
		Password: r.Password,
		Email:    r.Email,
		Phone:    r.Phone,
	}

	if err := u.srv.Users().Create(c, user, model.CreateOptions{}); err != nil {
		util.WriteResponse(c, err, nil)
		return
	}

	util.WriteResponse(c, nil, user)
}
