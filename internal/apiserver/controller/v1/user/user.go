package user

import (
	srvv1 "github.com/furadx/iam-go/internal/apiserver/service/v1"
	"github.com/furadx/iam-go/internal/apiserver/store"
	"github.com/furadx/iam-go/internal/pkg/authz"
	"github.com/furadx/iam-go/internal/pkg/password"
	"github.com/furadx/iam-go/internal/pkg/revoke"
	"github.com/furadx/iam-go/pkg/token"
)

// UserController 用户控制器。
type UserController struct {
	srv srvv1.Service
	tm  *token.Manager
	rv  revoke.Revoker
}

// Deps 是 UserController 的装配依赖。
type Deps struct {
	Store          store.Factory
	Token          *token.Manager
	Revoker        revoke.Revoker
	Authz          *authz.Manager
	LoginGuard     srvv1.LoginGuard
	PasswordPolicy password.Policy
}

// NewUserController 创建用户控制器。
func NewUserController(deps Deps) *UserController {
	return &UserController{
		srv: srvv1.NewService(deps.Store, deps.Authz, deps.Revoker, deps.LoginGuard, deps.PasswordPolicy),
		tm:  deps.Token,
		rv:  deps.Revoker,
	}
}
