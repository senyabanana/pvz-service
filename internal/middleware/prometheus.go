package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/senyabanana/pvz-service/internal/infrastructure/monitoring"
)

func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		duration := time.Since(start).Seconds()
		status := c.Writer.Status()
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		monitoring.TotalRequests.WithLabelValues(c.Request.Method, path, http.StatusText(status)).Inc()
		monitoring.RequestDuration.WithLabelValues(c.Request.Method, path).Observe(duration)
	}
}
