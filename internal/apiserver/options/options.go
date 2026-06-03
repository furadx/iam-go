package options

import (
	"github.com/spf13/pflag"
)

// Options 应用程序完整选项。
type Options struct {
	Server   *ServerOptions   `json:"server" mapstructure:"server"`
	Database *DatabaseOptions `json:"database" mapstructure:"database"`
	Log      *LogOptions      `json:"log" mapstructure:"log"`
}

// NewOptions 创建默认选项。
func NewOptions() *Options {
	return &Options{
		Server:   NewServerOptions(),
		Database: NewDatabaseOptions(),
		Log:      NewLogOptions(),
	}
}

// Validate 验证所有选项。
func (o *Options) Validate() []error {
	var errs []error

	errs = append(errs, o.Server.Validate()...)
	errs = append(errs, o.Database.Validate()...)
	errs = append(errs, o.Log.Validate()...)

	return errs
}

// AddFlags 添加所有命令行标志。
func (o *Options) AddFlags(fs *pflag.FlagSet) {
	o.Server.AddFlags(fs)
	o.Database.AddFlags(fs)
	o.Log.AddFlags(fs)
}

// Complete 填充默认值和派生字段。
func (o *Options) Complete() *CompletedOptions {
	// 可以在这里添加基于其他选项的派生逻辑
	// 例如：如果 mode 是 debug，自动设置 log level 为 debug
	if o.Server.Mode == "debug" && o.Log.Level == "info" {
		o.Log.Level = "debug"
	}

	return &CompletedOptions{o}
}

// CompletedOptions 包含已完成的选项。
type CompletedOptions struct {
	*Options
}
