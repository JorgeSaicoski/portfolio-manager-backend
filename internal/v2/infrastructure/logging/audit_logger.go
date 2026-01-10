package logging

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/contracts"
)

// auditLogger is the implementation of the AuditLogger contract
type auditLogger struct {
	createLogger *logrus.Logger
	updateLogger *logrus.Logger
	deleteLogger *logrus.Logger
	accessLogger *logrus.Logger
}

// NewAuditLogger creates a new audit logger instance
// Returns the interface type (contracts.AuditLogger), not the concrete type
func NewAuditLogger() contracts.AuditLogger {
	// Ensure logs directory exists
	logsDir := "logs"
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		// Fallback to current directory if can't create logs directory
		logsDir = "."
	}

	return &auditLogger{
		createLogger: setupLogger(filepath.Join(logsDir, "create.log")),
		updateLogger: setupLogger(filepath.Join(logsDir, "update.log")),
		deleteLogger: setupLogger(filepath.Join(logsDir, "delete.log")),
		accessLogger: setupLogger(filepath.Join(logsDir, "access.log")),
	}
}

// setupLogger creates and configures a logrus logger for a specific log file
func setupLogger(filename string) *logrus.Logger {
	logger := logrus.New()

	// Open log file
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		// If file can't be opened, log to stdout only
		logger.SetOutput(os.Stdout)
	} else {
		// Write to both file and stdout
		logger.SetOutput(io.MultiWriter(os.Stdout, file))
	}

	// Use JSON formatting for structured logging
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})

	// Set log level
	logger.SetLevel(logrus.InfoLevel)

	return logger
}

// LogCreate logs entity creation events
func (l *auditLogger) LogCreate(ctx context.Context, entity string, id uint, data map[string]interface{}) {
	l.createLogger.WithFields(logrus.Fields{
		"entity": entity,
		"id":     id,
		"data":   data,
	}).Info("Entity created")
}

// LogUpdate logs entity update events
func (l *auditLogger) LogUpdate(ctx context.Context, entity string, id uint, data map[string]interface{}) {
	l.updateLogger.WithFields(logrus.Fields{
		"entity": entity,
		"id":     id,
		"data":   data,
	}).Info("Entity updated")
}

// LogDelete logs entity deletion events
func (l *auditLogger) LogDelete(ctx context.Context, entity string, id uint, data map[string]interface{}) {
	l.deleteLogger.WithFields(logrus.Fields{
		"entity": entity,
		"id":     id,
		"data":   data,
	}).Info("Entity deleted")
}

// LogAccess logs entity access attempts (authorized or unauthorized)
func (l *auditLogger) LogAccess(ctx context.Context, entity string, id uint, userID string, allowed bool) {
	level := logrus.InfoLevel
	message := "Access granted"

	if !allowed {
		level = logrus.WarnLevel
		message = "Access denied"
	}

	l.accessLogger.WithFields(logrus.Fields{
		"entity":  entity,
		"id":      id,
		"userID":  userID,
		"allowed": allowed,
	}).Log(level, message)
}
