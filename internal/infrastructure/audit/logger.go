package audit

import (
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/sirupsen/logrus"
)

var (
	CreateLogger     *logrus.Logger
	UpdateLogger     *logrus.Logger
	DeleteLogger     *logrus.Logger
	BadRequestLogger *logrus.Logger
	once             sync.Once
)

// Initialize sets up all audit loggers
func Initialize() {
	once.Do(func() {
		CreateLogger = setupAuditLogger("create.log")
		UpdateLogger = setupAuditLogger("update.log")
		DeleteLogger = setupAuditLogger("delete.log")
		BadRequestLogger = setupAuditLogger("bad_request.log")
	})
}

// GetCreateLogger returns CreateLogger, initializing if needed
func GetCreateLogger() *logrus.Logger {
	if CreateLogger == nil {
		Initialize()
	}
	return CreateLogger
}

// GetUpdateLogger returns UpdateLogger, initializing if needed
func GetUpdateLogger() *logrus.Logger {
	if UpdateLogger == nil {
		Initialize()
	}
	return UpdateLogger
}

// GetDeleteLogger returns DeleteLogger, initializing if needed
func GetDeleteLogger() *logrus.Logger {
	if DeleteLogger == nil {
		Initialize()
	}
	return DeleteLogger
}

// GetBadRequestLogger returns BadRequestLogger, initializing if needed
func GetBadRequestLogger() *logrus.Logger {
	if BadRequestLogger == nil {
		Initialize()
	}
	return BadRequestLogger
}

// setupAuditLogger creates a dedicated logger for specific CRUD operations
func setupAuditLogger(filename string) *logrus.Logger {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	auditDir := filepath.Join(".", "audit")

	// Create audit directory if it doesn't exist
	if err := os.MkdirAll(auditDir, 0755); err != nil {
		// If we can't create the directory, log to stderr
		logger.SetOutput(os.Stderr)
		logger.WithError(err).Error("Failed to create audit directory")
		return logger
	}

	logFilePath := filepath.Join(auditDir, filename)

	// Open file directly with append mode
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		// If we can't open the file, log to stderr
		logger.SetOutput(os.Stderr)
		logger.WithError(err).Errorf("Failed to open audit log file: %s", logFilePath)
		return logger
	}

	// Write to BOTH stdout and file
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	logger.SetOutput(multiWriter)

	return logger
}
