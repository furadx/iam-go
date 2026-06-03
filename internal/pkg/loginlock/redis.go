package loginlock

import (
	"context"
	"errors"
	"net"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrLocked = errors.New("login locked")

// Config 定义登录失败锁定策略。
type Config struct {
	Enabled         bool
	UserMaxFailures int
	IPMaxFailures   int
	FailureWindow   time.Duration
	LockDuration    time.Duration
}

// Locker 维护登录失败计数与临时锁定。
type Locker struct {
	client *redis.Client
	cfg    Config
}

// NewRedisLocker 创建 Redis 登录锁定器。
func NewRedisLocker(client *redis.Client, cfg Config) *Locker {
	if cfg.UserMaxFailures <= 0 {
		cfg.UserMaxFailures = 5
	}
	if cfg.IPMaxFailures <= 0 {
		cfg.IPMaxFailures = 20
	}
	if cfg.FailureWindow <= 0 {
		cfg.FailureWindow = 15 * time.Minute
	}
	if cfg.LockDuration <= 0 {
		cfg.LockDuration = 15 * time.Minute
	}
	return &Locker{client: client, cfg: cfg}
}

// Check 检查用户名或 IP 是否处于锁定状态。
func (l *Locker) Check(ctx context.Context, username, ip string) error {
	if l == nil || !l.cfg.Enabled {
		return nil
	}
	locked, err := l.locked(ctx, userLockKey(username))
	if err != nil || locked {
		if locked {
			return ErrLocked
		}
		return err
	}
	locked, err = l.locked(ctx, ipLockKey(ip))
	if err != nil || locked {
		if locked {
			return ErrLocked
		}
		return err
	}
	return nil
}

// RecordFailure 记录一次失败，并在超过阈值时写入锁定 key。
func (l *Locker) RecordFailure(ctx context.Context, username, ip string) error {
	if l == nil || !l.cfg.Enabled {
		return nil
	}
	if err := l.record(ctx, userFailKey(username), userLockKey(username), l.cfg.UserMaxFailures); err != nil {
		return err
	}
	return l.record(ctx, ipFailKey(ip), ipLockKey(ip), l.cfg.IPMaxFailures)
}

// Reset 清理登录成功后的失败计数。
func (l *Locker) Reset(ctx context.Context, username, ip string) error {
	if l == nil || !l.cfg.Enabled {
		return nil
	}
	return l.client.Del(ctx, userFailKey(username), ipFailKey(ip)).Err()
}

func (l *Locker) locked(ctx context.Context, key string) (bool, error) {
	exists, err := l.client.Exists(ctx, key).Result()
	return exists > 0, err
}

func (l *Locker) record(ctx context.Context, failKey, lockKey string, maxFailures int) error {
	count, err := l.client.Incr(ctx, failKey).Result()
	if err != nil {
		return err
	}
	if count == 1 {
		if err := l.client.Expire(ctx, failKey, l.cfg.FailureWindow).Err(); err != nil {
			return err
		}
	}
	if int(count) >= maxFailures {
		if err := l.client.Set(ctx, lockKey, "1", l.cfg.LockDuration).Err(); err != nil {
			return err
		}
	}
	return nil
}

func userFailKey(username string) string {
	return "auth:login:fail:user:" + normalizeUsername(username)
}
func userLockKey(username string) string {
	return "auth:login:lock:user:" + normalizeUsername(username)
}
func ipFailKey(ip string) string { return "auth:login:fail:ip:" + normalizeIP(ip) }
func ipLockKey(ip string) string { return "auth:login:lock:ip:" + normalizeIP(ip) }

func normalizeUsername(username string) string {
	username = strings.TrimSpace(strings.ToLower(username))
	if username == "" {
		return "-"
	}
	return username
}

func normalizeIP(ip string) string {
	parsed := net.ParseIP(strings.TrimSpace(ip))
	if parsed == nil {
		return "-"
	}
	return parsed.String()
}
