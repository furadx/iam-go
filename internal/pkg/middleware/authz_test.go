package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type stubEnforcer struct {
	allow bool
	err   error
}

func (s stubEnforcer) Enforce(sub, obj, act string) (bool, error) { return s.allow, s.err }

func newAuthzRouter(e Enforcer, username string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/v1/users", func(c *gin.Context) {
		c.Set(XRequestIDKey, "rid-test")
		c.Set(ContextUsernameKey, username)
		c.Next()
	}, Authz(e), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	return r
}

func TestAuthzAllow(t *testing.T) {
	r := newAuthzRouter(stubEnforcer{allow: true}, "alice")
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/users", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestAuthzForbidden(t *testing.T) {
	r := newAuthzRouter(stubEnforcer{allow: false}, "bob")
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/users", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
	var body struct {
		Code      int    `json:"code"`
		Message   string `json:"message"`
		RequestID string `json:"request_id"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("expected JSON body, got %q: %v", w.Body.String(), err)
	}
	if body.Code != 110007 || body.Message == "" || body.RequestID != "rid-test" {
		t.Fatalf("unexpected forbidden body: %#v", body)
	}
}

func TestAuthzNoUsername(t *testing.T) {
	r := newAuthzRouter(stubEnforcer{allow: true}, "")
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/users", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}
