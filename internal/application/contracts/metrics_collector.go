package contracts

// MetricsCollector defines the interface for metrics collection
// This is a contract in the application layer that the infrastructure layer must implement
type MetricsCollector interface {
	// Portfolio metrics
	IncrementPortfoliosCreated()
	IncrementPortfoliosUpdated()
	IncrementPortfoliosDeleted()

	// Category metrics (for future use)
	IncrementCategoriesCreated()
	IncrementCategoriesUpdated()
	IncrementCategoriesDeleted()

	// Section metrics (for future use)
	IncrementSectionsCreated()
	IncrementSectionsUpdated()
	IncrementSectionsDeleted()

	// User metrics (for future use)
	IncrementUsersCreated()
	IncrementUsersUpdated()
	IncrementUsersDeleted()

	// Authentication metrics
	IncrementAuthAttempts(authType, status string)
	IncrementJwtTokens(tokenType string)

	// HTTP metrics
	RecordHttpDuration(method, path string, status int, duration float64)
	IncrementHttpRequests(method, path string, status int)
}
