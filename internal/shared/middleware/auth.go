package middleware

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/infrastructure/audit"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var (
	oidcVerifier    *oidc.IDTokenVerifier
	oidcProvider    *oidc.Provider
	oidcInitOnce    sync.Once
	oidcInitError   error
	authentikIssuer string
	logger          = logrus.New()
)

// InitOIDC initializes the OIDC provider and verifier
func InitOIDC() error {
	oidcInitOnce.Do(func() {
		// Use AUTHENTIK_URL for OIDC provider initialization (container-to-container communication)
		// Fall back to AUTHENTIK_ISSUER for backward compatibility
		authentikURL := os.Getenv("AUTHENTIK_URL")
		if authentikURL == "" {
			authentikURL = os.Getenv("AUTHENTIK_ISSUER")
		}

		if authentikURL == "" {
			oidcInitError = fmt.Errorf("AUTHENTIK_URL or AUTHENTIK_ISSUER environment variable not set")
			return
		}

		// Build the full issuer URL if needed
		if !strings.HasSuffix(authentikURL, "/") {
			authentikURL += "/"
		}

		// If URL doesn't already include the application path, append it
		if !strings.Contains(authentikURL, "/application/o/") {
			authentikIssuer = authentikURL + "application/o/portfolio-manager/"
		} else {
			authentikIssuer = authentikURL
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Get the public issuer URL (what tokens will contain)
		// This is typically the browser-accessible URL like http://localhost:9000
		// while authentikIssuer is the internal container URL like http://portfolio-authentik-server:9000
		publicIssuer := os.Getenv("AUTHENTIK_ISSUER")
		if publicIssuer != "" && publicIssuer != authentikIssuer {
			// Use InsecureIssuerURLContext to handle issuer mismatch in containerized environments
			// This allows backend to discover OIDC config via internal URL (authentikIssuer)
			// but validate tokens with public issuer URL (publicIssuer)
			logger.WithFields(logrus.Fields{
				"discovery_url": authentikIssuer,
				"token_issuer":  publicIssuer,
			}).Info("Using separate discovery and token issuer URLs for containerized environment")
			ctx = oidc.InsecureIssuerURLContext(ctx, publicIssuer)
		}

		// Initialize OIDC provider
		provider, err := oidc.NewProvider(ctx, authentikIssuer)
		if err != nil {
			oidcInitError = fmt.Errorf("failed to create OIDC provider: %w", err)
			return
		}
		oidcProvider = provider

		// Create ID token verifier
		oidcVerifier = provider.Verifier(&oidc.Config{
			SkipClientIDCheck: true, // We'll accept tokens from any client
		})

		logger.WithFields(logrus.Fields{
			"issuer": authentikIssuer,
		}).Info("OIDC provider initialized successfully")
	})

	return oidcInitError
}

// User represents the authenticated user from Authentik
type User struct {
	Sub               string `json:"sub"`
	Email             string `json:"email"`
	EmailVerified     bool   `json:"email_verified"`
	Name              string `json:"name"`
	PreferredUsername string `json:"preferred_username"`
	GivenName         string `json:"given_name"`
	FamilyName        string `json:"family_name"`
	Nickname          string `json:"nickname"`
}

// AuthMiddleware validates OAuth2/OIDC tokens from Authentik
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			audit.GetErrorLogger().WithFields(logrus.Fields{
				"operation": "AUTH_MIDDLEWARE_NO_HEADER",
				"where":     "backend/internal/shared/middleware/auth.go",
				"function":  "AuthMiddleware",
				"ip":        c.ClientIP(),
				"path":      c.Request.URL.Path,
			}).Warn("Authorization header required")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Check Bearer token format
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			audit.GetErrorLogger().WithFields(logrus.Fields{
				"operation": "AUTH_MIDDLEWARE_INVALID_FORMAT",
				"where":     "backend/internal/shared/middleware/auth.go",
				"function":  "AuthMiddleware",
				"ip":        c.ClientIP(),
				"path":      c.Request.URL.Path,
			}).Warn("Invalid authorization format")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			c.Abort()
			return
		}

		accessToken := tokenParts[1]

		// Check if we're in testing mode
		if os.Getenv("TESTING_MODE") == "true" {
			// In testing mode, use a simple test user
			user := getTestUser()
			c.Set("userID", user.Sub)
			c.Set("user", user)
			c.Next()
			return
		}

		// Ensure OIDC is initialized
		if err := InitOIDC(); err != nil {
			logger.WithError(err).Error("OIDC not initialized")
			audit.GetErrorLogger().WithFields(logrus.Fields{
				"operation": "AUTH_MIDDLEWARE_OIDC_INIT_FAILED",
				"where":     "backend/internal/shared/middleware/auth.go",
				"function":  "AuthMiddleware",
				"ip":        c.ClientIP(),
				"path":      c.Request.URL.Path,
				"error":     err.Error(),
			}).Error("OIDC not initialized")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication service unavailable"})
			c.Abort()
			return
		}

		// Verify the access token as an ID token
		// Note: Authentik's access tokens are JWTs that can be verified like ID tokens
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		idToken, err := oidcVerifier.Verify(ctx, accessToken)
		if err != nil {
			logger.WithError(err).WithField("token_prefix", accessToken[:smaller(20, len(accessToken))]).Warn("Token verification failed")
			audit.GetErrorLogger().WithFields(logrus.Fields{
				"operation":    "AUTH_MIDDLEWARE_TOKEN_VERIFICATION_FAILED",
				"where":        "backend/internal/shared/middleware/auth.go",
				"function":     "AuthMiddleware",
				"ip":           c.ClientIP(),
				"path":         c.Request.URL.Path,
				"token_prefix": accessToken[:smaller(20, len(accessToken))],
				"error":        err.Error(),
			}).Warn("Token verification failed")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Extract claims
		var user User
		if err := idToken.Claims(&user); err != nil {
			logger.WithError(err).Error("Failed to extract claims from token")
			audit.GetErrorLogger().WithFields(logrus.Fields{
				"operation": "AUTH_MIDDLEWARE_CLAIMS_EXTRACTION_FAILED",
				"where":     "backend/internal/shared/middleware/auth.go",
				"function":  "AuthMiddleware",
				"ip":        c.ClientIP(),
				"path":      c.Request.URL.Path,
				"error":     err.Error(),
			}).Error("Failed to extract claims from token")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		// Set user data in context for handlers to use
		c.Set("userID", user.Sub) // Use Authentik's subject (user ID) as userID
		c.Set("user", user)
		c.Set("email", user.Email)
		c.Set("username", user.PreferredUsername)

		logger.WithFields(logrus.Fields{
			"user_id":  user.Sub,
			"email":    user.Email,
			"username": user.PreferredUsername,
		}).Debug("User authenticated successfully")

		c.Next()
	}
}

// getTestUser creates a test user for testing purposes
func getTestUser() User {
	return User{
		Sub:               "test-user-123",
		Email:             "test@example.com",
		EmailVerified:     true,
		Name:              "Test User",
		PreferredUsername: "testuser",
		GivenName:         "Test",
		FamilyName:        "User",
		Nickname:          "testuser",
	}
}

// Helper function for smaller
func smaller(a, b int) int {
	if a < b {
		return a
	}
	return b
}
