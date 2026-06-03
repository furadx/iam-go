package authz

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/furadx/iam-go/internal/pkg/code"
)

type fakeManager struct {
	addPolicy    bool
	removePolicy bool
	assignRole   bool
	revokeRole   bool
}

func (m *fakeManager) AddPolicy(role, obj, act string) (bool, error) {
	return m.addPolicy, nil
}

func (m *fakeManager) RemovePolicy(role, obj, act string) (bool, error) {
	return m.removePolicy, nil
}

func (m *fakeManager) Policies() ([][]string, error) {
	return [][]string{{"admin", "/api/v1/*", "*"}}, nil
}

func (m *fakeManager) AssignRole(user, role string) (bool, error) {
	return m.assignRole, nil
}

func (m *fakeManager) RevokeRole(user, role string) (bool, error) {
	return m.revokeRole, nil
}

func (m *fakeManager) RolesForUser(user string) ([]string, error) {
	return []string{"admin"}, nil
}

func authzResponseCode(t *testing.T, body string) int {
	t.Helper()
	var resp struct {
		Code int `json:"code"`
	}
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	return resp.Code
}

func TestAddPolicyRejectsInvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctl := &Controller{m: &fakeManager{}}
	r := gin.New()
	r.POST("/authz/policies", ctl.AddPolicy)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/authz/policies", strings.NewReader(`{"role":"admin"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if got := authzResponseCode(t, w.Body.String()); got != code.ErrBind {
		t.Fatalf("expected ErrBind, got %d body=%s", got, w.Body.String())
	}
}

func TestAddPolicyReportsCasbinNoop(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctl := &Controller{m: &fakeManager{addPolicy: false}}
	r := gin.New()
	r.POST("/authz/policies", ctl.AddPolicy)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/authz/policies", strings.NewReader(`{"role":"admin","path":"/api/v1/*","method":"*"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	var resp struct {
		Code int `json:"code"`
		Data struct {
			Added  bool          `json:"added"`
			Policy policyRequest `json:"policy"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Code != code.OK {
		t.Fatalf("expected OK, got %d body=%s", resp.Code, w.Body.String())
	}
	if resp.Data.Added {
		t.Fatalf("expected no-op added=false")
	}
	if resp.Data.Policy.Role != "admin" || resp.Data.Policy.Path != "/api/v1/*" || resp.Data.Policy.Method != "*" {
		t.Fatalf("unexpected policy echo: %#v", resp.Data.Policy)
	}
}

func TestAssignRoleReportsCasbinNoop(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctl := &Controller{m: &fakeManager{assignRole: false}}
	r := gin.New()
	r.POST("/users/:name/roles", ctl.AssignRole)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/users/alice/roles", strings.NewReader(`{"role":"admin"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	var resp struct {
		Code int `json:"code"`
		Data struct {
			User     string `json:"user"`
			Role     string `json:"role"`
			Assigned bool   `json:"assigned"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Code != code.OK {
		t.Fatalf("expected OK, got %d body=%s", resp.Code, w.Body.String())
	}
	if resp.Data.User != "alice" || resp.Data.Role != "admin" || resp.Data.Assigned {
		t.Fatalf("unexpected assign role response: %#v", resp.Data)
	}
}
