package contracts

import "context"

// AuditLogger defines audit logging operations the application expects.
type AuditLogger interface {
	LogCreate(ctx context.Context, entity string, id uint, data map[string]interface{})
	LogUpdate(ctx context.Context, entity string, id uint, data map[string]interface{})
	LogDelete(ctx context.Context, entity string, id uint, data map[string]interface{})
	LogAccess(ctx context.Context, entity string, id uint, userID string, allowed bool)
}
