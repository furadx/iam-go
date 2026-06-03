package middleware

import (
	"log"
	"net/http"
	"runtime/debug"

	"github.com/furadx/iam-go/internal/pkg/code"
	"github.com/gin-gonic/gin"
)

// Recovery 捕获 panic 并返回 500 错误。
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("panic recovered: %v\n%s", r, debug.Stack())
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    code.ErrInternal,
					"message": code.Text(code.ErrInternal),
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}
