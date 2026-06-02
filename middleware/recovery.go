package middleware

import (
	"log"
	"runtime/debug"

	"github.com/furadx/iam-go/pkg/code"
	"github.com/furadx/iam-go/pkg/response"
	"github.com/gin-gonic/gin"
)

// Recovery 捕获请求处理链中的 panic，记录堆栈并返回统一的 500 JSON，
// 而不是让连接被 net/http 直接重置。panic 详情只进日志，不返回客户端。
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("panic recovered: %v\n%s", r, debug.Stack())
				response.Fail(c, 500, code.ErrInternal)
				c.Abort()
			}
		}()
		c.Next()
	}
}
