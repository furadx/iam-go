package main

import (
	"log"
	"net/http"

	"github.com/furadx/iam-go/middleware"
	"github.com/furadx/iam-go/pkg/jwt"
	"github.com/furadx/iam-go/pkg/response"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.New()

	// 使用中间件
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS())

	// 健康检查
	r.GET("/healthz", func(c *gin.Context) {
		response.OK(c, gin.H{"status": "ok"})
	})

	// JWT 管理器
	jwtManager := jwt.NewManager("your-secret-key", "example-app", 3600)

	// 登录接口（示例）
	r.POST("/login", func(c *gin.Context) {
		// 这里应该验证用户凭证
		userID := int64(12345)

		token, ttl, err := jwtManager.Sign(userID)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, 100003)
			return
		}

		response.OK(c, gin.H{
			"token":      token,
			"expires_in": ttl,
		})
	})

	// 受保护的接口
	r.GET("/protected", middleware.Auth(jwtManager), func(c *gin.Context) {
		userID, _ := middleware.GetUserID(c)
		response.OK(c, gin.H{
			"message": "这是受保护的资源",
			"user_id": userID,
		})
	})

	log.Println("服务启动在 :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
