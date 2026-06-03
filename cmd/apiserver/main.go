package main

import (
	"context"
	"flag"
	"fmt"
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
	"github.com/furadx/iam-go/internal/pkg/authz"
	"github.com/furadx/iam-go/internal/pkg/loginlock"
	"github.com/furadx/iam-go/internal/pkg/middleware"
	"github.com/furadx/iam-go/internal/pkg/password"
	"github.com/furadx/iam-go/internal/pkg/ratelimit"
	redispkg "github.com/furadx/iam-go/internal/pkg/redis"
	"github.com/furadx/iam-go/internal/pkg/revoke"
	pkglog "github.com/furadx/iam-go/pkg/log"
	"github.com/furadx/iam-go/pkg/token"
)

var (
	configFile  string
	showVersion bool
	Version     = "v1.0.0"
	BuildDate   = "unknown"
	GitCommit   = "unknown"
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
	flagOverrides, err := captureFlagOverrides(pflag.CommandLine)
	if err != nil {
		log.Fatalf("读取命令行参数失败: %v", err)
	}

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
	if err := applyFlagOverrides(opts, flagOverrides); err != nil {
		log.Fatalf("应用命令行参数失败: %v", err)
	}

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

	// 初始化 JWT Token 管理器（双令牌）
	jwtManager := token.NewManager(
		completedOpts.JWT.Secret,
		time.Duration(completedOpts.JWT.AccessExpire)*time.Second,
		time.Duration(completedOpts.JWT.RefreshExpire)*time.Second,
	)

	// 初始化 Redis 与吊销组件
	redisClient, err := redispkg.New(
		completedOpts.Redis.Addr,
		completedOpts.Redis.Password,
		completedOpts.Redis.DB,
	)
	if err != nil {
		pkglog.Fatalf("初始化 Redis 失败: %v", err)
	}
	revoker := revoke.NewRedisRevoker(redisClient)
	loginGuard := loginlock.NewRedisLocker(redisClient, loginlock.Config{
		Enabled:         completedOpts.Security.LoginLock.Enabled,
		UserMaxFailures: completedOpts.Security.LoginLock.UserMaxFailures,
		IPMaxFailures:   completedOpts.Security.LoginLock.IPMaxFailures,
		FailureWindow:   time.Duration(completedOpts.Security.LoginLock.FailureWindowMin) * time.Minute,
		LockDuration:    time.Duration(completedOpts.Security.LoginLock.LockMinutes) * time.Minute,
	})
	rateLimiter := ratelimit.NewRedisLimiter(redisClient)
	passwordPolicy := password.Policy{
		MinLength:             completedOpts.Security.PasswordPolicy.MinLength,
		MaxLength:             completedOpts.Security.PasswordPolicy.MaxLength,
		MinClasses:            completedOpts.Security.PasswordPolicy.MinClasses,
		RejectUsername:        completedOpts.Security.PasswordPolicy.RejectUsername,
		RejectCommonPasswords: completedOpts.Security.PasswordPolicy.RejectCommonPasswords,
	}
	pkglog.Info("Redis 初始化成功")

	// 初始化 Casbin enforcer（复用 gorm 连接）
	casbinDBGetter, casbinOK := store.(interface{ DB() (*gorm.DB, error) })
	if !casbinOK {
		pkglog.Fatal("store 不支持获取 *gorm.DB，无法初始化 Casbin")
	}
	gdb, err := casbinDBGetter.DB()
	if err != nil {
		pkglog.Fatalf("获取数据库连接失败: %v", err)
	}
	authzManager, err := authz.NewManager(gdb, "configs/rbac_model.conf")
	if err != nil {
		pkglog.Fatalf("初始化 Casbin 失败: %v", err)
	}
	if err := authzManager.SeedDefaults(); err != nil {
		pkglog.Fatalf("初始化默认策略失败: %v", err)
	}
	pkglog.Info("Casbin 初始化成功")

	// 初始化路由
	router := apiserver.InitRouter(apiserver.RouterDeps{
		Store:   store,
		Token:   jwtManager,
		Revoker: revoker,
		Authz:   authzManager,
		Security: apiserver.SecurityDeps{
			LoginGuard:     loginGuard,
			PasswordPolicy: passwordPolicy,
			RateLimiter:    rateLimiter,
			CORS: middleware.CORSConfig{
				AllowedOrigins:   completedOpts.Security.CORS.AllowedOrigins,
				AllowCredentials: completedOpts.Security.CORS.AllowCredentials,
				MaxAge:           time.Duration(completedOpts.Security.CORS.MaxAgeSeconds) * time.Second,
			},
			APIRateLimit: middleware.RateLimitConfig{
				Enabled:  completedOpts.Security.RateLimit.Enabled,
				Name:     "api",
				Limit:    completedOpts.Security.RateLimit.APILimit,
				Window:   time.Duration(completedOpts.Security.RateLimit.APIWindowSeconds) * time.Second,
				FailOpen: completedOpts.Security.RateLimit.FailOpen,
			},
			LoginRateLimit: middleware.RateLimitConfig{
				Enabled:  completedOpts.Security.RateLimit.Enabled,
				Name:     "login",
				Limit:    completedOpts.Security.RateLimit.LoginIPLimit,
				Window:   time.Duration(completedOpts.Security.RateLimit.LoginWindowSeconds) * time.Second,
				FailOpen: completedOpts.Security.RateLimit.FailOpen,
			},
			RevokeFailOpen: completedOpts.JWT.RevokeFailOpen,
		},
	})

	// 创建 HTTP 服务器
	srv := apiserver.NewServer(completedOpts.Server.Addr, router)

	pkglog.Infof("服务器启动在 %s (mode: %s)", completedOpts.Server.Addr, completedOpts.Server.Mode)

	// 优雅关闭
	run(srv, store, redisClient)
}

