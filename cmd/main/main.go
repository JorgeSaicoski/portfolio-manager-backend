package main

import (
	"os"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/db"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/server"
	"github.com/sirupsen/logrus"
)

func main() {
	logger := setupLogger()

	database := db.NewDatabase()
	database.Initialize()
	database.Migrate()

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

	return logger
}
