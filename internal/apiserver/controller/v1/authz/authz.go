package authz

import (
	"github.com/gin-gonic/gin"

	authzpkg "github.com/furadx/iam-go/internal/pkg/authz"
	"github.com/furadx/iam-go/internal/pkg/code"
	"github.com/furadx/iam-go/internal/pkg/util"
)

// Controller 角色/策略管理控制器（仅 admin 可达，由路由 Authz 中间件保证）。
type Controller struct {
	m manager
}

type manager interface {
	AddPolicy(role, obj, act string) (bool, error)
	RemovePolicy(role, obj, act string) (bool, error)
	Policies() ([][]string, error)
	AssignRole(user, role string) (bool, error)
	RevokeRole(user, role string) (bool, error)
	RolesForUser(user string) ([]string, error)
}

// NewController 创建控制器。
func NewController(m *authzpkg.Manager) *Controller {
	return &Controller{m: m}
}

type policyRequest struct {
	Role   string `json:"role" binding:"required"`
	Path   string `json:"path" binding:"required"`
	Method string `json:"method" binding:"required"`
}

// AddPolicy 增加权限策略。
func (ctl *Controller) AddPolicy(c *gin.Context) {
	var r policyRequest
	if err := c.ShouldBindJSON(&r); err != nil {
		util.WriteResponse(c, code.WithCode(code.ErrBind, err), nil)
		return
	}
	added, err := ctl.m.AddPolicy(r.Role, r.Path, r.Method)
	if err != nil {
		util.WriteResponse(c, code.WithCode(code.ErrInternal, err), nil)
		return
	}
	util.WriteResponse(c, nil, gin.H{"added": added, "policy": r})
}

// RemovePolicy 删除权限策略。
func (ctl *Controller) RemovePolicy(c *gin.Context) {
	var r policyRequest
	if err := c.ShouldBindJSON(&r); err != nil {
		util.WriteResponse(c, code.WithCode(code.ErrBind, err), nil)
		return
	}
	removed, err := ctl.m.RemovePolicy(r.Role, r.Path, r.Method)
	if err != nil {
		util.WriteResponse(c, code.WithCode(code.ErrInternal, err), nil)
		return
	}
	util.WriteResponse(c, nil, gin.H{"removed": removed, "policy": r})
}

// ListPolicies 列出全部权限策略。
func (ctl *Controller) ListPolicies(c *gin.Context) {
	policies, err := ctl.m.Policies()
	if err != nil {
		util.WriteResponse(c, code.WithCode(code.ErrInternal, err), nil)
		return
	}
	util.WriteResponse(c, nil, gin.H{"policies": policies})
}

type roleRequest struct {
	Role string `json:"role" binding:"required"`
}

// AssignRole 给用户分配角色。
func (ctl *Controller) AssignRole(c *gin.Context) {
	name := c.Param("name")
	var r roleRequest
	if err := c.ShouldBindJSON(&r); err != nil {
		util.WriteResponse(c, code.WithCode(code.ErrBind, err), nil)
		return
	}
	assigned, err := ctl.m.AssignRole(name, r.Role)
	if err != nil {
		util.WriteResponse(c, code.WithCode(code.ErrInternal, err), nil)
		return
	}
	util.WriteResponse(c, nil, gin.H{"user": name, "role": r.Role, "assigned": assigned})
}

// RevokeRole 撤销用户角色。
func (ctl *Controller) RevokeRole(c *gin.Context) {
	name := c.Param("name")
	role := c.Param("role")
	revoked, err := ctl.m.RevokeRole(name, role)
	if err != nil {
		util.WriteResponse(c, code.WithCode(code.ErrInternal, err), nil)
		return
	}
	util.WriteResponse(c, nil, gin.H{"user": name, "role": role, "revoked": revoked})
}

// ListRoles 查询用户角色。
func (ctl *Controller) ListRoles(c *gin.Context) {
	name := c.Param("name")
	roles, err := ctl.m.RolesForUser(name)
	if err != nil {
		util.WriteResponse(c, code.WithCode(code.ErrInternal, err), nil)
		return
	}
	util.WriteResponse(c, nil, gin.H{"user": name, "roles": roles})
}
