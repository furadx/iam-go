package code

import "errors"

// Common error codes.
const (
	// Success.
	OK = 0

	// Common errors.
	ErrBind       = 100001
	ErrValidation = 100002

	// Database errors.
	ErrDatabase = 100101

	// Authentication errors.
	ErrEncrypt         = 100201
	ErrSignToken       = 100202
	ErrTokenInvalid    = 100203
	ErrTokenExpired    = 100204
	ErrUnauthorized    = 100205

	// User errors.
	ErrUserNotFound     = 110001
	ErrUserAlreadyExist = 110002
	ErrPasswordIncorrect = 110003
)

var msgText = map[int]string{
	OK:                   "OK",
	ErrBind:              "参数绑定失败",
	ErrValidation:        "参数验证失败",
	ErrDatabase:          "数据库错误",
	ErrEncrypt:           "加密失败",
	ErrSignToken:         "签发 Token 失败",
	ErrTokenInvalid:      "Token 无效",
	ErrTokenExpired:      "Token 已过期",
	ErrUnauthorized:      "未授权",
	ErrUserNotFound:      "用户不存在",
	ErrUserAlreadyExist:  "用户已存在",
	ErrPasswordIncorrect: "密码错误",
}

// Text returns the text for the code.
func Text(code int) string {
	if msg, ok := msgText[code]; ok {
		return msg
	}
	return "未知错误"
}

// Error represents an error with a code and message.
type Error struct {
	Code int
	Msg  string
	Err  error
}

// Error returns the error message.
func (e *Error) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Msg
}

// Unwrap returns the wrapped error.
func (e *Error) Unwrap() error {
	return e.Err
}

// New creates a new error with the given code.
func New(code int) *Error {
	return &Error{
		Code: code,
		Msg:  Text(code),
	}
}

// WithCode wraps an error with a code.
func WithCode(code int, err error) *Error {
	return &Error{
		Code: code,
		Msg:  Text(code),
		Err:  err,
	}
}

// WithMessage creates an error with code and custom message.
func WithMessage(code int, message string) *Error {
	return &Error{
		Code: code,
		Msg:  message,
	}
}

// Code returns the code from an error.
func Code(err error) int {
	if err == nil {
		return OK
	}

	var e *Error
	if errors.As(err, &e) {
		return e.Code
	}

	return ErrDatabase
}

// Register registers a new error code with message.
func Register(code int, message string) {
	msgText[code] = message
}
