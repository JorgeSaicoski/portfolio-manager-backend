package middleware

import (
	"net/http"
	"strings"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/contracts"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware handles authentication for v2 API endpoints
type AuthMiddleware struct {
	authProvider contracts.AuthProvider
}

// NewAuthMiddleware creates a new authentication middleware instance
func NewAuthMiddleware(authProvider contracts.AuthProvider) *AuthMiddleware {
	return &AuthMiddleware{
		authProvider: authProvider,
	}
}

// Authenticate returns a Gin middleware that validates the access token
func (m *AuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			c.Abort()
			return
		}

		// Validate Bearer token format
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format, expected 'Bearer <token>'"})
			c.Abort()
			return
		}

		accessToken := parts[1]
		if accessToken == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing access token"})
			c.Abort()
			return
		}

		// Validate token using AuthProvider contract
		userID, err := m.authProvider.ValidateToken(accessToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}

		// Set userID in context for downstream handlers
		c.Set("userID", userID)

		// Continue to next handler
		c.Next()
	}
}
