package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
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
func SuccessWithKey(c *gin.Context, statusCode int, key string, data interface{}, message string) {
	c.JSON(statusCode, gin.H{
		key:       data,
		"message": message,
	})
}

// SuccessWithPagination sends a standardized paginated success response
func SuccessWithPagination(c *gin.Context, statusCode int, key string, data interface{}, page, limit int) {
	c.JSON(statusCode, gin.H{
		key:       data,
		"page":    page,
		"limit":   limit,
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
	Error(c, http.StatusInternalServerError, message)
}

// Forbidden is a convenience wrapper for http.StatusForbidden error responses
func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, message)
}
