package password

import (
	"errors"
	"strings"
	"unicode"
)

var (
	ErrTooShort         = errors.New("password too short")
	ErrTooLong          = errors.New("password too long")
	ErrTooFewClasses    = errors.New("password has too few character classes")
	ErrContainsUsername = errors.New("password contains username")
	ErrCommonPassword   = errors.New("password is too common")
)

// Policy 定义密码强度策略。
type Policy struct {
	MinLength             int
	MaxLength             int
	MinClasses            int
	RejectUsername        bool
	RejectCommonPasswords bool
}

// DefaultPolicy 返回默认企业密码策略。
func DefaultPolicy() Policy {
	return Policy{
		MinLength:             12,
		MaxLength:             64,
		MinClasses:            3,
		RejectUsername:        true,
		RejectCommonPasswords: true,
	}
}

// Validate 校验密码是否满足策略。
func (p Policy) Validate(username, value string) error {
	if p.MinLength <= 0 {
		p.MinLength = 12
	}
	if p.MaxLength <= 0 {
		p.MaxLength = 64
	}
	if p.MinClasses < 0 {
		p.MinClasses = 0
	}
	if p.MinClasses > 4 {
		p.MinClasses = 4
	}

	if len(value) < p.MinLength {
		return ErrTooShort
	}
	if len(value) > p.MaxLength {
		return ErrTooLong
	}
	if p.RejectUsername && containsUsername(username, value) {
		return ErrContainsUsername
	}
	if p.RejectCommonPasswords && commonPassword(value) {
		return ErrCommonPassword
	}
	if classes(value) < p.MinClasses {
		return ErrTooFewClasses
	}
	return nil
}

func containsUsername(username, value string) bool {
	username = strings.TrimSpace(strings.ToLower(username))
	if len(username) < 3 {
		return false
	}
	return strings.Contains(strings.ToLower(value), username)
}

func classes(value string) int {
	var lower, upper, digit, special bool
	for _, r := range value {
		switch {
		case unicode.IsLower(r):
			lower = true
		case unicode.IsUpper(r):
			upper = true
		case unicode.IsDigit(r):
			digit = true
		default:
			special = true
		}
	}
	total := 0
	for _, ok := range []bool{lower, upper, digit, special} {
		if ok {
			total++
		}
	}
	return total
}

func commonPassword(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "password", "password123", "password1234", "12345678", "123456789",
		"qwerty123", "admin123", "admin123456", "changeme", "letmein",
		"welcome123", "iamgo123":
		return true
	default:
		return false
	}
}
