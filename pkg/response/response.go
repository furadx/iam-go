package response

import (
	"github.com/furadx/iam-go/pkg/code"
	"github.com/gin-gonic/gin"
)

type Body struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type ErrorBody struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func OK(c *gin.Context, data interface{}) {
	c.JSON(200, Body{
		Code:    0,
		Message: "OK",
		Data:    data,
	})
}

func Fail(c *gin.Context, httpStatus int, bizCode int) {
	c.JSON(httpStatus, ErrorBody{
		Code:    bizCode,
		Message: code.Message(bizCode),
	})
}

func FailWithMessage(c *gin.Context, httpStatus int, bizCode int, message string) {
	c.JSON(httpStatus, ErrorBody{
		Code:    bizCode,
		Message: message,
	})
}

func FailError(c *gin.Context, httpStatus int, err error) {
	bizCode := code.FromError(err)
	FailWithMessage(c, httpStatus, bizCode, code.Message(bizCode))
}
