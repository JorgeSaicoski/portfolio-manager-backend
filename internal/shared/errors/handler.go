package errors

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

// SafeErrorResponse represents a safe error response to send to clients
type SafeErrorResponse struct {
	Error     string `json:"error"`
	RequestID string `json:"request_id,omitempty"`
}

// ErrorHandler is a middleware that sanitizes errors before sending to clients
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if there are any errors
		if len(c.Errors) > 0 {
			err := c.Errors.Last()

			// Get request ID for tracing
			requestID, _ := c.Get("request_id")

			// Log the actual error with full context
			logger.WithFields(logrus.Fields{
				"request_id": requestID,
				"method":     c.Request.Method,
				"path":       c.Request.URL.Path,
				"ip":         c.ClientIP(),
				"error":      err.Error(),
				"type":       err.Type,
			}).Error("Request error")

			// Determine status code
			statusCode := c.Writer.Status()
			if statusCode == http.StatusOK {
				statusCode = http.StatusInternalServerError
			}

			// Return sanitized error to client
			c.JSON(statusCode, SafeErrorResponse{
				Error:     sanitizeErrorMessage(err.Error(), statusCode),
				RequestID: fmt.Sprintf("%v", requestID),
			})
		}
	}
}

// sanitizeErrorMessage removes sensitive information from error messages
func sanitizeErrorMessage(errMsg string, statusCode int) string {
	// In production, return generic messages for server errors
	if os.Getenv("GIN_MODE") == "release" {
		switch statusCode {
		case http.StatusInternalServerError:
			return "An internal server error occurred"
		case http.StatusBadGateway:
			return "Bad gateway"
		case http.StatusServiceUnavailable:
			return "Service temporarily unavailable"
		case http.StatusGatewayTimeout:
			return "Request timeout"
		default:
			// For client errors (4xx), we can be more specific
			if statusCode >= 400 && statusCode < 500 {
				return errMsg
			}
			return "An error occurred"
		}
	}

	// In development, return the actual error
	return errMsg
}

// NewBadRequestError creates a 400 Bad Request error
func NewBadRequestError(message string) error {
	return &HTTPError{
		StatusCode: http.StatusBadRequest,
		Message:    message,
	}
}

// NewUnauthorizedError creates a 401 Unauthorized error
func NewUnauthorizedError(message string) error {
	return &HTTPError{
		StatusCode: http.StatusUnauthorized,
		Message:    message,
	}
}

// NewForbiddenError creates a 403 Forbidden error
func NewForbiddenError(message string) error {
	return &HTTPError{
		StatusCode: http.StatusForbidden,
		Message:    message,
	}
}

// NewNotFoundError creates a 404 Not Found error
func NewNotFoundError(message string) error {
	return &HTTPError{
		StatusCode: http.StatusNotFound,
		Message:    message,
	}
}

// NewInternalServerError creates a 500 Internal Server Error
func NewInternalServerError(message string) error {
	return &HTTPError{
		StatusCode: http.StatusInternalServerError,
		Message:    message,
	}
}

// HTTPError represents an HTTP error with status code
type HTTPError struct {
	StatusCode int
	Message    string
}

func (e *HTTPError) Error() string {
	return e.Message
}
