package util

import (
	"github.com/furadx/iam-go/internal/pkg/code"
	"github.com/gin-gonic/gin"
)

// Response 标准响应格式。
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// WriteResponse 写入响应到 gin.Context。
func WriteResponse(c *gin.Context, err error, data interface{}) {
	if err != nil {
		bizCode := code.Code(err)
		c.JSON(200, Response{
			Code:    bizCode,
			Message: code.Text(bizCode),
		})
		return
	}

	c.JSON(200, Response{
		Code:    code.OK,
		Message: code.Text(code.OK),
		Data:    data,
	})
}
