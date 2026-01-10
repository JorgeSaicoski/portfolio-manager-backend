package main

import (
	"io"
	"os"
	"path/filepath"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/infrastructure/audit"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/infrastructure/db"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/infrastructure/di"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/infrastructure/errorlog"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/infrastructure/server"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

// mainV2 demonstrates the new clean architecture approach
// This can run alongside the old main.go
func mainV2() {
	logger := setupLogger()

	// Initialize audit loggers for CRUD operations
	audit.Initialize()

	// Initialize error loggers for 4xx and 5xx errors
	errorlog.Initialize()

	// Initialize database
	database := db.NewDatabase()
	err := database.Initialize()
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize database")
		return
	}

	err = database.Migrate()
	if err != nil {
		logger.WithError(err).Fatal("Failed to run migrations")
		return
	}

	// Create Prometheus registry
	promRegistry := prometheus.NewRegistry()

	// Initialize dependency injection container with clean architecture
	container, err := di.NewContainer(
		database.DB,
		audit.GetCreateLogger(),
		audit.GetUpdateLogger(),
		audit.GetDeleteLogger(),
		audit.GetErrorLogger(),
		promRegistry,
	)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize DI container")
		return
	}

	logger.Info("Clean architecture DI container initialized successfully")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	// Create server with both old and new routes
	srv := server.NewServer(port, database, logger)

	// Set the clean architecture container to enable /api/v2 routes
	srv.SetContainer(container)

	logger.Info("Starting server with clean architecture routes on /api/v2")
	if err := srv.Start(); err != nil {
		logger.WithError(err).Fatal("Failed to start server")
	}
}

func setupLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	level := os.Getenv("LOG_LEVEL")
	if level == "" {
		level = "info"
	}

	if parsedLevel, err := logrus.ParseLevel(level); err == nil {
		logger.SetLevel(parsedLevel)
	}

	// Create audit directory if it doesn't exist
	auditDir := filepath.Join(".", "audit")
	if err := os.MkdirAll(auditDir, 0755); err != nil {
		logger.WithError(err).Error("Failed to create audit directory")
	}

	// Create upload directories if they don't exist
	uploadDirs := []string{
		filepath.Join(".", "uploads", "images", "original"),
		filepath.Join(".", "uploads", "images", "thumbnail"),
	}
	for _, dir := range uploadDirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			logger.WithError(err).Errorf("Failed to create upload directory: %s", dir)
		} else {
			logger.Infof("Upload directory ready: %s", dir)
		}
	}

	// Setup audit log rotation
	{
		// Setup main audit log with rotation (errors and important events only)
		logFile := &lumberjack.Logger{
			Filename:   filepath.Join(auditDir, "audit.log"),
			MaxSize:    100, // megabytes
			MaxBackups: 30,  // keep 30 old log files
			MaxAge:     90,  // days
			Compress:   true,
		}

		// Write to both stdout and file
		multiWriter := io.MultiWriter(os.Stdout, logFile)
		logger.SetOutput(multiWriter)
	}

	return logger
}
