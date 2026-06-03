package ratelimit

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// Result 是一次限流判定结果。
type Result struct {
	Allowed   bool
	Remaining int
	Reset     time.Time
}

// RedisLimiter 基于 Redis 固定窗口计数。
type RedisLimiter struct {
	client *redis.Client
	prefix string
	now    func() time.Time
}

// NewRedisLimiter 创建 Redis 限流器。
func NewRedisLimiter(client *redis.Client) *RedisLimiter {
	return &RedisLimiter{
		client: client,
		prefix: "rl",
		now:    time.Now,
	}
}

// Allow 判断 key 在指定窗口内是否还可访问。
func (l *RedisLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (Result, error) {
	if limit <= 0 || window <= 0 {
		return Result{Allowed: true, Remaining: limit}, nil
	}
	now := l.now()
	windowSeconds := int64(window.Seconds())
	if windowSeconds <= 0 {
		windowSeconds = 1
	}
	windowID := now.Unix() / windowSeconds
	reset := time.Unix((windowID+1)*windowSeconds, 0)
	redisKey := fmt.Sprintf("%s:%s:%d", l.prefix, sanitizeKey(key), windowID)

	count, err := l.client.Incr(ctx, redisKey).Result()
	if err != nil {
		return Result{}, err
	}
	if count == 1 {
		if err := l.client.Expire(ctx, redisKey, window).Err(); err != nil {
			return Result{}, err
		}
	}

	remaining := limit - int(count)
	if remaining < 0 {
		remaining = 0
	}
	return Result{
		Allowed:   int(count) <= limit,
		Remaining: remaining,
		Reset:     reset,
	}, nil
}

func sanitizeKey(key string) string {
	key = strings.TrimSpace(key)
	if key == "" {
		return "-"
	}
	replacer := strings.NewReplacer(" ", "_", ":", "_", "/", "_", "\\", "_")
	return replacer.Replace(key)
}
