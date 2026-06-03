package v1

import (
	"context"
	"errors"
	"time"

	"github.com/furadx/iam-go/internal/apiserver/model"
	"github.com/furadx/iam-go/internal/apiserver/store"
	"github.com/furadx/iam-go/internal/pkg/code"
	"github.com/furadx/iam-go/internal/pkg/loginlock"
	"github.com/furadx/iam-go/internal/pkg/password"
	"github.com/furadx/iam-go/pkg/auth"
	"github.com/furadx/iam-go/pkg/log"
)

const (
	defaultListLimit int64 = 20
	maxListLimit     int64 = 200
)

// UserSrv 用户服务接口。承担：参数业务校验、密码加解密、字段白名单、出参脱敏。
type UserSrv interface {
	Create(ctx context.Context, user *model.User, opts model.CreateOptions) error
	Update(ctx context.Context, user *model.User, opts model.UpdateOptions) error
	Delete(ctx context.Context, username string, opts model.DeleteOptions) error
	Get(ctx context.Context, username string, opts model.GetOptions) (*model.User, error)
	List(ctx context.Context, opts model.ListOptions) (*model.UserList, error)

	// ChangePassword 校验旧密码后写入新密码。
	ChangePassword(ctx context.Context, username, oldPassword, newPassword string) error
	// Login 校验用户名/密码并刷新 LoginedAt，返回脱敏后的用户。
	Login(ctx context.Context, username, password, ip string) (*model.User, error)
}

type userService struct {
	store          store.Factory
	roles          RoleAssigner
	revoker        SessionRevoker
	loginGuard     LoginGuard
	passwordPolicy password.Policy
}

var _ UserSrv = (*userService)(nil)

func newUsers(srv *service) *userService {
	return &userService{
		store:          srv.store,
		roles:          srv.roles,
		revoker:        srv.revoker,
		loginGuard:     srv.loginGuard,
		passwordPolicy: srv.passwordPolicy,
	}
}

// Create 创建用户。
// 业务：密码强度校验 -> 用户名预检 -> 密码加密 -> 默认值 -> 落库 -> 出参脱敏。
func (u *userService) Create(ctx context.Context, user *model.User, opts model.CreateOptions) error {
	if err := u.validatePassword(user.Name, user.Password); err != nil {
		return err
	}

	// 业务级预检：友好错误。并发竞态由 store 层翻译唯一约束错误兜底。
	if existing, _ := u.store.Users().Get(ctx, user.Name, model.GetOptions{}); existing != nil {
		return code.New(code.ErrUserAlreadyExist)
	}

	hashed, err := auth.Encrypt(user.Password)
	if err != nil {
		return code.WithCode(code.ErrEncrypt, err)
	}
	user.Password = hashed
	if user.Status == 0 {
		user.Status = 1
	}
	if user.LoginedAt.IsZero() {
		user.LoginedAt = time.Now()
	}

	if err := u.store.Users().Create(ctx, user, opts); err != nil {
		return err
	}
	if u.roles != nil {
		if _, err := u.roles.AssignRole(user.Name, "user"); err != nil {
			log.Warnf("assign default role to %s failed: %s", user.Name, err.Error())
		}
	}
	user.Password = ""
	return nil
}

// Update 更新用户。仅允许白名单字段，避免 gorm Save 全量覆盖踩坑。
// 改名走专用接口；改密码走 ChangePassword。
func (u *userService) Update(ctx context.Context, user *model.User, opts model.UpdateOptions) error {
	existing, err := u.store.Users().Get(ctx, user.Name, model.GetOptions{})
	if err != nil {
		return err
	}

	existing.Nickname = user.Nickname
	existing.Email = user.Email
	existing.Phone = user.Phone
	existing.Status = user.Status
	existing.IsAdmin = user.IsAdmin

	if err := u.store.Users().Update(ctx, existing, opts); err != nil {
		return code.WithCode(code.ErrDatabase, err)
	}
	return nil
}

// Delete 删除用户。先确认存在再删，让 404 与 200 语义分明。
func (u *userService) Delete(ctx context.Context, username string, opts model.DeleteOptions) error {
	if _, err := u.store.Users().Get(ctx, username, model.GetOptions{}); err != nil {
		return err
	}
	return u.store.Users().Delete(ctx, username, opts)
}

// Get 获取用户。返回前脱敏。
func (u *userService) Get(ctx context.Context, username string, opts model.GetOptions) (*model.User, error) {
	user, err := u.store.Users().Get(ctx, username, opts)
	if err != nil {
		return nil, err
	}
	user.Password = ""
	return user, nil
}

