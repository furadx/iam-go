package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCORSAllowsOnlyConfiguredOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(CORS(CORSConfig{
		AllowedOrigins:   []string{"https://console.example.com"},
		AllowCredentials: true,
	}))
	r.GET("/x", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	allowed := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Origin", "https://console.example.com")
	r.ServeHTTP(allowed, req)
	if got := allowed.Header().Get("Access-Control-Allow-Origin"); got != "https://console.example.com" {
		t.Fatalf("expected configured origin to be allowed, got %q", got)
	}
	if got := allowed.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Fatalf("expected credentials header, got %q", got)
	}

	blocked := httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Origin", "https://evil.example.com")
	r.ServeHTTP(blocked, req)
	if got := blocked.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("expected unconfigured origin to be blocked, got %q", got)
	}
}
