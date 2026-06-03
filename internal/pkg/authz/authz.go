package authz

import (
	"github.com/casbin/casbin/v3"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"
)

// Manager 封装 Casbin enforcer 与常用策略操作。
type Manager struct {
	enforcer *casbin.Enforcer
}

// NewManager 用现有 GORM 连接创建 enforcer（策略落 casbin_rule 表）。
func NewManager(db *gorm.DB, modelPath string) (*Manager, error) {
	adapter, err := gormadapter.NewAdapterByDB(db)
	if err != nil {
		return nil, err
	}
	e, err := casbin.NewEnforcer(modelPath, adapter)
	if err != nil {
		return nil, err
	}
	if err := e.LoadPolicy(); err != nil {
		return nil, err
	}
	return &Manager{enforcer: e}, nil
}

// Enforce 判定 (sub, obj, act) 是否放行。
func (m *Manager) Enforce(sub, obj, act string) (bool, error) {
	return m.enforcer.Enforce(sub, obj, act)
}

// AddPolicy 增加一条权限策略。
func (m *Manager) AddPolicy(role, obj, act string) (bool, error) {
	return m.enforcer.AddPolicy(role, obj, act)
}

// RemovePolicy 删除一条权限策略。
func (m *Manager) RemovePolicy(role, obj, act string) (bool, error) {
	return m.enforcer.RemovePolicy(role, obj, act)
}

// Policies 返回全部权限策略。
func (m *Manager) Policies() ([][]string, error) {
	return m.enforcer.GetPolicy()
}

// AssignRole 给用户分配角色。
func (m *Manager) AssignRole(user, role string) (bool, error) {
	return m.enforcer.AddGroupingPolicy(user, role)
}

// RevokeRole 撤销用户角色。
func (m *Manager) RevokeRole(user, role string) (bool, error) {
	return m.enforcer.RemoveGroupingPolicy(user, role)
}

// RolesForUser 返回用户拥有的角色。
func (m *Manager) RolesForUser(user string) ([]string, error) {
	return m.enforcer.GetRolesForUser(user)
}

// SeedDefaults 若无任何策略，写入默认 admin 全通配策略。
func (m *Manager) SeedDefaults() error {
	policies, err := m.enforcer.GetPolicy()
	if err != nil {
		return err
	}
	if len(policies) == 0 {
		_, err = m.enforcer.AddPolicy("admin", "/api/v1/*", "*")
		return err
	}
	return nil
}
