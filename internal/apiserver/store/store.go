package store

import (
	"context"

	"github.com/furadx/iam-go/internal/apiserver/model"
)

// Factory 定义存储层接口。
type Factory interface {
	Users() UserStore
	Close() error
}

// UserStore 定义用户存储接口。
type UserStore interface {
	Create(ctx context.Context, user *model.User, opts model.CreateOptions) error
	Update(ctx context.Context, user *model.User, opts model.UpdateOptions) error
	Delete(ctx context.Context, username string, opts model.DeleteOptions) error
	Get(ctx context.Context, username string, opts model.GetOptions) (*model.User, error)
	List(ctx context.Context, opts model.ListOptions) (*model.UserList, error)
}
