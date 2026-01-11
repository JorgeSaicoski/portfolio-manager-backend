package prometheus

import (
	"strconv"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/contracts"
	"github.com/prometheus/client_golang/prometheus"
)

// metricsCollector is the Prometheus implementation of MetricsCollector
type metricsCollector struct {
	// Portfolio metrics
	portfoliosCreated prometheus.Counter
	portfoliosUpdated prometheus.Counter
	portfoliosDeleted prometheus.Counter

	// Category metrics
	categoriesCreated prometheus.Counter
	categoriesUpdated prometheus.Counter
	categoriesDeleted prometheus.Counter

	// Section metrics
	sectionsCreated prometheus.Counter
	sectionsUpdated prometheus.Counter
	sectionsDeleted prometheus.Counter

	// User metrics
	usersCreated prometheus.Counter
	usersUpdated prometheus.Counter
	usersDeleted prometheus.Counter

	// Authentication metrics
	authAttempts *prometheus.CounterVec
	jwtTokens    *prometheus.CounterVec

	// HTTP metrics
	httpRequestsTotal   *prometheus.CounterVec
	httpRequestDuration *prometheus.HistogramVec
}

// NewMetricsCollector creates a new Prometheus metrics collector
// Returns the interface type (contracts.MetricsCollector), not the concrete type
func NewMetricsCollector() contracts.MetricsCollector {
	collector := &metricsCollector{
		// Portfolio metrics
		portfoliosCreated: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "portfolios_created_total",
			Help: "Total number of portfolios created",
		}),
		portfoliosUpdated: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "portfolios_updated_total",
			Help: "Total number of portfolios updated",
		}),
		portfoliosDeleted: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "portfolios_deleted_total",
			Help: "Total number of portfolios deleted",
		}),

		// Category metrics
		categoriesCreated: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "categories_created_total",
			Help: "Total number of categories created",
		}),
		categoriesUpdated: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "categories_updated_total",
			Help: "Total number of categories updated",
		}),
		categoriesDeleted: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "categories_deleted_total",
			Help: "Total number of categories deleted",
		}),

		// Section metrics
		sectionsCreated: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "sections_created_total",
			Help: "Total number of sections created",
		}),
		sectionsUpdated: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "sections_updated_total",
			Help: "Total number of sections updated",
		}),
		sectionsDeleted: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "sections_deleted_total",
			Help: "Total number of sections deleted",
		}),

		// User metrics
		usersCreated: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "users_created_total",
			Help: "Total number of users created",
		}),
		usersUpdated: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "users_updated_total",
			Help: "Total number of users updated",
		}),
		usersDeleted: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "users_deleted_total",
			Help: "Total number of users deleted",
		}),

		// Authentication metrics
		authAttempts: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "auth_attempts_total",
				Help: "Total number of authentication attempts",
			},
			[]string{"auth_type", "status"},
		),
		jwtTokens: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "jwt_tokens_issued_total",
				Help: "Total number of JWT tokens issued",
			},
			[]string{"token_type"},
		),

		// HTTP metrics
		httpRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		),
		httpRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path"},
		),
	}

	// Register all metrics with Prometheus
	prometheus.MustRegister(
		// Portfolio metrics
		collector.portfoliosCreated,
		collector.portfoliosUpdated,
		collector.portfoliosDeleted,

		// Category metrics
		collector.categoriesCreated,
		collector.categoriesUpdated,
		collector.categoriesDeleted,

		// Section metrics
		collector.sectionsCreated,
		collector.sectionsUpdated,
		collector.sectionsDeleted,

		// User metrics
		collector.usersCreated,
		collector.usersUpdated,
		collector.usersDeleted,

		// Authentication metrics
		collector.authAttempts,
		collector.jwtTokens,

		// HTTP metrics
		collector.httpRequestsTotal,
		collector.httpRequestDuration,
	)

	return collector
}

// Portfolio metrics implementation

func (m *metricsCollector) IncrementPortfoliosCreated() {
	m.portfoliosCreated.Inc()
}

func (m *metricsCollector) IncrementPortfoliosUpdated() {
	m.portfoliosUpdated.Inc()
}

func (m *metricsCollector) IncrementPortfoliosDeleted() {
	m.portfoliosDeleted.Inc()
}

// Category metrics implementation

func (m *metricsCollector) IncrementCategoriesCreated() {
	m.categoriesCreated.Inc()
}

func (m *metricsCollector) IncrementCategoriesUpdated() {
	m.categoriesUpdated.Inc()
}

func (m *metricsCollector) IncrementCategoriesDeleted() {
	m.categoriesDeleted.Inc()
}

// Section metrics implementation

func (m *metricsCollector) IncrementSectionsCreated() {
	m.sectionsCreated.Inc()
}

func (m *metricsCollector) IncrementSectionsUpdated() {
	m.sectionsUpdated.Inc()
}

func (m *metricsCollector) IncrementSectionsDeleted() {
	m.sectionsDeleted.Inc()
}

// User metrics implementation

func (m *metricsCollector) IncrementUsersCreated() {
	m.usersCreated.Inc()
}

func (m *metricsCollector) IncrementUsersUpdated() {
	m.usersUpdated.Inc()
}

func (m *metricsCollector) IncrementUsersDeleted() {
	m.usersDeleted.Inc()
}

// Authentication metrics implementation

func (m *metricsCollector) IncrementAuthAttempts(authType, status string) {
	m.authAttempts.WithLabelValues(authType, status).Inc()
}

func (m *metricsCollector) IncrementJwtTokens(tokenType string) {
	m.jwtTokens.WithLabelValues(tokenType).Inc()
}

// HTTP metrics implementation

func (m *metricsCollector) RecordHttpDuration(method, path string, status int, duration float64) {
	m.httpRequestDuration.WithLabelValues(method, path).Observe(duration)
}

func (m *metricsCollector) IncrementHttpRequests(method, path string, status int) {
	m.httpRequestsTotal.WithLabelValues(method, path, strconv.Itoa(status)).Inc()
}
