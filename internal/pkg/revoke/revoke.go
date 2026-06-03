package revoke

import (
	"context"
	"time"

	"github.com/furadx/iam-go/pkg/token"
)

// Revoker 定义令牌吊销与校验。
type Revoker interface {
	// Revoke 把单个令牌的 jti 拉黑，ttl 为该令牌剩余寿命。
	Revoke(ctx context.Context, jti string, ttl time.Duration) error
	// RevokeAllBefore 使某用户 iat 早于 t 的全部令牌失效。
	RevokeAllBefore(ctx context.Context, uid int64, t time.Time) error
	// Allowed 校验 claims 是否未被吊销（jti 未拉黑 且 iat >= revokeBefore）。
	Allowed(ctx context.Context, c *token.Claims) (bool, error)
}
