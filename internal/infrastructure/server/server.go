package server

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/router"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/infrastructure/db"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/infrastructure/metrics"
	middleware2 "github.com/JorgeSaicoski/portfolio-manager/backend/internal/shared/middleware"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Server struct {
	port    string
	db      *gorm.DB
	engine  *gin.Engine
	server  *http.Server
	metrics *metrics.Collector
	logger  *logrus.Logger
	router  *router.Router
}

func NewServer(port string, database *db.Database, logger *logrus.Logger) *Server {
	metricsCollector := metrics.NewCollector()

	return &Server{
		port:    port,
		db:      database.DB,
		logger:  logger,
		metrics: metricsCollector,
		router:  router.NewRouter(database.DB, metricsCollector),
	}
}

func (s *Server) Start() error {
	s.setupEngine()
	s.setupMiddleware()
	s.setupRoutes()

	s.server = &http.Server{
		Addr:         ":" + s.port,
		Handler:      s.engine,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	s.logger.WithField("port", s.port).Info("Server starting")
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Server shutting down")
	return s.server.Shutdown(ctx)
}

func (s *Server) setupEngine() {
	if gin.Mode() == gin.ReleaseMode {
		gin.SetMode(gin.ReleaseMode)
	}
	s.engine = gin.New()
}

func (s *Server) setupMiddleware() {
	// Security middleware (applied first)
	s.engine.Use(middleware2.RequestID())        // Add request ID for tracing
	s.engine.Use(middleware2.PanicRecovery())    // Enhanced panic recovery
	s.engine.Use(middleware2.SecurityHeaders())  // Security headers
	s.engine.Use(middleware2.RequestSizeLimit()) // Request size limits
	s.engine.Use(middleware2.RateLimit())        // Rate limiting

	// Logging and metrics
	s.engine.Use(s.loggingMiddleware())
	s.engine.Use(s.metricsMiddleware())

	// CORS
	s.engine.Use(s.corsMiddleware())

	// Performance middleware
	// Skip compression for /metrics endpoint (Prometheus expects plain text)
	s.engine.Use(func(c *gin.Context) {
		if c.Request.URL.Path == "/metrics" {
			c.Next()
			return
		}
		middleware2.Compression()(c)
	})
	s.engine.Use(middleware2.HTTPCache())

	s.logger.Info("Security middleware initialized")
}

func (s *Server) setupRoutes() {
	// Health check
	s.engine.GET("/health", s.healthHandler)
	s.engine.GET("/ready", s.readinessHandler)
	s.engine.HEAD("/health", s.readinessHandler)

	// Metrics endpoint with basic auth protection
	metricsUser := os.Getenv("PROMETHEUS_AUTH_USER")
	metricsPassword := os.Getenv("PROMETHEUS_AUTH_PASSWORD")

	if metricsUser != "" && metricsPassword != "" {
		// Protected metrics endpoint
		authorized := s.engine.Group("/", gin.BasicAuth(gin.Accounts{
			metricsUser: metricsPassword,
		}))
		authorized.GET("/metrics", gin.WrapH(promhttp.Handler()))
		s.logger.Info("Metrics endpoint protected with basic authentication")
	} else {
		// Unprotected metrics (development only)
		s.engine.GET("/metrics", gin.WrapH(promhttp.Handler()))
		s.logger.Warn("Metrics endpoint is unprotected - set PROMETHEUS_AUTH_USER and PROMETHEUS_AUTH_PASSWORD")
	}

	// Static file serving for uploaded images
	s.router.RegisterStaticRoutes(s.engine)

	// API group
	api := s.engine.Group("/api")
	s.router.RegisterPortfolioRoutes(api)
	s.router.RegisterCategoryRoutes(api)
	s.router.RegisterProjectRoutes(api)
	s.router.RegisterSectionRoutes(api)
	s.router.RegisterSectionContentRoutes(api)
	s.router.RegisterImageRoutes(api)
}

func (s *Server) healthHandler(c *gin.Context) {
	// Check database connection
	dbStatus := "connected"
	sqlDB, err := s.db.DB()
	if err != nil || sqlDB.Ping() != nil {
		dbStatus = "disconnected"
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"message":   "Backend service is running",
		"service":   "portfolio-backend",
		"database":  dbStatus,
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

func (s *Server) readinessHandler(c *gin.Context) {
	// Check database connection
	sqlDB, err := s.db.DB()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not ready",
			"error":  "database connection failed",
		})
		return
	}

	if err := sqlDB.Ping(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not ready",
			"error":  "database ping failed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "ready",
		"database": "connected",
	})
}

func (s *Server) loggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		s.logger.WithFields(logrus.Fields{
			"status":     param.StatusCode,
			"method":     param.Method,
			"path":       param.Path,
			"ip":         param.ClientIP,
			"latency":    param.Latency,
			"user_agent": param.Request.UserAgent(),
		}).Info("HTTP Request")
		return ""
	})
}

func (s *Server) metricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start).Seconds()
		status := c.Writer.Status()

		s.metrics.IncrementHttpRequests(c.Request.Method, c.FullPath(), status)
		s.metrics.RecordHttpDuration(c.Request.Method, c.FullPath(), status, duration)
	}
}

func (s *Server) corsMiddleware() gin.HandlerFunc {
	// Get allowed origins from environment
	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
	if allowedOrigins == "" {
		// Default for development
		allowedOrigins = "*"
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// In production, check against allowed origins
		if allowedOrigins == "*" {
			c.Header("Access-Control-Allow-Origin", "*")
		} else {
			// Check if origin is in allowed list
			allowed := false
			for _, allowedOrigin := range strings.Split(allowedOrigins, ",") {
				if strings.TrimSpace(allowedOrigin) == origin {
					allowed = true
					break
				}
			}

			if allowed {
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Access-Control-Allow-Credentials", "true")
			} else {
				// Log unauthorized origin attempt
				s.logger.WithFields(logrus.Fields{
					"origin": origin,
					"ip":     c.ClientIP(),
				}).Warn("CORS request from unauthorized origin")
			}
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
		c.Header("Access-Control-Expose-Headers", "X-Request-ID")
		c.Header("Access-Control-Max-Age", "86400") // 24 hours

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
