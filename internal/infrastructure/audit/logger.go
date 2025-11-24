package audit

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	CreateLogger *logrus.Logger
	UpdateLogger *logrus.Logger
	DeleteLogger *logrus.Logger
	once         sync.Once
)

// Initialize sets up all audit loggers
func Initialize() {
	once.Do(func() {
		fmt.Println("[DEBUG] Audit Initialize() called")
		CreateLogger = setupAuditLogger("create.log")
		UpdateLogger = setupAuditLogger("update.log")
		DeleteLogger = setupAuditLogger("delete.log")
		fmt.Println("[DEBUG] All audit loggers initialized")

		// Test log immediately
		CreateLogger.Info("[TEST] Create audit logger is working")
		UpdateLogger.Info("[TEST] Update audit logger is working")
		DeleteLogger.Info("[TEST] Delete audit logger is working")
		fmt.Println("[DEBUG] Test logs written")
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

// setupAuditLogger creates a dedicated logger for specific CRUD operations
func setupAuditLogger(filename string) *logrus.Logger {
	fmt.Printf("[DEBUG] Setting up audit logger: %s\n", filename)

	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	auditDir := filepath.Join(".", "audit")
	fmt.Printf("[DEBUG] Audit directory path: %s\n", auditDir)

	// Create audit directory if it doesn't exist
	if err := os.MkdirAll(auditDir, 0755); err != nil {
		fmt.Printf("[ERROR] Failed to create audit directory: %v\n", err)
		// If we can't create the directory, log to stderr
		logger.SetOutput(os.Stderr)
		logger.WithError(err).Error("Failed to create audit directory")
		return logger
	}
	fmt.Printf("[DEBUG] Audit directory created/verified\n")

	logFilePath := filepath.Join(auditDir, filename)
	fmt.Printf("[DEBUG] Log file path: %s\n", logFilePath)

	logFile := &lumberjack.Logger{
		Filename:   logFilePath,
		MaxSize:    10, // megabytes
		MaxBackups: 30, // keep 30 old log files
		MaxAge:     90, // days
		Compress:   true,
	}

	// Write to BOTH stdout and file for debugging
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	logger.SetOutput(multiWriter)

	fmt.Printf("[DEBUG] Logger configured for %s with MultiWriter (stdout + file)\n", filename)
	return logger
}
