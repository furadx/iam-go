package model

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型。
type User struct {
	ID        uint64         `gorm:"primaryKey;column:id"    json:"id"`
	Name      string         `gorm:"column:name;uniqueIndex:idx_name;not null" json:"name" binding:"required,min=1,max=32"`
	Nickname  string         `gorm:"column:nickname" json:"nickname"`
	Password  string         `gorm:"column:password;not null" json:"-"`
	Email     string         `gorm:"column:email" json:"email" binding:"required,email"`
	Phone     string         `gorm:"column:phone" json:"phone"`
	IsAdmin   int            `gorm:"column:is_admin;default:0" json:"isAdmin"`
	Status    int            `gorm:"column:status;default:1" json:"status"`
	LoginedAt time.Time      `gorm:"column:logined_at" json:"loginedAt"`
	CreatedAt time.Time      `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt time.Time      `gorm:"column:updated_at" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

// TableName 指定表名。
func (u *User) TableName() string {
	return "users"
}

// UserList 用户列表。
type UserList struct {
	TotalCount int64   `json:"totalCount"`
	Items      []*User `json:"items"`
}

// ListOptions 列表查询选项。
type ListOptions struct {
	Offset int64  `form:"offset"`
	Limit  int64  `form:"limit"`
	Name   string `form:"name"`
}

// CreateOptions 创建选项。
type CreateOptions struct{}

// UpdateOptions 更新选项。
type UpdateOptions struct{}

// GetOptions 获取选项。
type GetOptions struct{}

// DeleteOptions 删除选项。
type DeleteOptions struct {
	Unscoped bool
}
