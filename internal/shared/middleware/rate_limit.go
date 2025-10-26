package middleware

import (
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter manages rate limiting per IP address
type RateLimiter struct {
	visitors map[string]*Visitor
	mu       sync.RWMutex
	rate     int           // requests allowed
	window   time.Duration // time window
}

// Visitor tracks request count for a single IP
type Visitor struct {
	lastSeen time.Time
	count    int
	mu       sync.Mutex
}

var rateLimiter *RateLimiter
var rateLimiterOnce sync.Once

// initRateLimiter initializes the global rate limiter
func initRateLimiter() *RateLimiter {
	rateLimiterOnce.Do(func() {
		// Get configuration from environment or use defaults
		rate := 100                // default: 100 requests per window
		window := 60 * time.Second // default: 60 seconds

		if rateStr := os.Getenv("RATE_LIMIT_REQUESTS"); rateStr != "" {
			if r, err := strconv.Atoi(rateStr); err == nil {
				rate = r
			}
		}

		if windowStr := os.Getenv("RATE_LIMIT_WINDOW"); windowStr != "" {
			if w, err := strconv.Atoi(windowStr); err == nil {
				window = time.Duration(w) * time.Second
			}
		}

		rateLimiter = &RateLimiter{
			visitors: make(map[string]*Visitor),
			rate:     rate,
			window:   window,
		}

		// Start cleanup goroutine to remove old visitors
		go rateLimiter.cleanupVisitors()
	})

	return rateLimiter
}

// getVisitor retrieves or creates a visitor for an IP
func (rl *RateLimiter) getVisitor(ip string) *Visitor {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		v = &Visitor{
			lastSeen: time.Now(),
			count:    0,
		}
		rl.visitors[ip] = v
	}

	return v
}

// isAllowed checks if a request from this IP is allowed
func (rl *RateLimiter) isAllowed(ip string) bool {
	visitor := rl.getVisitor(ip)

	visitor.mu.Lock()
	defer visitor.mu.Unlock()

	now := time.Now()

	// Reset count if window has passed
	if now.Sub(visitor.lastSeen) > rl.window {
		visitor.count = 0
		visitor.lastSeen = now
	}

	// Check if under rate limit
	if visitor.count >= rl.rate {
		return false
	}

	visitor.count++
	visitor.lastSeen = now
	return true
}

// cleanupVisitors periodically removes old visitors to prevent memory leak
func (rl *RateLimiter) cleanupVisitors() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, visitor := range rl.visitors {
			visitor.mu.Lock()
			if now.Sub(visitor.lastSeen) > 10*time.Minute {
				delete(rl.visitors, ip)
			}
			visitor.mu.Unlock()
		}
		rl.mu.Unlock()
	}
}

// RateLimit is the middleware function for rate limiting
func RateLimit() gin.HandlerFunc {
	limiter := initRateLimiter()

	return func(c *gin.Context) {
		// Get client IP
		ip := c.ClientIP()

		// Check if request is allowed
		if !limiter.isAllowed(ip) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded. Please try again later.",
			})
			return
		}

		c.Next()
	}
}
