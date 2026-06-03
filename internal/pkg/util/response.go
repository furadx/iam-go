package util

import (
	"github.com/furadx/iam-go/internal/pkg/code"
	"github.com/gin-gonic/gin"
)

const (
	// XRequestIDKey 是 Request-ID 的键名。
	XRequestIDKey = "X-Request-ID"
)

// Response 标准响应格式。
type Response struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
}

// WriteResponse 写入响应到 gin.Context。
func WriteResponse(c *gin.Context, err error, data interface{}) {
	requestID, _ := c.Get(XRequestIDKey)
	rid, _ := requestID.(string)

	if err != nil {
		bizCode := code.Code(err)
		c.JSON(200, Response{
			Code:      bizCode,
			Message:   code.Text(bizCode),
			RequestID: rid,
		})
		return
	}

	c.JSON(200, Response{
		Code:      code.OK,
		Message:   code.Text(code.OK),
		Data:      data,
		RequestID: rid,
	})
}
