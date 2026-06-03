package options

import (
	"fmt"

	"github.com/spf13/pflag"
)

// JWTOptions JWT 配置选项。
type JWTOptions struct {
	Secret         string `json:"-" mapstructure:"secret"`
	AccessExpire   int    `json:"access_expire" mapstructure:"access_expire"`   // 秒
	RefreshExpire  int    `json:"refresh_expire" mapstructure:"refresh_expire"` // 秒
	RevokeFailOpen bool   `json:"revoke_fail_open" mapstructure:"revoke_fail_open"`
}

// NewJWTOptions 创建默认的 JWT 选项。
func NewJWTOptions() *JWTOptions {
	return &JWTOptions{
		Secret:         "change-me-in-production",
		AccessExpire:   900,    // 15 分钟
		RefreshExpire:  604800, // 7 天
		RevokeFailOpen: true,   // Redis 故障时放行（保可用性）
	}
}

// Validate 验证 JWT 选项。
func (o *JWTOptions) Validate() []error {
	var errs []error
	if o.Secret == "" {
		errs = append(errs, fmt.Errorf("jwt secret cannot be empty"))
	}
	if o.AccessExpire <= 0 {
		errs = append(errs, fmt.Errorf("jwt access_expire must be positive seconds"))
	}
	if o.RefreshExpire <= 0 {
		errs = append(errs, fmt.Errorf("jwt refresh_expire must be positive seconds"))
	}
	if o.RefreshExpire < o.AccessExpire {
		errs = append(errs, fmt.Errorf("jwt refresh_expire should be >= access_expire"))
	}
	return errs
}

// AddFlags 添加命令行标志。
func (o *JWTOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.Secret, "jwt.secret", o.Secret, "JWT signing secret")
	fs.IntVar(&o.AccessExpire, "jwt.access-expire", o.AccessExpire, "Access token expiration in seconds")
	fs.IntVar(&o.RefreshExpire, "jwt.refresh-expire", o.RefreshExpire, "Refresh token expiration in seconds")
	fs.BoolVar(&o.RevokeFailOpen, "jwt.revoke-fail-open", o.RevokeFailOpen, "Allow requests when revoke store (redis) is unavailable")
}
