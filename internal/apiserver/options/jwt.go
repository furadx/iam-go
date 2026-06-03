package options

import (
	"fmt"

	"github.com/spf13/pflag"
)

// insecureDefaultSecret 是开发用默认密钥，生产（release）模式禁止使用。
const insecureDefaultSecret = "change-me-in-production"

// minProductionSecretLen 是生产模式下 JWT 密钥的最小长度（字节）。
const minProductionSecretLen = 32

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

// ValidateForMode 在基础校验外，按运行模式做密钥强度校验：
// release（生产）模式禁止默认密钥并要求最小长度，避免用公开密钥签发可被伪造的令牌。
func (o *JWTOptions) ValidateForMode(mode string) []error {
	errs := o.Validate()
	if mode == "release" {
		if o.Secret == insecureDefaultSecret {
			errs = append(errs, fmt.Errorf("jwt secret must not use the insecure default in release mode"))
		} else if len(o.Secret) < minProductionSecretLen {
			errs = append(errs, fmt.Errorf("jwt secret must be at least %d bytes in release mode", minProductionSecretLen))
		}
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
