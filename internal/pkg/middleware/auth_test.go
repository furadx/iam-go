package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/furadx/iam-go/pkg/token"
)

type stubRevoker struct {
	allowed bool
	err     error
}

func (s stubRevoker) Revoke(ctx context.Context, jti string, ttl time.Duration) error   { return nil }
func (s stubRevoker) RevokeAllBefore(ctx context.Context, uid int64, t time.Time) error { return nil }
func (s stubRevoker) Allowed(ctx context.Context, c *token.Claims) (bool, error) {
	return s.allowed, s.err
}

func newTestRouter(tm *token.Manager, rv stubRevoker, failOpen bool) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/x", Auth(tm, rv, failOpen), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"uid": c.GetInt64(ContextUserIDKey), "name": c.GetString(ContextUsernameKey)})
	})
	return r
}

func TestAuthMissingHeader(t *testing.T) {
	tm := token.NewManager("s", time.Hour, time.Hour)
	r := newTestRouter(tm, stubRevoker{allowed: true}, true)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/x", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestAuthValidToken(t *testing.T) {
	tm := token.NewManager("s", time.Hour, time.Hour)
	tok, _, _ := tm.SignAccess(99, "alice")
	r := newTestRouter(tm, stubRevoker{allowed: true}, true)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (%s)", w.Code, w.Body.String())
	}
}

func TestAuthRejectsRefreshToken(t *testing.T) {
	tm := token.NewManager("s", time.Hour, time.Hour)
	refresh, _, _ := tm.SignRefresh(1, "x")
	r := newTestRouter(tm, stubRevoker{allowed: true}, true)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Bearer "+refresh)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for refresh-as-access, got %d", w.Code)
	}
}

func TestAuthRevoked(t *testing.T) {
	tm := token.NewManager("s", time.Hour, time.Hour)
	tok, _, _ := tm.SignAccess(1, "x")
	r := newTestRouter(tm, stubRevoker{allowed: false}, true)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for revoked, got %d", w.Code)
	}
}

func TestAuthFailOpen(t *testing.T) {
	tm := token.NewManager("s", time.Hour, time.Hour)
	tok, _, _ := tm.SignAccess(1, "x")
	r := newTestRouter(tm, stubRevoker{err: context.DeadlineExceeded}, true)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 with fail-open, got %d", w.Code)
	}
}
