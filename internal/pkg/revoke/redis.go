package revoke

import (
	"context"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/furadx/iam-go/pkg/token"
)

type redisRevoker struct {
	client *redis.Client
}

// NewRedisRevoker 用 Redis 实现 Revoker。
func NewRedisRevoker(client *redis.Client) Revoker {
	return &redisRevoker{client: client}
}

func blacklistKey(jti string) string { return "jwt:bl:" + jti }

func revokeBeforeKey(uid int64) string {
	return "jwt:revoke_before:" + strconv.FormatInt(uid, 10)
}

func (r *redisRevoker) Revoke(ctx context.Context, jti string, ttl time.Duration) error {
	if jti == "" || ttl <= 0 {
		return nil
	}
	return r.client.Set(ctx, blacklistKey(jti), "1", ttl).Err()
}

func (r *redisRevoker) RevokeAllBefore(ctx context.Context, uid int64, t time.Time) error {
	return r.client.Set(ctx, revokeBeforeKey(uid), strconv.FormatInt(t.Unix(), 10), 0).Err()
}

func (r *redisRevoker) Allowed(ctx context.Context, c *token.Claims) (bool, error) {
	exists, err := r.client.Exists(ctx, blacklistKey(c.ID)).Result()
	if err != nil {
		return false, err
	}
	if exists > 0 {
		return false, nil
	}

	val, err := r.client.Get(ctx, revokeBeforeKey(c.UserID)).Result()
	if err == redis.Nil {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	rb, perr := strconv.ParseInt(val, 10, 64)
	if perr != nil {
		return true, nil // 脏数据，放行
	}
	if c.IssuedAt != nil && c.IssuedAt.Time.Unix() < rb {
		return false, nil
	}
	return true, nil
}
