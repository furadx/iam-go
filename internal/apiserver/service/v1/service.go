package v1

import (
	"context"
	"time"

	"github.com/furadx/iam-go/internal/apiserver/store"
)

// RoleAssigner 抽象角色分配（由 authz.Manager 实现），避免 service 直接依赖 casbin。
type RoleAssigner interface {
	AssignRole(user, role string) (bool, error)
}

// SessionRevoker 抽象会话吊销（由 revoke.Revoker 实现）。
type SessionRevoker interface {
	RevokeAllBefore(ctx context.Context, uid int64, t time.Time) error
}

// Service 定义所有服务接口。
type Service interface {
	Users() UserSrv
}

type service struct {
	store   store.Factory
	roles   RoleAssigner
	revoker SessionRevoker
}

var _ Service = (*service)(nil)

// NewService 创建服务实例。roles/revoker 可为 nil（测试或未启用时降级跳过相关副作用）。
func NewService(store store.Factory, roles RoleAssigner, revoker SessionRevoker) Service {
	return &service{
		store:   store,
		roles:   roles,
		revoker: revoker,
	}
}

func (s *service) Users() UserSrv {
	return newUsers(s)
}
