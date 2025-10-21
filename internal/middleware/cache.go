package middleware

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

// cacheWriter wraps the response writer to capture the response body
type cacheWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *cacheWriter) Write(data []byte) (int, error) {
	w.body.Write(data)
	return w.ResponseWriter.Write(data)
}

func (w *cacheWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

// HTTPCache returns a middleware that implements HTTP caching with ETag support
func HTTPCache() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only cache GET and HEAD requests
		if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead {
			c.Next()
			return
		}

		// Skip caching for authenticated requests (to avoid caching user-specific data)
		if c.GetHeader("Authorization") != "" {
			// Set no-cache for authenticated requests
			c.Header("Cache-Control", "private, no-cache, no-store, must-revalidate")
			c.Header("Pragma", "no-cache")
			c.Header("Expires", "0")
			c.Next()
			return
		}

		// Create a response writer that captures the body
		writer := &cacheWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = writer

		c.Next()

		// Only process successful responses
		if c.Writer.Status() != http.StatusOK {
			return
		}

		// Generate ETag from response body
		etag := generateETag(writer.body.Bytes())

		// Check if client sent If-None-Match header
		clientETag := c.GetHeader("If-None-Match")
		if clientETag == etag {
			// Content hasn't changed, return 304
			c.AbortWithStatus(http.StatusNotModified)
			return
		}

		// Set ETag header
		c.Header("ETag", etag)

		// Set cache control headers for public resources
		// Adjust max-age based on your requirements
		c.Header("Cache-Control", "public, max-age=300") // 5 minutes
	}
}

// HTTPCacheWithTTL returns a middleware with custom cache TTL in seconds
func HTTPCacheWithTTL(maxAge int) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only cache GET and HEAD requests
		if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead {
			c.Next()
			return
		}

		// Skip caching for authenticated requests
		if c.GetHeader("Authorization") != "" {
			c.Header("Cache-Control", "private, no-cache, no-store, must-revalidate")
			c.Header("Pragma", "no-cache")
			c.Header("Expires", "0")
			c.Next()
			return
		}

		// Create a response writer that captures the body
		writer := &cacheWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = writer

		c.Next()

		// Only process successful responses
		if c.Writer.Status() != http.StatusOK {
			return
		}

		// Generate ETag from response body
		etag := generateETag(writer.body.Bytes())

		// Check if client sent If-None-Match header
		clientETag := c.GetHeader("If-None-Match")
		if clientETag == etag {
			// Content hasn't changed, return 304
			c.AbortWithStatus(http.StatusNotModified)
			return
		}

		// Set ETag header
		c.Header("ETag", etag)

		// Set cache control headers with custom max-age
		if maxAge > 0 {
			c.Header("Cache-Control", "public, max-age="+string(rune(maxAge)))
		}
	}
}

// NoCacheMiddleware sets headers to prevent caching
func NoCacheMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "private, no-cache, no-store, must-revalidate")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")
		c.Next()
	}
}

// generateETag creates an ETag from the response body using SHA-256
func generateETag(data []byte) string {
	if len(data) == 0 {
		return ""
	}
	hash := sha256.New()
	io.WriteString(hash, string(data))
	return `"` + hex.EncodeToString(hash.Sum(nil))[:16] + `"`
}
