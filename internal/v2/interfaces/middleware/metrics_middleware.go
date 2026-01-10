package middleware

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/contracts"
)

// MetricsMiddleware handles automatic metrics collection for HTTP requests
type MetricsMiddleware struct {
	metrics contracts.MetricsCollector
}

// NewMetricsMiddleware creates a new metrics middleware instance
func NewMetricsMiddleware(metrics contracts.MetricsCollector) *MetricsMiddleware {
	return &MetricsMiddleware{
		metrics: metrics,
	}
}

// Collect returns a Gin middleware that automatically collects HTTP metrics
func (m *MetricsMiddleware) Collect() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Record start time
		start := time.Now()

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start).Seconds()

		// Record metrics
		m.metrics.RecordHttpDuration(c.Request.Method, c.FullPath(), c.Writer.Status(), duration)
		m.metrics.IncrementHttpRequests(c.Request.Method, c.FullPath(), c.Writer.Status())
	}
}
