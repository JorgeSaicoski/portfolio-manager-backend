package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/infrastructure/audit"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// SecurityHeaders adds security-related HTTP headers to all responses
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent clickjacking attacks
		c.Header("X-Frame-Options", "DENY")

		// Prevent MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")

		// Enable XSS protection (legacy browsers)
		c.Header("X-XSS-Protection", "1; mode=block")

		// Enforce HTTPS in production (only if not in development)
		if os.Getenv("GIN_MODE") == "release" {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		// Content Security Policy - restrict resource loading
		csp := "default-src 'self'; " +
			"script-src 'self' 'unsafe-inline'; " + // Allow inline scripts for now (can be tightened)
			"style-src 'self' 'unsafe-inline'; " + // Allow inline styles
			"img-src 'self' data: https:; " + // Allow images from self, data URIs, and HTTPS
			"font-src 'self' data:; " +
			"connect-src 'self'; " +
			"frame-ancestors 'none'"
		c.Header("Content-Security-Policy", csp)

		// Referrer Policy - control referrer information
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions Policy - restrict browser features
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		c.Next()
	}
}

// RequestSizeLimit restricts the maximum size of request bodies
func RequestSizeLimit() gin.HandlerFunc {
	// Get max size from env or use default (10MB)
	maxSizeStr := os.Getenv("MAX_REQUEST_SIZE")
	maxSize := int64(10 * 1024 * 1024) // 10MB default

	if maxSizeStr != "" {
		if size, err := strconv.ParseInt(maxSizeStr, 10, 64); err == nil {
			maxSize = size
		}
	}

	return func(c *gin.Context) {
		// Limit request body size
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)

		c.Next()

		// Check if body was too large
		if c.Writer.Status() == http.StatusRequestEntityTooLarge {
			audit.GetErrorLogger().WithFields(logrus.Fields{
				"operation": "REQUEST_SIZE_LIMIT_EXCEEDED",
				"where":     "backend/internal/shared/middleware/security.go",
				"function":  "RequestSizeLimit",
				"ip":        c.ClientIP(),
				"path":      c.Request.URL.Path,
				"method":    c.Request.Method,
				"max_size":  maxSize,
			}).Warn("Request body too large")
			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{
				"error": fmt.Sprintf("Request body too large. Maximum size is %d bytes", maxSize),
			})
		}
	}
}

// RequestID adds a unique request ID to each request for tracing
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if request already has an ID (from load balancer, etc.)
		requestID := c.GetHeader("X-Request-ID")

		// Generate one if not present
		if requestID == "" {
			requestID = generateRequestID()
		}

		// Set it in context and response header
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}

// generateRequestID creates a simple unique request ID
// In production, consider using UUID or similar
func generateRequestID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// PanicRecovery is an enhanced panic recovery middleware
func PanicRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Return safe error to client (don't expose panic details)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": "An internal server error occurred",
				})
			}
		}()

		c.Next()
	}
}
