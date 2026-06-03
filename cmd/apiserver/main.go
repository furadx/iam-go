package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"gorm.io/gorm"

	"github.com/furadx/iam-go/internal/apiserver"
	"github.com/furadx/iam-go/internal/apiserver/options"
	"github.com/furadx/iam-go/internal/apiserver/store/postgres"
	pkglog "github.com/furadx/iam-go/pkg/log"
)

var (
	configFile string
	showVersion bool
	Version    = "v0.3.0"
	BuildDate  = "unknown"
	GitCommit  = "unknown"
)

func main() {
	// 创建默认选项
	opts := options.NewOptions()

	// 添加通用标志
	pflag.StringVarP(&configFile, "config", "c", "", "配置文件路径")
	pflag.BoolVarP(&showVersion, "version", "v", false, "显示版本信息")

	// 添加 Options 标志
	opts.AddFlags(pflag.CommandLine)

	// 解析命令行参数
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()

	// 显示版本信息
	if showVersion {
		log.Printf("Version: %s\n", Version)
		log.Printf("Build Date: %s\n", BuildDate)
		log.Printf("Git Commit: %s\n", GitCommit)
		os.Exit(0)
	}

	// 加载配置文件
	if configFile != "" {
		viper.SetConfigFile(configFile)
		if err := viper.ReadInConfig(); err != nil {
			log.Fatalf("读取配置文件失败: %v", err)
		}

		// 将配置文件内容绑定到 Options
		if err := viper.Unmarshal(opts); err != nil {
			log.Fatalf("解析配置失败: %v", err)
		}
	}

	// 验证配置
	if errs := opts.Validate(); len(errs) > 0 {
		for _, err := range errs {
			log.Printf("配置错误: %v", err)
		}
		log.Fatal("配置验证失败")
	}

	// 完成配置（填充默认值和派生字段）
	completedOpts := opts.Complete()

	// 初始化日志系统
	logger := pkglog.New(&pkglog.Options{
		Level:             completedOpts.Log.Level,
		Format:            completedOpts.Log.Format,
		EnableColor:       completedOpts.Log.EnableColor,
		DisableCaller:     completedOpts.Log.DisableCaller,
		DisableStacktrace: completedOpts.Log.DisableStacktrace,
		OutputPaths:       completedOpts.Log.OutputPaths,
		ErrorOutputPaths:  completedOpts.Log.ErrorOutputPaths,
	})
	pkglog.SetLogger(logger)
	defer pkglog.Sync()

	pkglog.Infof("Starting apiserver %s (commit: %s, built: %s)", Version, GitCommit, BuildDate)

	// 设置 Gin 模式
	gin.SetMode(completedOpts.Server.Mode)

	// 初始化数据库
	dsn := completedOpts.Database.DSN()
	store, err := postgres.GetPostgresFactoryOr(dsn)
	if err != nil {
		pkglog.Fatalf("初始化数据库失败: %v", err)
	}
	defer store.Close()

	// 配置数据库连接池
	if dbGetter, ok := store.(interface{ DB() (*gorm.DB, error) }); ok {
		if db, err := dbGetter.DB(); err == nil {
			sqlDB, _ := db.DB()
			sqlDB.SetMaxOpenConns(completedOpts.Database.MaxOpenConnections)
			sqlDB.SetMaxIdleConns(completedOpts.Database.MaxIdleConnections)
			sqlDB.SetConnMaxLifetime(time.Duration(completedOpts.Database.MaxLifetime) * time.Second)
			pkglog.Infof("数据库连接池配置: MaxOpen=%d, MaxIdle=%d, MaxLifetime=%ds",
				completedOpts.Database.MaxOpenConnections,
				completedOpts.Database.MaxIdleConnections,
				completedOpts.Database.MaxLifetime)
		}
	}

	pkglog.Info("数据库初始化成功")

	// 初始化路由
	router := apiserver.InitRouter(store)

	// 创建 HTTP 服务器
	srv := apiserver.NewServer(completedOpts.Server.Addr, router)

	pkglog.Infof("服务器启动在 %s (mode: %s)", completedOpts.Server.Addr, completedOpts.Server.Mode)

	// 优雅关闭
	run(srv, store)
}

// run 运行服务器并处理优雅关闭。
func run(srv *apiserver.Server, store interface{ Close() error }) {
	// 捕获信号
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// 启动服务器
	go func() {
		if err := srv.Run(); err != nil {
			pkglog.Fatalf("服务器启动失败: %v", err)
		}
	}()

	pkglog.Info("服务器启动成功，按 Ctrl+C 停止")

	// 等待信号
	<-ctx.Done()
	stop()

	pkglog.Info("收到关闭信号，开始优雅关闭...")

	// 关闭所有 WebSocket 连接
	apiserver.GetHub().CloseAll()

	// 优雅关闭 HTTP 服务器
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		pkglog.Errorf("服务器关闭失败: %v", err)
	}

	// 关闭数据库连接
	if err := store.Close(); err != nil {
		pkglog.Errorf("关闭数据库连接失败: %v", err)
	}

	pkglog.Info("服务器已关闭")
}
