package middleware

import (
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
