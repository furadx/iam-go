package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// Config 应用配置。
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
}

// ServerConfig 服务器配置。
type ServerConfig struct {
	Mode string `mapstructure:"mode"` // debug, release
	Addr string `mapstructure:"addr"` // :8080
}

// DatabaseConfig 数据库配置。
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
}

// DSN 返回 PostgreSQL 连接字符串。
func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.DBName, d.SSLMode)
}

// Load 加载配置文件。
func Load(configFile string) (*Config, error) {
	v := viper.New()

	// 设置默认值
	v.SetDefault("server.mode", "release")
	v.SetDefault("server.addr", ":8080")
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.user", "postgres")
	v.SetDefault("database.password", "postgres")
	v.SetDefault("database.dbname", "iam")
	v.SetDefault("database.sslmode", "disable")

	// 如果提供了配置文件，则加载它
	if configFile != "" {
		v.SetConfigFile(configFile)
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("读取配置文件失败: %w", err)
		}
	}

	// 支持环境变量覆盖
	v.SetEnvPrefix("IAM")
	v.AutomaticEnv()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	return &cfg, nil
}

// LoadFromEnv 从环境变量加载配置。
func LoadFromEnv() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Mode: getEnv("IAM_SERVER_MODE", "release"),
			Addr: getEnv("IAM_SERVER_ADDR", ":8080"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("IAM_DB_HOST", "localhost"),
			Port:     5432,
			User:     getEnv("IAM_DB_USER", "postgres"),
			Password: getEnv("IAM_DB_PASSWORD", "postgres"),
			DBName:   getEnv("IAM_DB_NAME", "iam"),
			SSLMode:  getEnv("IAM_DB_SSLMODE", "disable"),
		},
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
