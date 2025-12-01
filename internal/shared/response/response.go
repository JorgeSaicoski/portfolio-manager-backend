package response

import (
	"net/http"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/infrastructure/errorlog"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Error sends a standardized error response
func Error(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, gin.H{
		"error": message,
	})
}

// Success sends a standardized success response with data
func Success(c *gin.Context, statusCode int, data interface{}, message string) {
	c.JSON(statusCode, gin.H{
		"data":    data,
		"message": message,
	})
}

// SuccessWithKey sends a standardized success response with a custom data key
// Useful when you want the response to have a specific key like "project", "category", etc.
// Note: Now uses "data" as the standard key for consistency with REST API best practices
func SuccessWithKey(c *gin.Context, statusCode int, key string, data interface{}, message string) {
	c.JSON(statusCode, gin.H{
		"data":    data,
		"message": message,
	})
}

// SuccessWithPagination sends a standardized paginated success response
// Note: Uses "data" as the standard key for consistency with REST API best practices
func SuccessWithPagination(c *gin.Context, statusCode int, key string, data interface{}, page, limit int, total int64) {
	c.JSON(statusCode, gin.H{
		"data":    data,
		"page":    page,
		"limit":   limit,
		"total":   total,
		"message": "Success",
	})
}

// OK is a convenience wrapper for http.StatusOK success responses
func OK(c *gin.Context, key string, data interface{}, message string) {
	SuccessWithKey(c, http.StatusOK, key, data, message)
}

// Created is a convenience wrapper for http.StatusCreated success responses
func Created(c *gin.Context, key string, data interface{}, message string) {
	SuccessWithKey(c, http.StatusCreated, key, data, message)
}

// BadRequest is a convenience wrapper for http.StatusBadRequest error responses
func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, message)
}

// NotFound is a convenience wrapper for http.StatusNotFound error responses
func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, message)
}

// InternalError is a convenience wrapper for http.StatusInternalServerError error responses
func InternalError(c *gin.Context, message string) {
	// Store the error message in context for error logging middleware
	c.Set("error", message)
	Error(c, http.StatusInternalServerError, message)
}

// InternalErrorWithDetails logs detailed error information while showing user-friendly message
func InternalErrorWithDetails(c *gin.Context, userMessage string, detailedError error) {
	// Store detailed error for logging middleware
	if detailedError != nil {
		c.Error(detailedError) // Add to gin's error chain for middleware
		c.Set("error", detailedError.Error())
	}
	Error(c, http.StatusInternalServerError, userMessage)
}

// Forbidden is a convenience wrapper for http.StatusForbidden error responses
func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, message)
}

// ForbiddenWithDetails logs unauthorized access attempts with full context for audit trail
// This creates a detailed audit log entry showing who tried to access whose resource
func ForbiddenWithDetails(c *gin.Context, message string, details map[string]interface{}) {
	logger := errorlog.GetClientErrorLogger()

	// Get user information
	requestingUserID, _ := c.Get("userID")
	requestID, _ := c.Get("request_id")

	// Build audit log fields
	auditFields := logrus.Fields{
		"event_type":        "unauthorized_access_attempt",
		"requesting_user":   requestingUserID,
		"request_id":        requestID,
		"method":            c.Request.Method,
		"path":              c.Request.URL.Path,
		"ip":                c.ClientIP(),
		"user_agent":        c.Request.UserAgent(),
		"forbidden_message": message,
	}

	// Add all provided details to audit log
	for key, value := range details {
		auditFields[key] = value
	}

	// Log unauthorized access attempt
	logger.WithFields(auditFields).Warn("Unauthorized access attempt blocked")

	// Store error context for error logging middleware
	c.Set("forbidden_details", details)

	// Return 403 response
	Error(c, http.StatusForbidden, message)
}
