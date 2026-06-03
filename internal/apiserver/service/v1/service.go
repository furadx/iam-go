package v1

import (
	"context"
	"time"

	"github.com/furadx/iam-go/internal/apiserver/store"
	"github.com/furadx/iam-go/internal/pkg/password"
)

// RoleAssigner 抽象角色分配（由 authz.Manager 实现），避免 service 直接依赖 casbin。
type RoleAssigner interface {
	AssignRole(user, role string) (bool, error)
}

// SessionRevoker 抽象会话吊销（由 revoke.Revoker 实现）。
type SessionRevoker interface {
	RevokeAllBefore(ctx context.Context, uid int64, t time.Time) error
}

// LoginGuard 抽象登录失败锁定能力。
type LoginGuard interface {
	Check(ctx context.Context, username, ip string) error
	RecordFailure(ctx context.Context, username, ip string) error
	Reset(ctx context.Context, username, ip string) error
}

// Service 定义所有服务接口。
type Service interface {
	Users() UserSrv
}

type service struct {
	store          store.Factory
	roles          RoleAssigner
	revoker        SessionRevoker
	loginGuard     LoginGuard
	passwordPolicy password.Policy
}

var _ Service = (*service)(nil)

// NewService 创建服务实例。roles/revoker 可为 nil（测试或未启用时降级跳过相关副作用）。
func NewService(store store.Factory, roles RoleAssigner, revoker SessionRevoker, loginGuard LoginGuard, passwordPolicy password.Policy) Service {
	return &service{
		store:          store,
		roles:          roles,
		revoker:        revoker,
		loginGuard:     loginGuard,
		passwordPolicy: passwordPolicy,
	}
}

func (s *service) Users() UserSrv {
	return newUsers(s)
}
