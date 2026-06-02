package v1

import "github.com/furadx/iam-go/internal/apiserver/store"

// Service 定义所有服务接口。
type Service interface {
	Users() UserSrv
}

type service struct {
	store store.Factory
}

var _ Service = (*service)(nil)

// NewService 创建服务实例。
func NewService(store store.Factory) Service {
	return &service{
		store: store,
	}
}

func (s *service) Users() UserSrv {
	return newUsers(s)
}
