package options

import (
	"fmt"

	"github.com/spf13/pflag"
)

// SecurityOptions contains security-related configuration.
type SecurityOptions struct {
	RateLimit      *RateLimitOptions      `json:"rate_limit" mapstructure:"rate_limit"`
	LoginLock      *LoginLockOptions      `json:"login_lock" mapstructure:"login_lock"`
	PasswordPolicy *PasswordPolicyOptions `json:"password_policy" mapstructure:"password_policy"`
	CORS           *CORSOptions           `json:"cors" mapstructure:"cors"`
}

// RateLimitOptions configures Redis-backed rate limiting.
type RateLimitOptions struct {
	Enabled            bool `json:"enabled" mapstructure:"enabled"`
	APILimit           int  `json:"api_limit" mapstructure:"api_limit"`
	APIWindowSeconds   int  `json:"api_window_seconds" mapstructure:"api_window_seconds"`
	LoginIPLimit       int  `json:"login_ip_limit" mapstructure:"login_ip_limit"`
	LoginWindowSeconds int  `json:"login_window_seconds" mapstructure:"login_window_seconds"`
	FailOpen           bool `json:"fail_open" mapstructure:"fail_open"`
}

// LoginLockOptions configures failed-login lockout.
type LoginLockOptions struct {
	Enabled          bool `json:"enabled" mapstructure:"enabled"`
	UserMaxFailures  int  `json:"user_max_failures" mapstructure:"user_max_failures"`
	IPMaxFailures    int  `json:"ip_max_failures" mapstructure:"ip_max_failures"`
	FailureWindowMin int  `json:"failure_window_minutes" mapstructure:"failure_window_minutes"`
	LockMinutes      int  `json:"lock_minutes" mapstructure:"lock_minutes"`
}

// PasswordPolicyOptions configures password strength checks.
type PasswordPolicyOptions struct {
	MinLength             int  `json:"min_length" mapstructure:"min_length"`
	MaxLength             int  `json:"max_length" mapstructure:"max_length"`
	MinClasses            int  `json:"min_classes" mapstructure:"min_classes"`
	RejectUsername        bool `json:"reject_username" mapstructure:"reject_username"`
	RejectCommonPasswords bool `json:"reject_common_passwords" mapstructure:"reject_common_passwords"`
}

// CORSOptions configures browser cross-origin access.
type CORSOptions struct {
	AllowedOrigins   []string `json:"allowed_origins" mapstructure:"allowed_origins"`
	AllowCredentials bool     `json:"allow_credentials" mapstructure:"allow_credentials"`
	MaxAgeSeconds    int      `json:"max_age_seconds" mapstructure:"max_age_seconds"`
}

// NewSecurityOptions creates default security configuration.
func NewSecurityOptions() *SecurityOptions {
	return &SecurityOptions{
		RateLimit: &RateLimitOptions{
			Enabled:            true,
			APILimit:           300,
			APIWindowSeconds:   60,
			LoginIPLimit:       10,
			LoginWindowSeconds: 60,
			FailOpen:           true,
		},
		LoginLock: &LoginLockOptions{
			Enabled:          true,
			UserMaxFailures:  5,
			IPMaxFailures:    20,
			FailureWindowMin: 15,
			LockMinutes:      15,
		},
		PasswordPolicy: &PasswordPolicyOptions{
			MinLength:             12,
			MaxLength:             64,
			MinClasses:            3,
			RejectUsername:        true,
			RejectCommonPasswords: true,
		},
		CORS: &CORSOptions{
			AllowedOrigins: []string{
				"http://localhost:3000",
				"http://localhost:5173",
			},
			AllowCredentials: true,
			MaxAgeSeconds:    43200,
		},
	}
}

// Validate validates security configuration.
func (o *SecurityOptions) Validate() []error {
	var errs []error
	if o.RateLimit == nil {
		errs = append(errs, fmt.Errorf("security rate_limit cannot be nil"))
	} else {
		errs = append(errs, o.RateLimit.Validate()...)
	}
	if o.LoginLock == nil {
		errs = append(errs, fmt.Errorf("security login_lock cannot be nil"))
	} else {
		errs = append(errs, o.LoginLock.Validate()...)
	}
	if o.PasswordPolicy == nil {
		errs = append(errs, fmt.Errorf("security password_policy cannot be nil"))
	} else {
		errs = append(errs, o.PasswordPolicy.Validate()...)
	}
	if o.CORS == nil {
		errs = append(errs, fmt.Errorf("security cors cannot be nil"))
	} else {
		errs = append(errs, o.CORS.Validate()...)
	}
	return errs
}

// Validate validates rate limit configuration.
func (o *RateLimitOptions) Validate() []error {
	var errs []error
	if o.APILimit <= 0 {
		errs = append(errs, fmt.Errorf("security rate_limit api_limit must be positive"))
	}
	if o.APIWindowSeconds <= 0 {
		errs = append(errs, fmt.Errorf("security rate_limit api_window_seconds must be positive"))
	}
	if o.LoginIPLimit <= 0 {
		errs = append(errs, fmt.Errorf("security rate_limit login_ip_limit must be positive"))
	}
	if o.LoginWindowSeconds <= 0 {
		errs = append(errs, fmt.Errorf("security rate_limit login_window_seconds must be positive"))
	}
	return errs
}

