package options

import (
	"fmt"

	"github.com/spf13/pflag"
)

// LogOptions 日志配置选项。
type LogOptions struct {
	Level            string   `json:"level" mapstructure:"level"`
	Format           string   `json:"format" mapstructure:"format"`
	EnableColor      bool     `json:"enable_color" mapstructure:"enable_color"`
	DisableCaller    bool     `json:"disable_caller" mapstructure:"disable_caller"`
	DisableStacktrace bool     `json:"disable_stacktrace" mapstructure:"disable_stacktrace"`
	OutputPaths      []string `json:"output_paths" mapstructure:"output_paths"`
	ErrorOutputPaths []string `json:"error_output_paths" mapstructure:"error_output_paths"`
}

// NewLogOptions 创建默认的日志选项。
func NewLogOptions() *LogOptions {
	return &LogOptions{
		Level:            "info",
		Format:           "console",
		EnableColor:      true,
		DisableCaller:    false,
		DisableStacktrace: false,
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}
}

// Validate 验证日志选项。
func (l *LogOptions) Validate() []error {
	var errs []error

	// 验证日志级别
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
		"fatal": true,
	}
	if !validLevels[l.Level] {
		errs = append(errs, fmt.Errorf("invalid log level: %s (must be debug, info, warn, error, or fatal)", l.Level))
	}

	// 验证日志格式
	if l.Format != "json" && l.Format != "console" {
		errs = append(errs, fmt.Errorf("invalid log format: %s (must be json or console)", l.Format))
	}

	// 验证输出路径
	if len(l.OutputPaths) == 0 {
		errs = append(errs, fmt.Errorf("output paths cannot be empty"))
	}

	if len(l.ErrorOutputPaths) == 0 {
		errs = append(errs, fmt.Errorf("error output paths cannot be empty"))
	}

	return errs
}

// AddFlags 添加命令行标志。
func (l *LogOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&l.Level, "log.level", l.Level, "Log level (debug|info|warn|error|fatal)")
	fs.StringVar(&l.Format, "log.format", l.Format, "Log format (json|console)")
	fs.BoolVar(&l.EnableColor, "log.enable-color", l.EnableColor, "Enable colored output (console format only)")
	fs.BoolVar(&l.DisableCaller, "log.disable-caller", l.DisableCaller, "Disable caller information")
	fs.BoolVar(&l.DisableStacktrace, "log.disable-stacktrace", l.DisableStacktrace, "Disable stacktrace")
	fs.StringSliceVar(&l.OutputPaths, "log.output-paths", l.OutputPaths, "Output paths")
	fs.StringSliceVar(&l.ErrorOutputPaths, "log.error-output-paths", l.ErrorOutputPaths, "Error output paths")
}