// List 获取用户列表。补齐分页默认值与上限，逐条脱敏。
func (u *userService) List(ctx context.Context, opts model.ListOptions) (*model.UserList, error) {
	if opts.Offset < 0 {
		opts.Offset = 0
	}
	if opts.Limit <= 0 {
		opts.Limit = defaultListLimit
	}
	if opts.Limit > maxListLimit {
		opts.Limit = maxListLimit
	}

	list, err := u.store.Users().List(ctx, opts)
	if err != nil {
		log.Errorf("list users from storage failed: %s", err.Error())
		return nil, code.WithCode(code.ErrDatabase, err)
	}
	for i := range list.Items {
		list.Items[i].Password = ""
	}
	return list, nil
}

// ChangePassword 修改密码。
func (u *userService) ChangePassword(ctx context.Context, username, oldPassword, newPassword string) error {
	user, err := u.store.Users().Get(ctx, username, model.GetOptions{})
	if err != nil {
		return err
	}
	if err := auth.Compare(user.Password, oldPassword); err != nil {
		return code.WithCode(code.ErrPasswordIncorrect, err)
	}
	if err := u.validatePassword(username, newPassword); err != nil {
		return err
	}
	hashed, err := auth.Encrypt(newPassword)
	if err != nil {
		return code.WithCode(code.ErrEncrypt, err)
	}
	user.Password = hashed
	if err := u.store.Users().Update(ctx, user, model.UpdateOptions{}); err != nil {
		return code.WithCode(code.ErrDatabase, err)
	}
	if u.revoker != nil {
		if err := u.revoker.RevokeAllBefore(ctx, int64(user.ID), time.Now()); err != nil {
			log.Warnf("revoke sessions after password change failed: %s", err.Error())
		}
	}
	return nil
}

// Login 校验用户名/密码并刷新 LoginedAt。
func (u *userService) Login(ctx context.Context, username, password, ip string) (*model.User, error) {
	if err := u.checkLoginLock(ctx, username, ip); err != nil {
		return nil, err
	}
	user, err := u.store.Users().Get(ctx, username, model.GetOptions{})
	if err != nil {
		if code.Code(err) == code.ErrUserNotFound {
			u.recordLoginFailure(ctx, username, ip)
			return nil, code.New(code.ErrPasswordIncorrect)
		}
		return nil, err
	}
	if user.Status != 1 {
		u.recordLoginFailure(ctx, username, ip)
		return nil, code.New(code.ErrPasswordIncorrect)
	}
	if err := auth.Compare(user.Password, password); err != nil {
		u.recordLoginFailure(ctx, username, ip)
		return nil, code.WithCode(code.ErrPasswordIncorrect, err)
	}

	user.LoginedAt = time.Now()
	if err := u.store.Users().Update(ctx, user, model.UpdateOptions{}); err != nil {
		// 主流程已成功，不阻断登录。
		log.Warnf("update LoginedAt failed: %s", err.Error())
	}
	u.resetLoginFailures(ctx, username, ip)
	user.Password = ""
	return user, nil
}

func (u *userService) checkLoginLock(ctx context.Context, username, ip string) error {
	if u.loginGuard == nil {
		return nil
	}
	if err := u.loginGuard.Check(ctx, username, ip); err != nil {
		if errors.Is(err, loginlock.ErrLocked) {
			return code.New(code.ErrLoginLocked)
		}
		log.Warnf("check login lock failed: %s", err.Error())
	}
	return nil
}

func (u *userService) recordLoginFailure(ctx context.Context, username, ip string) {
	if u.loginGuard == nil {
		return
	}
	if err := u.loginGuard.RecordFailure(ctx, username, ip); err != nil {
		log.Warnf("record login failure failed: %s", err.Error())
	}
}

func (u *userService) resetLoginFailures(ctx context.Context, username, ip string) {
	if u.loginGuard == nil {
		return
	}
	if err := u.loginGuard.Reset(ctx, username, ip); err != nil {
		log.Warnf("reset login failures failed: %s", err.Error())
	}
}

func (u *userService) validatePassword(username, value string) error {
	policy := u.passwordPolicy
	if policy.MinLength == 0 && policy.MaxLength == 0 {
		policy = password.DefaultPolicy()
	}
	err := policy.Validate(username, value)
	switch {
	case err == nil:
		return nil
	case errors.Is(err, password.ErrTooShort):
		return code.New(code.ErrPasswordTooShort)
	default:
		return code.New(code.ErrPasswordTooWeak)
	}
}
