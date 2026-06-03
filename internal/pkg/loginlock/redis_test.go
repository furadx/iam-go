package loginlock

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func setupLocker(t *testing.T) (*Locker, *miniredis.Miniredis) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis: %v", err)
	}
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	locker := NewRedisLocker(client, Config{
		Enabled:         true,
		UserMaxFailures: 2,
		IPMaxFailures:   10,
		FailureWindow:   time.Minute,
		LockDuration:    15 * time.Minute,
	})
	return locker, mr
}

func TestRecordFailureLocksUserAtThreshold(t *testing.T) {
	locker, mr := setupLocker(t)
	defer mr.Close()
	ctx := context.Background()

	if err := locker.RecordFailure(ctx, "Alice", "127.0.0.1"); err != nil {
		t.Fatalf("record first failure: %v", err)
	}
	if err := locker.Check(ctx, "alice", "127.0.0.1"); err != nil {
		t.Fatalf("expected user to remain unlocked after first failure, got %v", err)
	}
	if err := locker.RecordFailure(ctx, "alice", "127.0.0.1"); err != nil {
		t.Fatalf("record second failure: %v", err)
	}
	if err := locker.Check(ctx, "alice", "127.0.0.1"); !errors.Is(err, ErrLocked) {
		t.Fatalf("expected ErrLocked, got %v", err)
	}
}

func TestResetClearsFailureCounters(t *testing.T) {
	locker, mr := setupLocker(t)
	defer mr.Close()
	ctx := context.Background()

	if err := locker.RecordFailure(ctx, "alice", "127.0.0.1"); err != nil {
		t.Fatalf("record failure: %v", err)
	}
	if err := locker.Reset(ctx, "alice", "127.0.0.1"); err != nil {
		t.Fatalf("reset: %v", err)
	}
	if mr.Exists(userFailKey("alice")) {
		t.Fatalf("expected user failure key to be deleted")
	}
	if mr.Exists(ipFailKey("127.0.0.1")) {
		t.Fatalf("expected ip failure key to be deleted")
	}
}
