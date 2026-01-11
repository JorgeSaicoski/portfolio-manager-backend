package response

import "time"

// HealthResponse represents basic health check response
type HealthResponse struct {
	Status    string    `json:"status"`   // "healthy" or "unhealthy"
	Database  string    `json:"database"` // "connected" or "disconnected"
	Timestamp time.Time `json:"timestamp"`
}

// DatabaseHealthResponse represents detailed database health with connection pool stats
type DatabaseHealthResponse struct {
	Status    string         `json:"status"` // "healthy" or "unhealthy"
	Database  DatabaseStatus `json:"database"`
	Timestamp time.Time      `json:"timestamp"`
}

// DatabaseStatus represents database connection details
type DatabaseStatus struct {
	Connected         bool   `json:"connected"`
	MaxOpenConns      int    `json:"max_open_connections"`
	OpenConnections   int    `json:"open_connections"`
	InUse             int    `json:"in_use"`
	Idle              int    `json:"idle"`
	WaitCount         int64  `json:"wait_count"`
	WaitDuration      string `json:"wait_duration"`
	MaxIdleClosed     int64  `json:"max_idle_closed"`
	MaxLifetimeClosed int64  `json:"max_lifetime_closed"`
}
