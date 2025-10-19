package server

import (
	"context"
	"net/http"
	"time"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/db"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/metrics"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/router"
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
	s.engine.Use(s.loggingMiddleware())
	s.engine.Use(gin.Recovery())
	s.engine.Use(s.metricsMiddleware())
	s.engine.Use(s.corsMiddleware())
}

func (s *Server) setupRoutes() {
	// Health check
	s.engine.GET("/health", s.healthHandler)
	s.engine.GET("/ready", s.readinessHandler)
	s.engine.GET("/metrics", gin.WrapH(promhttp.Handler()))
	s.engine.HEAD("/health", s.readinessHandler)

	// API group
	api := s.engine.Group("/api")
	s.router.RegisterPortfolioRoutes(api)
	s.router.RegisterCategoryRoutes(api)
	s.router.RegisterProjectRoutes(api)
	s.router.RegisterSectionRoutes(api)
}

func (s *Server) healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"message": "Backend service is running",
		"service": "portfolio-backend",
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
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