// Validate validates login lock configuration.
func (o *LoginLockOptions) Validate() []error {
	var errs []error
	if o.UserMaxFailures <= 0 {
		errs = append(errs, fmt.Errorf("security login_lock user_max_failures must be positive"))
	}
	if o.IPMaxFailures <= 0 {
		errs = append(errs, fmt.Errorf("security login_lock ip_max_failures must be positive"))
	}
	if o.FailureWindowMin <= 0 {
		errs = append(errs, fmt.Errorf("security login_lock failure_window_minutes must be positive"))
	}
	if o.LockMinutes <= 0 {
		errs = append(errs, fmt.Errorf("security login_lock lock_minutes must be positive"))
	}
	return errs
}

// Validate validates password policy configuration.
func (o *PasswordPolicyOptions) Validate() []error {
	var errs []error
	if o.MinLength <= 0 {
		errs = append(errs, fmt.Errorf("security password_policy min_length must be positive"))
	}
	if o.MaxLength < o.MinLength {
		errs = append(errs, fmt.Errorf("security password_policy max_length must be >= min_length"))
	}
	if o.MinClasses < 0 || o.MinClasses > 4 {
		errs = append(errs, fmt.Errorf("security password_policy min_classes must be between 0 and 4"))
	}
	return errs
}

// Validate validates CORS configuration.
func (o *CORSOptions) Validate() []error {
	var errs []error
	if o.MaxAgeSeconds < 0 {
		errs = append(errs, fmt.Errorf("security cors max_age_seconds cannot be negative"))
	}
	return errs
}

// AddFlags adds security configuration flags.
func (o *SecurityOptions) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVar(&o.RateLimit.Enabled, "security.rate-limit.enabled", o.RateLimit.Enabled, "Enable Redis-backed rate limiting")
	fs.IntVar(&o.RateLimit.APILimit, "security.rate-limit.api-limit", o.RateLimit.APILimit, "API requests allowed per client IP per window")
	fs.IntVar(&o.RateLimit.APIWindowSeconds, "security.rate-limit.api-window-seconds", o.RateLimit.APIWindowSeconds, "API rate limit window in seconds")
	fs.IntVar(&o.RateLimit.LoginIPLimit, "security.rate-limit.login-ip-limit", o.RateLimit.LoginIPLimit, "Login requests allowed per client IP per window")
	fs.IntVar(&o.RateLimit.LoginWindowSeconds, "security.rate-limit.login-window-seconds", o.RateLimit.LoginWindowSeconds, "Login rate limit window in seconds")
	fs.BoolVar(&o.RateLimit.FailOpen, "security.rate-limit.fail-open", o.RateLimit.FailOpen, "Allow requests when rate limit store is unavailable")

	fs.BoolVar(&o.LoginLock.Enabled, "security.login-lock.enabled", o.LoginLock.Enabled, "Enable failed-login lockout")
	fs.IntVar(&o.LoginLock.UserMaxFailures, "security.login-lock.user-max-failures", o.LoginLock.UserMaxFailures, "Failed login attempts per username before lock")
	fs.IntVar(&o.LoginLock.IPMaxFailures, "security.login-lock.ip-max-failures", o.LoginLock.IPMaxFailures, "Failed login attempts per IP before lock")
	fs.IntVar(&o.LoginLock.FailureWindowMin, "security.login-lock.failure-window-minutes", o.LoginLock.FailureWindowMin, "Failed login counter window in minutes")
	fs.IntVar(&o.LoginLock.LockMinutes, "security.login-lock.lock-minutes", o.LoginLock.LockMinutes, "Login lock duration in minutes")

	fs.IntVar(&o.PasswordPolicy.MinLength, "security.password-policy.min-length", o.PasswordPolicy.MinLength, "Minimum password length")
	fs.IntVar(&o.PasswordPolicy.MaxLength, "security.password-policy.max-length", o.PasswordPolicy.MaxLength, "Maximum password length")
	fs.IntVar(&o.PasswordPolicy.MinClasses, "security.password-policy.min-classes", o.PasswordPolicy.MinClasses, "Minimum character classes required")
	fs.BoolVar(&o.PasswordPolicy.RejectUsername, "security.password-policy.reject-username", o.PasswordPolicy.RejectUsername, "Reject passwords containing username")
	fs.BoolVar(&o.PasswordPolicy.RejectCommonPasswords, "security.password-policy.reject-common-passwords", o.PasswordPolicy.RejectCommonPasswords, "Reject common weak passwords")

	fs.StringSliceVar(&o.CORS.AllowedOrigins, "security.cors.allowed-origins", o.CORS.AllowedOrigins, "Allowed browser origins for CORS")
	fs.BoolVar(&o.CORS.AllowCredentials, "security.cors.allow-credentials", o.CORS.AllowCredentials, "Allow browser credentials for CORS")
	fs.IntVar(&o.CORS.MaxAgeSeconds, "security.cors.max-age-seconds", o.CORS.MaxAgeSeconds, "CORS preflight max age in seconds")
}
