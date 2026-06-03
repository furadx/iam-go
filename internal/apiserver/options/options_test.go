package options

import (
	"strings"
	"testing"
)

// strongSecret 是一个满足生产强度的示例密钥（>=32 字节且非默认值）。
const strongSecret = "a-sufficiently-long-production-secret-1234567890"

func hasJWTSecretError(errs []error) bool {
	for _, err := range errs {
		if strings.Contains(err.Error(), "jwt secret") {
			return true
		}
	}
	return false
}

func TestValidate_ReleaseModeRejectsDefaultSecret(t *testing.T) {
	o := NewOptions()
	o.Server.Mode = "release"
	// 保持默认 secret = "change-me-in-production"
	errs := o.Validate()
	if !hasJWTSecretError(errs) {
		t.Fatalf("release 模式下默认密钥应被拒绝，但 Validate 未报 jwt secret 错误: %v", errs)
	}
}

func TestValidate_ReleaseModeRejectsShortSecret(t *testing.T) {
	o := NewOptions()
	o.Server.Mode = "release"
	o.JWT.Secret = "short"
	errs := o.Validate()
	if !hasJWTSecretError(errs) {
		t.Fatalf("release 模式下过短密钥应被拒绝，但 Validate 未报 jwt secret 错误: %v", errs)
	}
}

func TestValidate_ReleaseModeAcceptsStrongSecret(t *testing.T) {
	o := NewOptions()
	o.Server.Mode = "release"
	o.JWT.Secret = strongSecret
	errs := o.Validate()
	if hasJWTSecretError(errs) {
		t.Fatalf("release 模式下强密钥不应报错: %v", errs)
	}
}

func TestValidate_DebugModeAllowsDefaultSecret(t *testing.T) {
	o := NewOptions()
	o.Server.Mode = "debug"
	// 默认 secret 在开发模式应被放行
	errs := o.Validate()
	if hasJWTSecretError(errs) {
		t.Fatalf("debug 模式下默认密钥应被放行，但报了 jwt secret 错误: %v", errs)
	}
}
