package metrics

import (
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/contracts"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// metricsCollector implements the MetricsCollector contract
type metricsCollector struct {
	// Portfolio metrics
	portfoliosCreated prometheus.Counter
	portfoliosUpdated prometheus.Counter
	portfoliosDeleted prometheus.Counter

	// User metrics
	usersLoggedIn   prometheus.Counter
	usersRegistered prometheus.Counter
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(registry prometheus.Registerer) contracts.MetricsCollector {
	return &metricsCollector{
		portfoliosCreated: promauto.With(registry).NewCounter(prometheus.CounterOpts{
			Name: "portfolios_created_total",
			Help: "Total number of portfolios created",
		}),
		portfoliosUpdated: promauto.With(registry).NewCounter(prometheus.CounterOpts{
			Name: "portfolios_updated_total",
			Help: "Total number of portfolios updated",
		}),
		portfoliosDeleted: promauto.With(registry).NewCounter(prometheus.CounterOpts{
			Name: "portfolios_deleted_total",
			Help: "Total number of portfolios deleted",
		}),
		usersLoggedIn: promauto.With(registry).NewCounter(prometheus.CounterOpts{
			Name: "users_logged_in_total",
			Help: "Total number of user logins",
		}),
		usersRegistered: promauto.With(registry).NewCounter(prometheus.CounterOpts{
			Name: "users_registered_total",
			Help: "Total number of user registrations",
		}),
	}
}

// Portfolio metrics
func (m *metricsCollector) IncrementPortfoliosCreated() {
	if m.portfoliosCreated != nil {
		m.portfoliosCreated.Inc()
	}
}

func (m *metricsCollector) IncrementPortfoliosUpdated() {
	if m.portfoliosUpdated != nil {
		m.portfoliosUpdated.Inc()
	}
}

func (m *metricsCollector) IncrementPortfoliosDeleted() {
	if m.portfoliosDeleted != nil {
		m.portfoliosDeleted.Inc()
	}
}

// User metrics
func (m *metricsCollector) IncrementUsersLoggedIn() {
	if m.usersLoggedIn != nil {
		m.usersLoggedIn.Inc()
	}
}

func (m *metricsCollector) IncrementUsersRegistered() {
	if m.usersRegistered != nil {
		m.usersRegistered.Inc()
	}
}
