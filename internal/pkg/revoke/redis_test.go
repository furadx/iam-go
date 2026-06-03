package revoke

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"

	"github.com/furadx/iam-go/pkg/token"
)

func setup(t *testing.T) (Revoker, *miniredis.Miniredis) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis: %v", err)
	}
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	return NewRedisRevoker(client), mr
}

func claim(uid int64, jti string, iat time.Time, exp time.Time) *token.Claims {
	return &token.Claims{
		UserID: uid,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			IssuedAt:  jwt.NewNumericDate(iat),
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}
}

func TestAllowedByDefault(t *testing.T) {
	rv, mr := setup(t)
	defer mr.Close()
	now := time.Now()
	ok, err := rv.Allowed(context.Background(), claim(1, "jti-1", now, now.Add(time.Hour)))
	if err != nil || !ok {
		t.Fatalf("expected allowed, got ok=%v err=%v", ok, err)
	}
}

func TestRevokeJTI(t *testing.T) {
	rv, mr := setup(t)
	defer mr.Close()
	ctx := context.Background()
	now := time.Now()
	c := claim(1, "jti-2", now, now.Add(time.Hour))
	if err := rv.Revoke(ctx, c.ID, time.Hour); err != nil {
		t.Fatalf("revoke: %v", err)
	}
	ok, _ := rv.Allowed(ctx, c)
	if ok {
		t.Fatalf("expected revoked jti to be disallowed")
	}
}

func TestRevokeAllBefore(t *testing.T) {
	rv, mr := setup(t)
	defer mr.Close()
	ctx := context.Background()
	now := time.Now()
	old := claim(7, "jti-old", now.Add(-time.Hour), now.Add(time.Hour))
	if err := rv.RevokeAllBefore(ctx, 7, now); err != nil {
		t.Fatalf("revokeAllBefore: %v", err)
	}
	if ok, _ := rv.Allowed(ctx, old); ok {
		t.Fatalf("expected old token to be disallowed")
	}
	fresh := claim(7, "jti-new", now.Add(time.Minute), now.Add(time.Hour))
	if ok, _ := rv.Allowed(ctx, fresh); !ok {
		t.Fatalf("expected fresh token to be allowed")
	}
}
