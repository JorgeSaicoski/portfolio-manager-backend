package main

import (
	"io"
	"os"
	"path/filepath"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/infrastructure/db"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/infrastructure/server"
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

func main() {
	logger := setupLogger()

	database := db.NewDatabase()
	err := database.Initialize()
	if err != nil {
		return
	}

	err = database.Migrate()
	if err != nil {
		return
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	srv := server.NewServer(port, database, logger)

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
	} else {
		// Setup file logging with rotation
		logFile := &lumberjack.Logger{
			Filename:   filepath.Join(auditDir, "audit.log"),
			MaxSize:    10, // megabytes
			MaxBackups: 30, // keep 30 old log files
			MaxAge:     90, // days
			Compress:   true,
		}

		// Write to both stdout and file
		multiWriter := io.MultiWriter(os.Stdout, logFile)
		logger.SetOutput(multiWriter)
	}

	return logger
}
