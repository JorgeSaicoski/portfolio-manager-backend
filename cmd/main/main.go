package main

import (
	"net/http"
	"os"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/db"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/router"
	"github.com/gin-gonic/gin"
)

func main() {
	// Create a Gin router with default middleware (logger and recovery)
	r := gin.Default()

	// Define the /health route (matching Docker health check)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"message": "Backend service is running",
			"service": "portfolio-backend",
		})
	})

	database := db.NewDatabase()
	database.Migrate()
	database.Initialize()

	router := router.NewRouter(database.DB)
	router.RegisterPortfolioRoutes(r)

	// Get port from environment or default to 8000 (matching Docker)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	// Start the server on the correct port
	r.Run(":" + port)
}
