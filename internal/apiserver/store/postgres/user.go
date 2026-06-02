package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/furadx/iam-go/internal/apiserver/model"
	"github.com/furadx/iam-go/internal/pkg/code"
)

type users struct {
	db *gorm.DB
}

func newUsers(ds *datastore) *users {
	return &users{ds.db}
}

// Create 创建新用户。
func (u *users) Create(ctx context.Context, user *model.User, opts model.CreateOptions) error {
	return u.db.Create(&user).Error
}

// Update 更新用户信息。
func (u *users) Update(ctx context.Context, user *model.User, opts model.UpdateOptions) error {
	return u.db.Save(user).Error
}

// Delete 删除用户。
func (u *users) Delete(ctx context.Context, username string, opts model.DeleteOptions) error {
	if opts.Unscoped {
		u.db = u.db.Unscoped()
	}

	err := u.db.Where("name = ?", username).Delete(&model.User{}).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return code.WithCode(code.ErrDatabase, err)
	}

	return nil
}

// Get 根据用户名获取用户。
func (u *users) Get(ctx context.Context, username string, opts model.GetOptions) (*model.User, error) {
	user := &model.User{}
	err := u.db.Where("name = ?", username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, code.WithCode(code.ErrUserNotFound, err)
		}
		return nil, code.WithCode(code.ErrDatabase, err)
	}

	return user, nil
}

// List 获取用户列表。
func (u *users) List(ctx context.Context, opts model.ListOptions) (*model.UserList, error) {
	ret := &model.UserList{}

	query := u.db.Where("status = 1")
	if opts.Name != "" {
		query = query.Where("name LIKE ?", "%"+opts.Name+"%")
	}

	if opts.Offset > 0 {
		query = query.Offset(int(opts.Offset))
	}
	if opts.Limit > 0 {
		query = query.Limit(int(opts.Limit))
	} else {
		query = query.Limit(20) // 默认限制 20 条
	}

	err := query.Order("id desc").
		Find(&ret.Items).
		Offset(-1).
		Limit(-1).
		Count(&ret.TotalCount).
		Error

	return ret, err
}
