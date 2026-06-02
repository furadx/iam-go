package apiserver

import (
	"context"
	"log"
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

// Run 启动服务器。
func (s *Server) Run() error {
	log.Printf("服务启动在 %s", s.Addr)
	if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Shutdown 优雅关闭服务器。
func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("开始优雅关闭服务器...")
	if err := s.Server.Shutdown(ctx); err != nil {
		return err
	}
	log.Println("服务器已关闭")
	return nil
}
