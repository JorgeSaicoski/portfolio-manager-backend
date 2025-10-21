package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/models"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates tokens by forwarding to auth service
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Check Bearer token format
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			c.Abort()
			return
		}

		var user *models.User
		var err error

		// Check if we're in testing mode
		if os.Getenv("TESTING_MODE") == "true" {
			// In testing mode, use a simple test user from the token
			user = getTestUser(tokenParts[1])
		} else {
			// Forward request to auth service to validate token
			user, err = validateTokenWithAuthService(authHeader)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
				c.Abort()
				return
			}
		}

		// Set user data in context for handlers to use
		// Convert user ID to string for consistent usage in handlers
		c.Set("userID", fmt.Sprintf("%d", user.ID))
		c.Set("user", user)
		c.Next()
	}
}

// getTestUser creates a test user for testing purposes
// Token format: any value returns the default test user
func getTestUser(token string) *models.User {
	return &models.User{
		ID:       123, // Test user ID as uint
		Username: "testuser",
		Email:    "test@example.com",
	}
}

// validateTokenWithAuthService calls the auth service to validate token and get user
func validateTokenWithAuthService(authHeader string) (*models.User, error) {
	authServiceURL := os.Getenv("AUTH_SERVICE_URL")
	if authServiceURL == "" {
		authServiceURL = "http://localhost:8080" // fallback for development
	}

	// Create request to auth service profile endpoint
	req, err := http.NewRequest("GET", authServiceURL+"/api/profile", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Forward the Authorization header
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", "application/json")

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call auth service: %w", err)
	}
	defer resp.Body.Close()

	// Check if auth service returned success
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("auth service returned status: %d", resp.StatusCode)
	}

	// Decode user data from response
	var user models.User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to decode user data: %w", err)
	}

	return &user, nil
}
