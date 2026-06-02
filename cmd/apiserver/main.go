package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/furadx/iam-go/internal/apiserver/controller/v1/user"
	"github.com/furadx/iam-go/internal/apiserver/store/postgres"
	"github.com/furadx/iam-go/internal/pkg/middleware"
)

func main() {
	// 数据库连接字符串（实际使用时应该从配置文件读取）
	dsn := "host=localhost user=postgres password=postgres dbname=iam port=5432 sslmode=disable"

	// 初始化数据库
	store, err := postgres.GetPostgresFactoryOr(dsn)
	if err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	defer store.Close()

	// 创建 Gin 引擎
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	// 全局中间件
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS())

	// 健康检查
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 初始化控制器
	userController := user.NewUserController(store)

	// API 路由
	v1 := r.Group("/api/v1")
	{
		// 用户相关接口
		v1.POST("/users", userController.Create)
		v1.GET("/users", userController.List)
		v1.GET("/users/:name", userController.Get)
	}

	// 创建 HTTP 服务器
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	// 优雅关闭
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Println("服务启动在 :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("服务启动失败: %v", err)
		}
	}()

	<-ctx.Done()
	stop()
	log.Println("收到关闭信号，开始优雅关闭...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("服务关闭失败: %v", err)
	}

	log.Println("服务已关闭")
}
