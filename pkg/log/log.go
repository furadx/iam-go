package log

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var std *zap.Logger

func init() {
	std = New(NewOptions())
}

// Options 日志配置选项。
type Options struct {
	Level            string // 日志级别：debug, info, warn, error
	Format           string // 日志格式：json, console
	EnableColor      bool   // 是否启用颜色（仅 console 格式）
	DisableCaller    bool   // 是否禁用调用者信息
	DisableStacktrace bool   // 是否禁用堆栈跟踪
	OutputPaths      []string // 输出路径
	ErrorOutputPaths []string // 错误输出路径
}

// NewOptions 创建默认日志选项。
func NewOptions() *Options {
	return &Options{
		Level:            "info",
		Format:           "console",
		EnableColor:      true,
		DisableCaller:    false,
		DisableStacktrace: false,
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}
}

// New 创建新的日志记录器。
func New(opts *Options) *zap.Logger {
	// 解析日志级别
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(opts.Level)); err != nil {
		level = zapcore.InfoLevel
	}

	// 创建编码配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 如果是 console 格式且启用颜色
	if opts.Format == "console" && opts.EnableColor {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	// 创建编码器
	var encoder zapcore.Encoder
	if opts.Format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// 创建输出
	writeSyncer := getWriteSyncer(opts.OutputPaths)

	// 创建 core
	core := zapcore.NewCore(encoder, writeSyncer, level)

	// 创建 logger 配置
	logger := zap.New(core)

	// 添加调用者信息
	if !opts.DisableCaller {
		logger = logger.WithOptions(zap.AddCaller())
	}

	// 添加堆栈跟踪
	if !opts.DisableStacktrace {
		logger = logger.WithOptions(zap.AddStacktrace(zapcore.ErrorLevel))
	}

	return logger
}

func getWriteSyncer(paths []string) zapcore.WriteSyncer {
	var writers []zapcore.WriteSyncer

	for _, path := range paths {
		switch path {
		case "stdout":
			writers = append(writers, zapcore.AddSync(os.Stdout))
		case "stderr":
			writers = append(writers, zapcore.AddSync(os.Stderr))
		default:
			// 文件输出（简化版本，生产环境应使用 lumberjack 进行日志轮转）
			file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err == nil {
				writers = append(writers, zapcore.AddSync(file))
			}
		}
	}

	return zapcore.NewMultiWriteSyncer(writers...)
}

// SetLogger 设置全局日志记录器。
func SetLogger(logger *zap.Logger) {
	std = logger
}

// Default 返回全局日志记录器。
func Default() *zap.Logger {
	return std
}

// Debug 记录调试级别日志。
func Debug(msg string, fields ...zap.Field) {
	std.Debug(msg, fields...)
}

// Info 记录信息级别日志。
func Info(msg string, fields ...zap.Field) {
	std.Info(msg, fields...)
}

// Warn 记录警告级别日志。
func Warn(msg string, fields ...zap.Field) {
	std.Warn(msg, fields...)
}

// Error 记录错误级别日志。
func Error(msg string, fields ...zap.Field) {
	std.Error(msg, fields...)
}

// Fatal 记录致命级别日志并退出程序。
func Fatal(msg string, fields ...zap.Field) {
	std.Fatal(msg, fields...)
}

// Debugf 记录格式化的调试级别日志。
func Debugf(format string, args ...interface{}) {
	std.Sugar().Debugf(format, args...)
}

// Infof 记录格式化的信息级别日志。
func Infof(format string, args ...interface{}) {
	std.Sugar().Infof(format, args...)
}

// Warnf 记录格式化的警告级别日志。
func Warnf(format string, args ...interface{}) {
	std.Sugar().Warnf(format, args...)
}

// Errorf 记录格式化的错误级别日志。
func Errorf(format string, args ...interface{}) {
	std.Sugar().Errorf(format, args...)
}

// Fatalf 记录格式化的致命级别日志并退出程序。
func Fatalf(format string, args ...interface{}) {
	std.Sugar().Fatalf(format, args...)
}

// With 添加字段到日志记录器。
func With(fields ...zap.Field) *zap.Logger {
	return std.With(fields...)
}

// Sync 刷新日志缓冲区。
func Sync() error {
	return std.Sync()
}
