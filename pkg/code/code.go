package code

import "errors"

const (
	OK = 0

	ErrInvalidParam = 100001
	ErrUnauthorized = 100002
	ErrInternal     = 100003
	ErrTooManyReq   = 100004

	ErrInvalidToken = 300001
	ErrExpiredToken = 300002
)

var messages = map[int]string{
	OK:              "ok",
	ErrInvalidParam: "参数错误",
	ErrUnauthorized: "未认证或登录已过期",
	ErrInternal:     "服务器内部错误",
	ErrTooManyReq:   "请求过于频繁",
	ErrInvalidToken: "无效的访问令牌",
	ErrExpiredToken: "访问令牌已过期",
}

func Message(c int) string {
	msg, ok := messages[c]
	if !ok {
		return "未知错误"
	}
	return msg
}

type Error struct {
	Code int
	Err  error
}

func New(c int) error {
	return &Error{
		Code: c,
		Err:  errors.New(Message(c)),
	}
}

func Wrap(c int, err error) error {
	if err == nil {
		return New(c)
	}
	return &Error{
		Code: c,
		Err:  err,
	}
}

func (e *Error) Error() string {
	if e == nil || e.Err == nil {
		return Message(ErrInternal)
	}
	return e.Err.Error()
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func FromError(err error) int {
	if err == nil {
		return OK
	}

	var coded *Error
	if errors.As(err, &coded) && coded.Code != OK {
		return coded.Code
	}

	return ErrInternal
}

// RegisterMessage 允许应用程序注册自定义错误码消息
func RegisterMessage(code int, message string) {
	messages[code] = message
}
