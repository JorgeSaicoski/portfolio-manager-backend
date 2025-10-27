package middleware

import (
	"html"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

// SanitizeInput is a middleware that sanitizes user input to prevent XSS attacks
func SanitizeInput() gin.HandlerFunc {
	return func(c *gin.Context) {
		// For JSON requests, we'll sanitize after binding
		// This middleware primarily serves as a reminder to sanitize
		// Actual sanitization should happen in handlers/DTOs

		c.Next()
	}
}

// SanitizeString removes potentially dangerous HTML/script tags and escapes HTML entities
func SanitizeString(input string) string {
	// Trim whitespace
	sanitized := strings.TrimSpace(input)

	// Remove script tags and their content
	scriptRegex := regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`)
	sanitized = scriptRegex.ReplaceAllString(sanitized, "")

	// Remove event handlers (onclick, onerror, etc.)
	eventRegex := regexp.MustCompile(`(?i)\s*on\w+\s*=\s*["'][^"']*["']`)
	sanitized = eventRegex.ReplaceAllString(sanitized, "")

	// Remove javascript: protocol
	jsProtocolRegex := regexp.MustCompile(`(?i)javascript:`)
	sanitized = jsProtocolRegex.ReplaceAllString(sanitized, "")

	// Remove data: protocol (can be used for XSS)
	dataProtocolRegex := regexp.MustCompile(`(?i)data:text/html`)
	sanitized = dataProtocolRegex.ReplaceAllString(sanitized, "")

	// Escape HTML entities
	sanitized = html.EscapeString(sanitized)

	return sanitized
}

// SanitizeHTML allows safe HTML tags while removing dangerous ones
func SanitizeHTML(input string) string {
	sanitized := input

	// Remove script tags
	scriptRegex := regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`)
	sanitized = scriptRegex.ReplaceAllString(sanitized, "")

	// Remove style tags
	styleRegex := regexp.MustCompile(`(?i)<style[^>]*>.*?</style>`)
	sanitized = styleRegex.ReplaceAllString(sanitized, "")

	// Remove event handlers
	eventRegex := regexp.MustCompile(`(?i)\s*on\w+\s*=\s*["'][^"']*["']`)
	sanitized = eventRegex.ReplaceAllString(sanitized, "")

	// Remove javascript: and data: protocols
	jsProtocolRegex := regexp.MustCompile(`(?i)(javascript|data):`)
	sanitized = jsProtocolRegex.ReplaceAllString(sanitized, "")

	// Remove iframe, object, embed tags
	dangerousTagsRegex := regexp.MustCompile(`(?i)<(iframe|object|embed|applet|meta|link)[^>]*>.*?</\1>`)
	sanitized = dangerousTagsRegex.ReplaceAllString(sanitized, "")

	return sanitized
}

// ValidateEmail checks if an email address is valid
func ValidateEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// ValidateURL checks if a URL is valid and safe
func ValidateURL(url string) bool {
	// Basic URL validation
	urlRegex := regexp.MustCompile(`^https?://[a-zA-Z0-9\-._~:/?#\[\]@!$&'()*+,;=]+$`)
	if !urlRegex.MatchString(url) {
		return false
	}

	// Ensure it doesn't contain javascript: or data: protocols
	if strings.Contains(strings.ToLower(url), "javascript:") ||
		strings.Contains(strings.ToLower(url), "data:") {
		return false
	}

	return true
}

// SanitizeFilename removes dangerous characters from filenames
func SanitizeFilename(filename string) string {
	// Remove path traversal attempts
	filename = strings.ReplaceAll(filename, "../", "")
	filename = strings.ReplaceAll(filename, "..\\", "")

	// Remove null bytes
	filename = strings.ReplaceAll(filename, "\x00", "")

	// Allow only alphanumeric, dash, underscore, and dot
	filenameRegex := regexp.MustCompile(`[^a-zA-Z0-9._\-]`)
	filename = filenameRegex.ReplaceAllString(filename, "_")

	// Limit length
	if len(filename) > 255 {
		filename = filename[:255]
	}

	return filename
}

// ValidateFileExtension checks if a file extension is in the allowed list
func ValidateFileExtension(filename string, allowedExtensions []string) bool {
	// Get file extension
	parts := strings.Split(filename, ".")
	if len(parts) < 2 {
		return false
	}

	ext := strings.ToLower(parts[len(parts)-1])

	// Check if extension is allowed
	for _, allowedExt := range allowedExtensions {
		if ext == strings.ToLower(allowedExt) {
			return true
		}
	}

	return false
}
