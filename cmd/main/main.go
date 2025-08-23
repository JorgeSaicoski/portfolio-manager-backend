package main

import (
	"net/http"
	"os"

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

	// Keep the /healthy route for compatibility
	r.GET("/healthy", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"message": "Backend service is running",
			"service": "portfolio-backend",
		})
	})

	// API routes placeholder
	api := r.Group("/api")
	{
		api.GET("/status", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"service": "portfolio-backend",
				"version": "1.0.0",
				"status":  "running",
			})
		})
	}

	// Get port from environment or default to 8000 (matching Docker)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	// Start the server on the correct port
	r.Run(":" + port)
}
