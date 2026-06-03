package options

import (
	"fmt"

	"github.com/spf13/pflag"
)

// JWTOptions JWT 配置选项。
type JWTOptions struct {
	Secret string `json:"-" mapstructure:"secret"`      // 不输出到 JSON
	Expire int    `json:"expire" mapstructure:"expire"` // Token 有效期（秒）
}

// NewJWTOptions 创建默认的 JWT 选项。
func NewJWTOptions() *JWTOptions {
	return &JWTOptions{
		Secret: "change-me-in-production",
		Expire: 86400, // 24 小时
	}
}

// Validate 验证 JWT 选项。
func (o *JWTOptions) Validate() []error {
	var errs []error

	if o.Secret == "" {
		errs = append(errs, fmt.Errorf("jwt secret cannot be empty"))
	}

	if o.Expire <= 0 {
		errs = append(errs, fmt.Errorf("jwt expire must be a positive number of seconds"))
	}

	return errs
}

// AddFlags 添加命令行标志。
func (o *JWTOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.Secret, "jwt.secret", o.Secret, "JWT signing secret")
	fs.IntVar(&o.Expire, "jwt.expire", o.Expire, "JWT token expiration in seconds")
}
