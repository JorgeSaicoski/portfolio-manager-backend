package errorlog

import (
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	ClientErrorLogger *logrus.Logger // 4xx errors
	ServerErrorLogger *logrus.Logger // 5xx errors
	once              sync.Once
)

// Initialize sets up the error loggers for 4xx and 5xx errors
func Initialize() {
	once.Do(func() {
		ClientErrorLogger = setupErrorLogger("400s.log")
		ServerErrorLogger = setupErrorLogger("500s.log")
	})
}

// setupErrorLogger creates a logger with rotation and JSON formatting
func setupErrorLogger(filename string) *logrus.Logger {
	logger := logrus.New()

	// JSON formatting for structured logs
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02T15:04:05Z07:00",
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "level",
			logrus.FieldKeyMsg:   "message",
		},
	})

	logger.SetLevel(logrus.ErrorLevel)

	// Create errors directory
	errorsDir := filepath.Join(".", "errors")
	if err := os.MkdirAll(errorsDir, 0755); err != nil {
		logger.SetOutput(os.Stderr)
		logger.WithError(err).Error("Failed to create errors directory")
		return logger
	}

	// Configure log rotation
	logFile := &lumberjack.Logger{
		Filename:   filepath.Join(errorsDir, filename),
		MaxSize:    10,   // megabytes
		MaxBackups: 30,   // files
		MaxAge:     90,   // days
		Compress:   true, // gzip
	}

	// Write to both stdout and file
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	logger.SetOutput(multiWriter)

	return logger
}

// GetClientErrorLogger returns the 4xx error logger
func GetClientErrorLogger() *logrus.Logger {
	if ClientErrorLogger == nil {
		Initialize()
	}
	return ClientErrorLogger
}

// GetServerErrorLogger returns the 5xx error logger
func GetServerErrorLogger() *logrus.Logger {
	if ServerErrorLogger == nil {
		Initialize()
	}
	return ServerErrorLogger
}
