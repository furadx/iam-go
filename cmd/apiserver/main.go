package main

import (
	"context"
	"flag"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/furadx/iam-go/internal/apiserver"
	"github.com/furadx/iam-go/internal/apiserver/config"
	"github.com/furadx/iam-go/internal/apiserver/store/postgres"
)

func main() {
	// 解析命令行参数
	var configFile string
	flag.StringVar(&configFile, "config", "", "配置文件路径")
	flag.Parse()

	// 加载配置
	cfg, err := loadConfig(configFile)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 设置 Gin 模式
	gin.SetMode(cfg.Server.Mode)

	// 初始化数据库
	store, err := postgres.GetPostgresFactoryOr(cfg.Database.DSN())
	if err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	defer store.Close()

	// 初始化路由
	router := apiserver.InitRouter(store)

	// 创建 HTTP 服务器
	srv := apiserver.NewServer(cfg.Server.Addr, router)

	// 优雅关闭
	run(srv, store)
}

// loadConfig 加载配置（优先从配置文件，其次从环境变量）。
func loadConfig(configFile string) (*config.Config, error) {
	if configFile != "" {
		return config.Load(configFile)
	}
	return config.LoadFromEnv()
}

// run 运行服务器并处理优雅关闭。
func run(srv *apiserver.Server, store interface{ Close() error }) {
	// 捕获信号
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// 启动服务器
	go func() {
		if err := srv.Run(); err != nil {
			log.Fatalf("服务器启动失败: %v", err)
		}
	}()

	// 等待信号
	<-ctx.Done()
	stop()

	// 优雅关闭
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("服务器关闭失败: %v", err)
	}

	// 关闭数据库连接
	if err := store.Close(); err != nil {
		log.Printf("关闭数据库连接失败: %v", err)
	}
}
