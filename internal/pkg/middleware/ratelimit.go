package middleware

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/furadx/iam-go/internal/pkg/code"
	"github.com/furadx/iam-go/internal/pkg/ratelimit"
	"github.com/furadx/iam-go/pkg/log"
)

// RateLimiter 是限流器接口。
type RateLimiter interface {
	Allow(ctx context.Context, key string, limit int, window time.Duration) (ratelimit.Result, error)
}

// RateLimitConfig 是限流中间件配置。
type RateLimitConfig struct {
	Enabled  bool
	Name     string
	Limit    int
	Window   time.Duration
	FailOpen bool
}

// RateLimit 基于客户端 IP 做固定窗口限流。
func RateLimit(limiter RateLimiter, cfg RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !cfg.Enabled || limiter == nil || cfg.Limit <= 0 || cfg.Window <= 0 {
			c.Next()
			return
		}
		name := cfg.Name
		if name == "" {
			name = "api"
		}
		key := name + ":ip:" + c.ClientIP()
		result, err := limiter.Allow(c.Request.Context(), key, cfg.Limit, cfg.Window)
		if err != nil {
			if cfg.FailOpen {
				log.Warnf("rate limit store unavailable, fail-open allows request: %v", err)
				c.Next()
				return
			}
			abortRateLimited(c, cfg, ratelimit.Result{})
			return
		}
		setRateLimitHeaders(c, cfg, result)
		if !result.Allowed {
			abortRateLimited(c, cfg, result)
			return
		}
		c.Next()
	}
}

func setRateLimitHeaders(c *gin.Context, cfg RateLimitConfig, result ratelimit.Result) {
	c.Header("X-RateLimit-Limit", strconv.Itoa(cfg.Limit))
	c.Header("X-RateLimit-Remaining", strconv.Itoa(result.Remaining))
	if !result.Reset.IsZero() {
		c.Header("X-RateLimit-Reset", strconv.FormatInt(result.Reset.Unix(), 10))
	}
}

func abortRateLimited(c *gin.Context, cfg RateLimitConfig, result ratelimit.Result) {
	setRateLimitHeaders(c, cfg, result)
	c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
		"code":    code.ErrTooManyRequests,
		"message": code.Text(code.ErrTooManyRequests),
	})
}
