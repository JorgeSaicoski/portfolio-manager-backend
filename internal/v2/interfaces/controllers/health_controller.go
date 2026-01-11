package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/interfaces/dto/response"
)

// HealthController handles health check endpoints
// NOTE: These are infrastructure concerns, not business logic
// Therefore, no use cases needed - direct database access is appropriate
type HealthController struct {
	db *gorm.DB
}

// NewHealthController creates a new health controller instance
func NewHealthController(db *gorm.DB) *HealthController {
	return &HealthController{
		db: db,
	}
}

// Health handles GET /health
// Basic health check: returns 200 if service is up and database is connected
func (ctrl *HealthController) Health(c *gin.Context) {
	// Check database connection
	dbStatus := "connected"
	sqlDB, err := ctrl.db.DB()
	if err != nil || sqlDB.Ping() != nil {
		dbStatus = "disconnected"

		// Return 503 Service Unavailable if database is down
		c.JSON(http.StatusServiceUnavailable, response.HealthResponse{
			Status:    "unhealthy",
			Database:  dbStatus,
			Timestamp: time.Now(),
		})
		return
	}

	// Return 200 OK if everything is healthy
	c.JSON(http.StatusOK, response.HealthResponse{
		Status:    "healthy",
		Database:  dbStatus,
		Timestamp: time.Now(),
	})
}

// DatabaseHealth handles GET /health/db
// Detailed database health check with connection pool stats
func (ctrl *HealthController) DatabaseHealth(c *gin.Context) {
	// Get database connection
	sqlDB, err := ctrl.db.DB()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, response.DatabaseHealthResponse{
			Status: "unhealthy",
			Database: response.DatabaseStatus{
				Connected: false,
			},
			Timestamp: time.Now(),
		})
		return
	}

	// Ping database
	if err := sqlDB.Ping(); err != nil {
		c.JSON(http.StatusServiceUnavailable, response.DatabaseHealthResponse{
			Status: "unhealthy",
			Database: response.DatabaseStatus{
				Connected: false,
			},
			Timestamp: time.Now(),
		})
		return
	}

	// Get connection pool statistics
	stats := sqlDB.Stats()

	// Return detailed health information
	c.JSON(http.StatusOK, response.DatabaseHealthResponse{
		Status: "healthy",
		Database: response.DatabaseStatus{
			Connected:         true,
			MaxOpenConns:      stats.MaxOpenConnections,
			OpenConnections:   stats.OpenConnections,
			InUse:             stats.InUse,
			Idle:              stats.Idle,
			WaitCount:         stats.WaitCount,
			WaitDuration:      stats.WaitDuration.String(),
			MaxIdleClosed:     stats.MaxIdleClosed,
			MaxLifetimeClosed: stats.MaxLifetimeClosed,
		},
		Timestamp: time.Now(),
	})
}
