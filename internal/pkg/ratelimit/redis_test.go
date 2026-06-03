package ratelimit

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestAllowBlocksAfterLimit(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis: %v", err)
	}
	defer mr.Close()
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	limiter := NewRedisLimiter(client)
	limiter.now = func() time.Time { return time.Unix(100, 0) }
	ctx := context.Background()

	first, err := limiter.Allow(ctx, "api:ip:127.0.0.1", 2, time.Minute)
	if err != nil {
		t.Fatalf("first allow: %v", err)
	}
	if !first.Allowed || first.Remaining != 1 {
		t.Fatalf("unexpected first result: %#v", first)
	}
	second, err := limiter.Allow(ctx, "api:ip:127.0.0.1", 2, time.Minute)
	if err != nil {
		t.Fatalf("second allow: %v", err)
	}
	if !second.Allowed || second.Remaining != 0 {
		t.Fatalf("unexpected second result: %#v", second)
	}
	third, err := limiter.Allow(ctx, "api:ip:127.0.0.1", 2, time.Minute)
	if err != nil {
		t.Fatalf("third allow: %v", err)
	}
	if third.Allowed || third.Remaining != 0 {
		t.Fatalf("expected third request to be blocked, got %#v", third)
	}
}

func TestAllowResetsAcrossWindows(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis: %v", err)
	}
	defer mr.Close()
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	limiter := NewRedisLimiter(client)
	ctx := context.Background()

	limiter.now = func() time.Time { return time.Unix(100, 0) }
	if _, err := limiter.Allow(ctx, "login:ip:127.0.0.1", 1, time.Minute); err != nil {
		t.Fatalf("first allow: %v", err)
	}
	if blocked, err := limiter.Allow(ctx, "login:ip:127.0.0.1", 1, time.Minute); err != nil || blocked.Allowed {
		t.Fatalf("expected blocked in same window, got %#v err=%v", blocked, err)
	}

	limiter.now = func() time.Time { return time.Unix(180, 0) }
	next, err := limiter.Allow(ctx, "login:ip:127.0.0.1", 1, time.Minute)
	if err != nil {
		t.Fatalf("next window allow: %v", err)
	}
	if !next.Allowed {
		t.Fatalf("expected next window to allow, got %#v", next)
	}
}
