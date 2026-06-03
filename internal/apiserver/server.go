package apiserver

import (
	"context"
	"net/http"
	"time"
)

// Server HTTP 服务器。
type Server struct {
	*http.Server
}

// NewServer 创建 HTTP 服务器。
func NewServer(addr string, handler http.Handler) *Server {
	return &Server{
		Server: &http.Server{
			Addr:              addr,
			Handler:           handler,
			ReadHeaderTimeout: 5 * time.Second,
			ReadTimeout:       15 * time.Second,
			WriteTimeout:      60 * time.Second,
			IdleTimeout:       60 * time.Second,
		},
	}
}

// Run 启动服务器，阻塞直到服务器关闭。
// 生命周期日志由调用方（main）统一打印，这里不重复记录。
func (s *Server) Run() error {
	if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Shutdown 优雅关闭服务器。
func (s *Server) Shutdown(ctx context.Context) error {
	return s.Server.Shutdown(ctx)
}
