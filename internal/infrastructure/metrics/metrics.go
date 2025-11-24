package metrics

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"gorm.io/gorm"
)

type Collector struct {
	HttpRequestsTotal   *prometheus.CounterVec
	HttpRequestDuration *prometheus.HistogramVec
	DatabaseConnections *prometheus.GaugeVec
	PortfoliosTotal     prometheus.Gauge
	AuthAttempts        *prometheus.CounterVec
	JwtTokensGenerated  *prometheus.CounterVec
	ImagesUploaded      prometheus.Counter
	ImagesDeleted       prometheus.Counter
}

func NewCollector() *Collector {
	collector := &Collector{
		HttpRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		),

		HttpRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "Duration of HTTP requests in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path", "status"},
		),

		DatabaseConnections: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "database_connections",
				Help: "Number of database connections",
			},
			[]string{"state"}, // active, idle, in_use
		),

		PortfoliosTotal: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "portfolios_total",
				Help: "Total number of portfolios created",
			},
		),

		AuthAttempts: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "auth_attempts_total",
				Help: "Total number of authentication attempts",
			},
			[]string{"type", "status"}, // login/register, success/failure
		),

		JwtTokensGenerated: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "jwt_tokens_generated_total",
				Help: "Total number of JWT tokens generated",
			},
			[]string{"type"}, // access, refresh
		),

		ImagesUploaded: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "images_uploaded_total",
				Help: "Total number of images uploaded",
			},
		),

		ImagesDeleted: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "images_deleted_total",
				Help: "Total number of images deleted",
			},
		),
	}

	collector.registerMetrics()
	return collector
}

func (c *Collector) registerMetrics() {
	prometheus.MustRegister(
		c.HttpRequestsTotal,
		c.HttpRequestDuration,
		c.DatabaseConnections,
		c.PortfoliosTotal,
		c.AuthAttempts,
		c.JwtTokensGenerated,
		c.ImagesUploaded,
		c.ImagesDeleted,
	)
}

// HTTP Metrics
func (c *Collector) IncrementHttpRequests(method, path string, status int) {
	c.HttpRequestsTotal.WithLabelValues(method, path, strconv.Itoa(status)).Inc()
}

func (c *Collector) RecordHttpDuration(method, path string, status int, duration float64) {
	c.HttpRequestDuration.WithLabelValues(method, path, strconv.Itoa(status)).Observe(duration)
}

// Database Metrics
func (c *Collector) UpdateDatabaseConnections(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err != nil {
		return
	}

	stats := sqlDB.Stats()
	c.DatabaseConnections.WithLabelValues("active").Set(float64(stats.OpenConnections))
	c.DatabaseConnections.WithLabelValues("idle").Set(float64(stats.Idle))
	c.DatabaseConnections.WithLabelValues("in_use").Set(float64(stats.InUse))
}

// Business Metrics

func (c *Collector) UpdatePortfoliosTotal(count int64) {
	c.PortfoliosTotal.Set(float64(count))
}

func (c *Collector) IncrementAuthAttempts(authType, status string) {
	c.AuthAttempts.WithLabelValues(authType, status).Inc()
}

func (c *Collector) IncrementJwtTokens(tokenType string) {
	c.JwtTokensGenerated.WithLabelValues(tokenType).Inc()
}

// Image Metrics
func (c *Collector) IncImagesUploaded() {
	c.ImagesUploaded.Inc()
}

func (c *Collector) IncImagesDeleted() {
	c.ImagesDeleted.Inc()
}

// Background metrics collection
func (c *Collector) StartMetricsCollection(db *gorm.DB) {
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		for range ticker.C {
			c.collectBusinessMetrics(db)
			c.UpdateDatabaseConnections(db)
		}
	}()
}

func (c *Collector) collectBusinessMetrics(db *gorm.DB) {

	// Count portfolios
	var portfolioCount int64
	if err := db.Table("portfolios").Count(&portfolioCount).Error; err == nil {
		c.UpdatePortfoliosTotal(portfolioCount)
	}
}
