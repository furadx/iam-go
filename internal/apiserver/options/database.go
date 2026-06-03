package options

import (
	"fmt"

	"github.com/spf13/pflag"
)

// DatabaseOptions 数据库配置选项。
type DatabaseOptions struct {
	Host               string `json:"host" mapstructure:"host"`
	Port               int    `json:"port" mapstructure:"port"`
	User               string `json:"user" mapstructure:"user"`
	Password           string `json:"-" mapstructure:"password"` // 不输出到 JSON
	DBName             string `json:"dbname" mapstructure:"dbname"`
	SSLMode            string `json:"sslmode" mapstructure:"sslmode"`
	MaxOpenConnections int    `json:"max_open_connections" mapstructure:"max_open_connections"`
	MaxIdleConnections int    `json:"max_idle_connections" mapstructure:"max_idle_connections"`
	MaxLifetime        int    `json:"max_lifetime" mapstructure:"max_lifetime"` // 秒
}

// NewDatabaseOptions 创建默认的数据库选项。
func NewDatabaseOptions() *DatabaseOptions {
	return &DatabaseOptions{
		Host:               "localhost",
		Port:               5432,
		User:               "postgres",
		Password:           "postgres",
		DBName:             "iam",
		SSLMode:            "disable",
		MaxOpenConnections: 100,
		MaxIdleConnections: 10,
		MaxLifetime:        3600, // 1 小时
	}
}

// Validate 验证数据库选项。
func (d *DatabaseOptions) Validate() []error {
	var errs []error

	if d.Host == "" {
		errs = append(errs, fmt.Errorf("database host cannot be empty"))
	}

	if d.Port < 1 || d.Port > 65535 {
		errs = append(errs, fmt.Errorf("invalid database port: %d", d.Port))
	}

	if d.User == "" {
		errs = append(errs, fmt.Errorf("database user cannot be empty"))
	}

	if d.DBName == "" {
		errs = append(errs, fmt.Errorf("database name cannot be empty"))
	}

	if d.MaxOpenConnections < 1 {
		errs = append(errs, fmt.Errorf("max open connections must be at least 1"))
	}

	if d.MaxIdleConnections < 0 {
		errs = append(errs, fmt.Errorf("max idle connections cannot be negative"))
	}

	if d.MaxIdleConnections > d.MaxOpenConnections {
		errs = append(errs, fmt.Errorf("max idle connections (%d) cannot exceed max open connections (%d)",
			d.MaxIdleConnections, d.MaxOpenConnections))
	}

	return errs
}

// AddFlags 添加命令行标志。
func (d *DatabaseOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&d.Host, "db.host", d.Host, "Database host")
	fs.IntVar(&d.Port, "db.port", d.Port, "Database port")
	fs.StringVar(&d.User, "db.user", d.User, "Database user")
	fs.StringVar(&d.Password, "db.password", d.Password, "Database password")
	fs.StringVar(&d.DBName, "db.name", d.DBName, "Database name")
	fs.StringVar(&d.SSLMode, "db.sslmode", d.SSLMode, "Database SSL mode")
	fs.IntVar(&d.MaxOpenConnections, "db.max-open-connections", d.MaxOpenConnections, "Maximum open database connections")
	fs.IntVar(&d.MaxIdleConnections, "db.max-idle-connections", d.MaxIdleConnections, "Maximum idle database connections")
	fs.IntVar(&d.MaxLifetime, "db.max-lifetime", d.MaxLifetime, "Maximum connection lifetime in seconds")
}

// DSN 返回 PostgreSQL 连接字符串。
func (d *DatabaseOptions) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.DBName, d.SSLMode)
}
