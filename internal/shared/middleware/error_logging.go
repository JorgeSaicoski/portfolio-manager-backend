package middleware

import (
	"fmt"
	"regexp"
	"runtime/debug"
	"strings"
	"time"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/infrastructure/errorlog"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ErrorLogging middleware captures and logs all HTTP errors (4xx and 5xx)
func ErrorLogging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Capture panics and convert to 500 errors
		defer func() {
			if err := recover(); err != nil {
				stackTrace := string(debug.Stack())
				file, line, function := extractActualErrorLocation(stackTrace)

				logServerError(c, 500, fmt.Sprintf("Panic: %v", err),
					file, line, function, stackTrace, time.Since(start))

				c.AbortWithStatusJSON(500, gin.H{
					"error": "An internal server error occurred",
				})
			}
		}()

		// Process request
		c.Next()

		// Log errors based on status code
		status := c.Writer.Status()
		if status >= 400 {
			latency := time.Since(start)
			stackTrace := string(debug.Stack())
			file, line, function := extractActualErrorLocation(stackTrace)

			// Get error message from context
			errorMsg := getErrorMessage(c)

			if status >= 500 {
				logServerError(c, status, errorMsg, file, line, function, stackTrace, latency)
			} else if status >= 400 {
				logClientError(c, status, errorMsg, file, line, function, latency)
			}
		}
	}
}

// extractActualErrorLocation parses stack trace to find the actual application code location
// (skips middleware, runtime, and third-party library frames)
func extractActualErrorLocation(stackTrace string) (string, int, string) {
	// Look for lines in the stack trace that contain /app/internal/application/handler/
	// This is where the actual business logic is
	lines := strings.Split(stackTrace, "\n")

	// Regex to match file:line patterns
	fileLineRegex := regexp.MustCompile(`(/app/internal/application/handler/\S+\.go):(\d+)`)

	// Also capture function names
	funcRegex := regexp.MustCompile(`^([^/\s]+\.[^(]+)`)

	var lastFunction string
	for i, line := range lines {
		// Check if this is a function name line
		if funcMatch := funcRegex.FindStringSubmatch(strings.TrimSpace(line)); funcMatch != nil {
			lastFunction = funcMatch[1]
		}

		// Check for file:line in handler code
		if matches := fileLineRegex.FindStringSubmatch(line); matches != nil {
			file := matches[1]
			lineStr := matches[2]

			// Parse line number
			var lineNum int
			fmt.Sscanf(lineStr, "%d", &lineNum)

			// Extract just the handler name from function
			handlerName := extractHandlerName(lastFunction)

			return file, lineNum, handlerName
		}

		// If we find handler code in the previous line, check current line for file:line
		if i > 0 && strings.Contains(lines[i-1], "handler") {
			if matches := fileLineRegex.FindStringSubmatch(line); matches != nil {
				file := matches[1]
				lineStr := matches[2]
				var lineNum int
				fmt.Sscanf(lineStr, "%d", &lineNum)
				handlerName := extractHandlerName(lastFunction)
				return file, lineNum, handlerName
			}
		}
	}

	// Fallback: look for any /app/ code
	appCodeRegex := regexp.MustCompile(`(/app/\S+\.go):(\d+)`)
	for _, line := range lines {
		if matches := appCodeRegex.FindStringSubmatch(line); matches != nil {
			file := matches[1]
			lineStr := matches[2]
			var lineNum int
			fmt.Sscanf(lineStr, "%d", &lineNum)
			return file, lineNum, "unknown"
		}
	}

	// Final fallback
	return "unknown", 0, "unknown"
}

// extractHandlerName extracts a clean handler name from the full function path
func extractHandlerName(fullFunc string) string {
	// Convert something like "github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/handler.(*SectionHandler).GetByPortfolio"
	// To: "section.GetByPortfolio"

	parts := strings.Split(fullFunc, ".")
	if len(parts) >= 2 {
		// Get last two parts
		handlerType := parts[len(parts)-2]
		method := parts[len(parts)-1]

		// Clean handler type (remove (*Handler) wrapper)
		handlerType = strings.TrimPrefix(handlerType, "(*")
		handlerType = strings.TrimSuffix(handlerType, "Handler)")
		handlerType = strings.ToLower(handlerType)

		return fmt.Sprintf("%s.%s", handlerType, method)
	}

	return fullFunc
}

// logServerError logs 5xx errors with full stack traces
func logServerError(c *gin.Context, status int, errorMsg, file string, line int, function string, stackTrace string, latency time.Duration) {
	logger := errorlog.GetServerErrorLogger()

	requestID, _ := c.Get("request_id")
	userID, _ := c.Get("userID")

	// Get query parameters
	queryParams := make(map[string]string)
	for key, values := range c.Request.URL.Query() {
		if len(values) > 0 {
			queryParams[key] = values[0]
		}
	}

	fields := logrus.Fields{
		"request_id":  requestID,
		"method":      c.Request.Method,
		"path":        c.Request.URL.Path,
		"status":      status,
		"error":       errorMsg,
		"error_type":  "ServerError",
		"ip":          c.ClientIP(),
		"user_agent":  c.Request.UserAgent(),
		"user_id":     userID,
		"file":        file,
		"line":        line,
		"handler":     function,
		"stack_trace": stackTrace,
		"latency_ms":  latency.Milliseconds(),
	}

	// Add query params if present
	if len(queryParams) > 0 {
		fields["query_params"] = queryParams
	}

	logger.WithFields(fields).Error("Server error occurred")
}

// logClientError logs 4xx errors
func logClientError(c *gin.Context, status int, errorMsg, file string, line int, function string, latency time.Duration) {
	logger := errorlog.GetClientErrorLogger()

	requestID, _ := c.Get("request_id")
	userID, _ := c.Get("userID")

	// Get query parameters
	queryParams := make(map[string]string)
	for key, values := range c.Request.URL.Query() {
		if len(values) > 0 {
			queryParams[key] = values[0]
		}
	}

	fields := logrus.Fields{
		"request_id": requestID,
		"method":     c.Request.Method,
		"path":       c.Request.URL.Path,
		"status":     status,
		"error":      errorMsg,
		"error_type": "ClientError",
		"ip":         c.ClientIP(),
		"user_agent": c.Request.UserAgent(),
		"user_id":    userID,
		"file":       file,
		"line":       line,
		"handler":    function,
		"latency_ms": latency.Milliseconds(),
	}

	// Add query params if present
	if len(queryParams) > 0 {
		fields["query_params"] = queryParams
	}

	logger.WithFields(fields).Warn("Client error occurred")
}

// getErrorMessage extracts error message from gin context or response
func getErrorMessage(c *gin.Context) string {
	// Priority 1: Check Gin's error context
	if len(c.Errors) > 0 {
		return c.Errors.Last().Error()
	}

	// Priority 2: Check if there's an error in the response (for JSON responses)
	// This requires checking what was written to the response body
	// For now, we'll use a generic message based on status code
	status := c.Writer.Status()

	// Common HTTP status messages
	switch status {
	case 400:
		return "Bad Request"
	case 401:
		return "Unauthorized"
	case 403:
		return "Forbidden"
	case 404:
		return "Not Found"
	case 422:
		return "Unprocessable Entity"
	case 500:
		return "Internal Server Error"
	case 502:
		return "Bad Gateway"
	case 503:
		return "Service Unavailable"
	case 504:
		return "Gateway Timeout"
	default:
		return fmt.Sprintf("HTTP %d error", status)
	}
}