// run 运行服务器并处理优雅关闭。
func run(srv *apiserver.Server, store interface{ Close() error }, redisClient interface{ Close() error }) {
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

	if err := redisClient.Close(); err != nil {
		pkglog.Errorf("关闭 Redis 连接失败: %v", err)
	}

	pkglog.Info("服务器已关闭")
}

type flagOverrides map[string]any

func captureFlagOverrides(fs *pflag.FlagSet) (flagOverrides, error) {
	values := flagOverrides{}
	var errs []error
	fs.Visit(func(f *pflag.Flag) {
		value, err := getFlagValue(fs, f.Name)
		if err != nil {
			errs = append(errs, err)
			return
		}
		if value != nil {
			values[f.Name] = value
		}
	})
	if len(errs) > 0 {
		return nil, fmt.Errorf("capture flag overrides: %v", errs)
	}
	return values, nil
}

func getFlagValue(fs *pflag.FlagSet, name string) (any, error) {
	switch name {
	case "server.mode", "server.addr", "db.host", "db.user", "db.password", "db.name", "db.sslmode",
		"log.level", "log.format", "jwt.secret", "redis.addr", "redis.password":
		return fs.GetString(name)
	case "server.middlewares", "log.output-paths", "log.error-output-paths", "security.cors.allowed-origins":
		return fs.GetStringSlice(name)
	case "server.healthz", "log.enable-color", "log.disable-caller", "log.disable-stacktrace",
		"jwt.revoke-fail-open", "security.rate-limit.enabled", "security.rate-limit.fail-open",
		"security.login-lock.enabled", "security.password-policy.reject-username",
		"security.password-policy.reject-common-passwords", "security.cors.allow-credentials":
		return fs.GetBool(name)
	case "db.port", "db.max-open-connections", "db.max-idle-connections", "db.max-lifetime",
		"jwt.access-expire", "jwt.refresh-expire", "redis.db", "security.rate-limit.api-limit",
		"security.rate-limit.api-window-seconds", "security.rate-limit.login-ip-limit",
		"security.rate-limit.login-window-seconds", "security.login-lock.user-max-failures",
		"security.login-lock.ip-max-failures", "security.login-lock.failure-window-minutes",
		"security.login-lock.lock-minutes", "security.password-policy.min-length",
		"security.password-policy.max-length", "security.password-policy.min-classes",
		"security.cors.max-age-seconds":
		return fs.GetInt(name)
	}
	return nil, nil
}

func applyFlagOverrides(opts *options.Options, values flagOverrides) error {
	for name, value := range values {
		if err := applyFlagOverride(opts, name, value); err != nil {
			return err
		}
	}
	return nil
}

