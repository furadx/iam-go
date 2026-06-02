package v1

import (
	"context"
	"regexp"

	"github.com/furadx/iam-go/internal/apiserver/model"
	"github.com/furadx/iam-go/internal/apiserver/store"
	"github.com/furadx/iam-go/internal/pkg/code"
	"github.com/furadx/iam-go/pkg/log"
)

// UserSrv 用户服务接口。
type UserSrv interface {
	Create(ctx context.Context, user *model.User, opts model.CreateOptions) error
	Update(ctx context.Context, user *model.User, opts model.UpdateOptions) error
	Delete(ctx context.Context, username string, opts model.DeleteOptions) error
	Get(ctx context.Context, username string, opts model.GetOptions) (*model.User, error)
	List(ctx context.Context, opts model.ListOptions) (*model.UserList, error)
}

type userService struct {
	store store.Factory
}

var _ UserSrv = (*userService)(nil)

func newUsers(srv *service) *userService {
	return &userService{store: srv.store}
}

// Create 创建用户。
func (u *userService) Create(ctx context.Context, user *model.User, opts model.CreateOptions) error {
	if err := u.store.Users().Create(ctx, user, opts); err != nil {
		// 检查是否是重复用户名错误
		if match, _ := regexp.MatchString("duplicate key value violates unique constraint|Duplicate entry", err.Error()); match {
			return code.WithCode(code.ErrUserAlreadyExist, err)
		}
		return code.WithCode(code.ErrDatabase, err)
	}

	return nil
}

// Update 更新用户。
func (u *userService) Update(ctx context.Context, user *model.User, opts model.UpdateOptions) error {
	if err := u.store.Users().Update(ctx, user, opts); err != nil {
		return code.WithCode(code.ErrDatabase, err)
	}

	return nil
}

// Delete 删除用户。
func (u *userService) Delete(ctx context.Context, username string, opts model.DeleteOptions) error {
	return u.store.Users().Delete(ctx, username, opts)
}

// Get 获取用户。
func (u *userService) Get(ctx context.Context, username string, opts model.GetOptions) (*model.User, error) {
	user, err := u.store.Users().Get(ctx, username, opts)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// List 获取用户列表。
func (u *userService) List(ctx context.Context, opts model.ListOptions) (*model.UserList, error) {
	users, err := u.store.Users().List(ctx, opts)
	if err != nil {
		log.Errorf("list users from storage failed: %s", err.Error())
		return nil, code.WithCode(code.ErrDatabase, err)
	}

	return users, nil
}
