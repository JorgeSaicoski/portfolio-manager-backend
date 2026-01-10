package logging

import (
	"context"
	"strings"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/contracts"
	"github.com/sirupsen/logrus"
)

// auditLoggerAdapter implements the AuditLogger contract from application layer
type auditLoggerAdapter struct {
	createLogger *logrus.Logger
	updateLogger *logrus.Logger
	deleteLogger *logrus.Logger
	accessLogger *logrus.Logger
}

// NewAuditLogger creates a new audit logger adapter
// This adapts the existing audit infrastructure to the application contract
func NewAuditLogger(createLogger, updateLogger, deleteLogger, accessLogger *logrus.Logger) contracts.AuditLogger {
	return &auditLoggerAdapter{
		createLogger: createLogger,
		updateLogger: updateLogger,
		deleteLogger: deleteLogger,
		accessLogger: accessLogger,
	}
}

// LogCreate logs a resource creation event
func (a *auditLoggerAdapter) LogCreate(ctx context.Context, entity string, id uint, data map[string]interface{}) {
	if a.createLogger == nil {
		return
	}

	a.createLogger.WithFields(logrus.Fields{
		"operation": "CREATE_" + strings.ToUpper(entity),
		"id":        id,
		"data":      data,
	}).Info("Resource created")
}

// LogUpdate logs a resource update event
func (a *auditLoggerAdapter) LogUpdate(ctx context.Context, entity string, id uint, data map[string]interface{}) {
	if a.updateLogger == nil {
		return
	}

	a.updateLogger.WithFields(logrus.Fields{
		"operation": "UPDATE_" + strings.ToUpper(entity),
		"id":        id,
		"data":      data,
	}).Info("Resource updated")
}

// LogDelete logs a resource deletion event
func (a *auditLoggerAdapter) LogDelete(ctx context.Context, entity string, id uint, data map[string]interface{}) {
	if a.deleteLogger == nil {
		return
	}

	a.deleteLogger.WithFields(logrus.Fields{
		"operation": "DELETE_" + strings.ToUpper(entity),
		"id":        id,
		"data":      data,
	}).Info("Resource deleted")
}

// LogAccess logs an access attempt (for forbidden access tracking)
func (a *auditLoggerAdapter) LogAccess(ctx context.Context, entity string, id uint, userID string, allowed bool) {
	if a.accessLogger == nil {
		return
	}

	status := "ALLOWED"
	if !allowed {
		status = "FORBIDDEN"
	}

	a.accessLogger.WithFields(logrus.Fields{
		"operation": "ACCESS_" + strings.ToUpper(entity),
		"id":        id,
		"userID":    userID,
		"status":    status,
	}).Info("Access attempt")
}
