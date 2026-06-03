package middleware

import (
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// CORSConfig defines browser cross-origin policy.
type CORSConfig struct {
	AllowedOrigins   []string
	AllowCredentials bool
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposeHeaders    []string
	MaxAge           time.Duration
}

// CORS allows browser cross-origin access for configured origins.
func CORS(cfg CORSConfig) gin.HandlerFunc {
	methods := joinOrDefault(cfg.AllowedMethods, "GET,POST,PUT,PATCH,DELETE,OPTIONS")
	headers := joinOrDefault(cfg.AllowedHeaders, "Origin,Content-Type,Accept,Authorization")
	expose := joinOrDefault(cfg.ExposeHeaders, "Content-Length")
	maxAge := cfg.MaxAge
	if maxAge <= 0 {
		maxAge = 12 * time.Hour
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" && originAllowed(origin, cfg.AllowedOrigins) {
			c.Header("Access-Control-Allow-Origin", origin)
			if cfg.AllowCredentials {
				c.Header("Access-Control-Allow-Credentials", "true")
			}
			c.Header("Vary", "Origin")
		}

		c.Header("Access-Control-Allow-Methods", methods)
		c.Header("Access-Control-Allow-Headers", headers)
		c.Header("Access-Control-Expose-Headers", expose)
		c.Header("Access-Control-Max-Age", strconv.FormatInt(int64(maxAge.Seconds()), 10))

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func originAllowed(origin string, allowed []string) bool {
	origin = strings.TrimSpace(origin)
	if origin == "" {
		return false
	}

	cleaned := make([]string, 0, len(allowed))
	for _, item := range allowed {
		item = strings.TrimSpace(item)
		if item != "" {
			cleaned = append(cleaned, item)
		}
	}
	return slices.Contains(cleaned, origin)
}

func joinOrDefault(values []string, fallback string) string {
	if len(values) == 0 {
		return fallback
	}
	return strings.Join(values, ",")
}
