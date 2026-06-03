package user

import (
	srvv1 "github.com/furadx/iam-go/internal/apiserver/service/v1"
	"github.com/furadx/iam-go/internal/apiserver/store"
	"github.com/furadx/iam-go/pkg/token"
)

// UserController 用户控制器。
type UserController struct {
	srv srvv1.Service
	tm  *token.Manager
}

// NewUserController 创建用户控制器。
func NewUserController(store store.Factory, tm *token.Manager) *UserController {
	return &UserController{
		srv: srvv1.NewService(store),
		tm:  tm,
	}
}