func applyFlagOverride(opts *options.Options, name string, value any) error {
	switch name {
	case "server.mode":
		opts.Server.Mode = value.(string)
	case "server.addr":
		opts.Server.Addr = value.(string)
	case "server.healthz":
		opts.Server.Healthz = value.(bool)
	case "server.middlewares":
		opts.Server.Middlewares = value.([]string)
	case "db.host":
		opts.Database.Host = value.(string)
	case "db.port":
		opts.Database.Port = value.(int)
	case "db.user":
		opts.Database.User = value.(string)
	case "db.password":
		opts.Database.Password = value.(string)
	case "db.name":
		opts.Database.DBName = value.(string)
	case "db.sslmode":
		opts.Database.SSLMode = value.(string)
	case "db.max-open-connections":
		opts.Database.MaxOpenConnections = value.(int)
	case "db.max-idle-connections":
		opts.Database.MaxIdleConnections = value.(int)
	case "db.max-lifetime":
		opts.Database.MaxLifetime = value.(int)
	case "log.level":
		opts.Log.Level = value.(string)
	case "log.format":
		opts.Log.Format = value.(string)
	case "log.enable-color":
		opts.Log.EnableColor = value.(bool)
	case "log.disable-caller":
		opts.Log.DisableCaller = value.(bool)
	case "log.disable-stacktrace":
		opts.Log.DisableStacktrace = value.(bool)
	case "log.output-paths":
		opts.Log.OutputPaths = value.([]string)
	case "log.error-output-paths":
		opts.Log.ErrorOutputPaths = value.([]string)
	case "jwt.secret":
		opts.JWT.Secret = value.(string)
	case "jwt.access-expire":
		opts.JWT.AccessExpire = value.(int)
	case "jwt.refresh-expire":
		opts.JWT.RefreshExpire = value.(int)
	case "jwt.revoke-fail-open":
		opts.JWT.RevokeFailOpen = value.(bool)
	case "redis.addr":
		opts.Redis.Addr = value.(string)
	case "redis.password":
		opts.Redis.Password = value.(string)
	case "redis.db":
		opts.Redis.DB = value.(int)
	case "security.rate-limit.enabled":
		opts.Security.RateLimit.Enabled = value.(bool)
	case "security.rate-limit.api-limit":
		opts.Security.RateLimit.APILimit = value.(int)
	case "security.rate-limit.api-window-seconds":
		opts.Security.RateLimit.APIWindowSeconds = value.(int)
	case "security.rate-limit.login-ip-limit":
		opts.Security.RateLimit.LoginIPLimit = value.(int)
	case "security.rate-limit.login-window-seconds":
		opts.Security.RateLimit.LoginWindowSeconds = value.(int)
	case "security.rate-limit.fail-open":
		opts.Security.RateLimit.FailOpen = value.(bool)
	case "security.login-lock.enabled":
		opts.Security.LoginLock.Enabled = value.(bool)
	case "security.login-lock.user-max-failures":
		opts.Security.LoginLock.UserMaxFailures = value.(int)
	case "security.login-lock.ip-max-failures":
		opts.Security.LoginLock.IPMaxFailures = value.(int)
	case "security.login-lock.failure-window-minutes":
		opts.Security.LoginLock.FailureWindowMin = value.(int)
	case "security.login-lock.lock-minutes":
		opts.Security.LoginLock.LockMinutes = value.(int)
	case "security.password-policy.min-length":
		opts.Security.PasswordPolicy.MinLength = value.(int)
	case "security.password-policy.max-length":
		opts.Security.PasswordPolicy.MaxLength = value.(int)
	case "security.password-policy.min-classes":
		opts.Security.PasswordPolicy.MinClasses = value.(int)
	case "security.password-policy.reject-username":
		opts.Security.PasswordPolicy.RejectUsername = value.(bool)
	case "security.password-policy.reject-common-passwords":
		opts.Security.PasswordPolicy.RejectCommonPasswords = value.(bool)
	case "security.cors.allowed-origins":
		opts.Security.CORS.AllowedOrigins = value.([]string)
	case "security.cors.allow-credentials":
		opts.Security.CORS.AllowCredentials = value.(bool)
	case "security.cors.max-age-seconds":
		opts.Security.CORS.MaxAgeSeconds = value.(int)
	}
	return nil
}
