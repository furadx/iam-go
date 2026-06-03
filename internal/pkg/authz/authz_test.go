package authz

import (
	"path/filepath"
	"testing"

	"github.com/casbin/casbin/v3"
)

func newTestEnforcer(t *testing.T) *casbin.Enforcer {
	t.Helper()
	modelPath := filepath.Join("..", "..", "..", "configs", "rbac_model.conf")
	e, err := casbin.NewEnforcer(modelPath)
	if err != nil {
		t.Fatalf("new enforcer: %v", err)
	}
	if _, err := e.AddPolicy("admin", "/api/v1/*", "*"); err != nil {
		t.Fatalf("add policy: %v", err)
	}
	if _, err := e.AddPolicy("user", "/api/v1/users/:name", "GET"); err != nil {
		t.Fatalf("add policy: %v", err)
	}
	if _, err := e.AddGroupingPolicy("alice", "admin"); err != nil {
		t.Fatalf("add grouping: %v", err)
	}
	if _, err := e.AddGroupingPolicy("bob", "user"); err != nil {
		t.Fatalf("add grouping: %v", err)
	}
	return e
}

func TestModelMatching(t *testing.T) {
	e := newTestEnforcer(t)
	cases := []struct {
		sub, obj, act string
		want          bool
	}{
		{"alice", "/api/v1/users", "POST", true},
		{"alice", "/api/v1/authz/policies", "DELETE", true},
		{"bob", "/api/v1/users/bob", "GET", true},
		{"bob", "/api/v1/users", "POST", false},
		{"bob", "/api/v1/users/bob", "DELETE", false},
		{"carol", "/api/v1/users", "GET", false},
	}
	for _, c := range cases {
		got, err := e.Enforce(c.sub, c.obj, c.act)
		if err != nil {
			t.Fatalf("enforce(%s,%s,%s): %v", c.sub, c.obj, c.act, err)
		}
		if got != c.want {
			t.Errorf("enforce(%s,%s,%s)=%v, want %v", c.sub, c.obj, c.act, got, c.want)
		}
	}
}
