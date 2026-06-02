package user

import (
	srvv1 "github.com/furadx/iam-go/internal/apiserver/service/v1"
	"github.com/furadx/iam-go/internal/apiserver/store"
)

// UserController 用户控制器。
type UserController struct {
	srv srvv1.Service
}

// NewUserController 创建用户控制器。
func NewUserController(store store.Factory) *UserController {
	return &UserController{
		srv: srvv1.NewService(store),
	}
}
