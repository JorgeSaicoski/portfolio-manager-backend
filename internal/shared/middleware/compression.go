package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

const (
	// MinCompressionSize is the minimum response size to trigger compression (1KB)
	MinCompressionSize = 1024
	// DefaultCompressionLevel is the default gzip compression level
	DefaultCompressionLevel


var gzipWriterPool = sync.Pool{
	New: func() interface{} {
		gz, _ := gzip.NewWriterLevel(io.Discard, DefaultCompressionLevel)
		return gz
	},
}

type gzipResponseWriter struct {
	gin.ResponseWriter
	writer *gzip.Writer
}

func (g *gzipResponseWriter) Write(data []byte) (int, error) {
	return g.writer.Write(data)
}

func (g *gzipResponseWriter) WriteString(s string) (int, error) {
	return g.writer.Write([]byte(s))
}

// Compression returns a middleware that compresses HTTP responses using gzip
func Compression() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip compression if client doesn't accept gzip
		if !shouldCompress(c.Request) {
			c.Next()
			return
		}

		// Get gzip writer from pool
		gz := gzipWriterPool.Get().(*gzip.Writer)
		defer gzipWriterPool.Put(gz)

		gz.Reset(c.Writer)
		defer gz.Close()

		// Set compression headers
		c.Header("Content-Encoding", "gzip")
		c.Header("Vary", "Accept-Encoding")

		// Wrap the response writer
		c.Writer = &gzipResponseWriter{
			ResponseWriter: c.Writer,
			writer:         gz,
		}

		c.Next()
	}
}

// shouldCompress checks if the request accepts gzip encoding
func shouldCompress(req *http.Request) bool {
	// Check if client accepts gzip
	if !strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") {
		return false
	}

	// Don't compress if response is already compressed
	if req.Header.Get("Content-Encoding") != "" {
		return false
	}

	// Don't compress websocket connections
	if strings.ToLower(req.Header.Get("Upgrade")) == "websocket" {
		return false
	}

	return true
}
