package options

import (
	"fmt"
	"net"

	"github.com/spf13/pflag"
)

// ServerOptions 服务器配置选项。
type ServerOptions struct {
	Mode        string   `json:"mode" mapstructure:"mode"`
	Addr        string   `json:"addr" mapstructure:"addr"`
	Healthz     bool     `json:"healthz" mapstructure:"healthz"`
	Middlewares []string `json:"middlewares" mapstructure:"middlewares"`
}

// NewServerOptions 创建默认的服务器选项。
func NewServerOptions() *ServerOptions {
	return &ServerOptions{
		Mode:        "release",
		Addr:        ":8080",
		Healthz:     true,
		Middlewares: []string{},
	}
}

// Validate 验证服务器选项。
func (s *ServerOptions) Validate() []error {
	var errs []error

	// 验证模式
	if s.Mode != "debug" && s.Mode != "release" && s.Mode != "test" {
		errs = append(errs, fmt.Errorf("invalid mode: %s (must be debug, release, or test)", s.Mode))
	}

	// 验证地址
	if _, err := net.ResolveTCPAddr("tcp", s.Addr); err != nil {
		errs = append(errs, fmt.Errorf("invalid address: %s (%v)", s.Addr, err))
	}

	return errs
}

// AddFlags 添加命令行标志。
func (s *ServerOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&s.Mode, "server.mode", s.Mode, "Server mode (debug|release|test)")
	fs.StringVar(&s.Addr, "server.addr", s.Addr, "Server listen address")
	fs.BoolVar(&s.Healthz, "server.healthz", s.Healthz, "Enable health check endpoint")
	fs.StringSliceVar(&s.Middlewares, "server.middlewares", s.Middlewares, "List of middlewares to enable")
}
