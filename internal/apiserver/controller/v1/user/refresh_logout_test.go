package user

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/furadx/iam-go/internal/pkg/code"
	"github.com/furadx/iam-go/internal/pkg/middleware"
	"github.com/furadx/iam-go/pkg/token"
)

type controllerRevoker struct {
	allowed    bool
	allowedErr error
	revokeErr  error
	revoked    []string
}

func (r *controllerRevoker) Revoke(ctx context.Context, jti string, ttl time.Duration) error {
	r.revoked = append(r.revoked, jti)
	return r.revokeErr
}

func (r *controllerRevoker) RevokeAllBefore(ctx context.Context, uid int64, t time.Time) error {
	return nil
}

func (r *controllerRevoker) Allowed(ctx context.Context, claims *token.Claims) (bool, error) {
	return r.allowed, r.allowedErr
}

func responseCode(t *testing.T, body string) int {
	t.Helper()
	var resp struct {
		Code int `json:"code"`
	}
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	return resp.Code
}

func TestRefreshRotatesRefreshToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tm := token.NewManager("secret", time.Hour, time.Hour)
	refresh, claims, err := tm.SignRefresh(1, "alice")
	if err != nil {
		t.Fatalf("sign refresh: %v", err)
	}
	rv := &controllerRevoker{allowed: true}
	ctl := &UserController{tm: tm, rv: rv}
	r := gin.New()
	r.POST("/refresh", ctl.Refresh)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/refresh", strings.NewReader(`{"refresh_token":"`+refresh+`"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if got := responseCode(t, w.Body.String()); got != code.OK {
		t.Fatalf("expected OK, got %d body=%s", got, w.Body.String())
	}
	if len(rv.revoked) != 1 || rv.revoked[0] != claims.ID {
		t.Fatalf("expected old refresh jti %q to be revoked, got %#v", claims.ID, rv.revoked)
	}
	var resp struct {
		Data struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Data.AccessToken == "" || resp.Data.RefreshToken == "" {
		t.Fatalf("expected new access and refresh tokens, body=%s", w.Body.String())
	}
	if resp.Data.RefreshToken == refresh {
		t.Fatalf("expected refresh token to rotate")
	}
}

func TestRefreshFailsWhenRevokerAllowedErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tm := token.NewManager("secret", time.Hour, time.Hour)
	refresh, _, err := tm.SignRefresh(1, "alice")
	if err != nil {
		t.Fatalf("sign refresh: %v", err)
	}

	ctl := &UserController{
		tm: tm,
		rv: &controllerRevoker{allowedErr: context.DeadlineExceeded},
	}
	r := gin.New()
	r.POST("/refresh", ctl.Refresh)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/refresh", strings.NewReader(`{"refresh_token":"`+refresh+`"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if got := responseCode(t, w.Body.String()); got != code.ErrInternal {
		t.Fatalf("expected ErrInternal, got %d body=%s", got, w.Body.String())
	}
}

func TestRefreshFailsWhenOldRefreshCannotBeRevoked(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tm := token.NewManager("secret", time.Hour, time.Hour)
	refresh, _, err := tm.SignRefresh(1, "alice")
	if err != nil {
		t.Fatalf("sign refresh: %v", err)
	}

	ctl := &UserController{
		tm: tm,
		rv: &controllerRevoker{allowed: true, revokeErr: errors.New("redis down")},
	}
	r := gin.New()
	r.POST("/refresh", ctl.Refresh)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/refresh", strings.NewReader(`{"refresh_token":"`+refresh+`"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if got := responseCode(t, w.Body.String()); got != code.ErrInternal {
		t.Fatalf("expected ErrInternal, got %d body=%s", got, w.Body.String())
	}
}

func TestLogoutRejectsInvalidRefreshToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tm := token.NewManager("secret", time.Hour, time.Hour)
	claims := &token.Claims{
		UserID:   1,
		Username: "alice",
		Type:     token.TypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        "access-jti",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}

	ctl := &UserController{
		tm: tm,
		rv: &controllerRevoker{},
	}
	r := gin.New()
	r.POST("/logout", func(c *gin.Context) {
		c.Set(middleware.ContextClaimsKey, claims)
		ctl.Logout(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/logout", strings.NewReader(`{"refresh_token":"not-a-token"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if got := responseCode(t, w.Body.String()); got != code.ErrRefreshTokenInvalid {
		t.Fatalf("expected ErrRefreshTokenInvalid, got %d body=%s", got, w.Body.String())
	}
}

func TestLogoutRejectsMalformedBodyBeforeRevokingAccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tm := token.NewManager("secret", time.Hour, time.Hour)
	claims := &token.Claims{
		UserID:   1,
		Username: "alice",
		Type:     token.TypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        "access-jti",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	rv := &controllerRevoker{}
	ctl := &UserController{tm: tm, rv: rv}
	r := gin.New()
	r.POST("/logout", func(c *gin.Context) {
		c.Set(middleware.ContextClaimsKey, claims)
		ctl.Logout(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/logout", strings.NewReader(`{"refresh_token":`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if got := responseCode(t, w.Body.String()); got != code.ErrBind {
		t.Fatalf("expected ErrBind, got %d body=%s", got, w.Body.String())
	}
	if len(rv.revoked) != 0 {
		t.Fatalf("expected no tokens to be revoked on malformed body, got %#v", rv.revoked)
	}
}
