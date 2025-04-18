package metrics

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"strconv"
	"time"
)

// GinMiddleware returns a gin middleware that collects metrics for HTTP requests
func GinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip metrics endpoint to avoid circular measurements
		if c.Request.URL.Path == "/metrics" {
			c.Next()
			return
		}

		// Start timer
		start := time.Now()

		// Process request
		c.Next()

		// Stop timer
		duration := time.Since(start).Seconds()

		// Record metrics
		status := strconv.Itoa(c.Writer.Status())
		RequestCounter.WithLabelValues(c.Request.Method, c.Request.URL.Path, status).Inc()
		RequestDuration.WithLabelValues(c.Request.Method, c.Request.URL.Path).Observe(duration)
	}
}

// RegisterMetricsEndpoint registers the /metrics endpoint with the gin engine
func RegisterMetricsEndpoint(r *gin.Engine) {
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
}
