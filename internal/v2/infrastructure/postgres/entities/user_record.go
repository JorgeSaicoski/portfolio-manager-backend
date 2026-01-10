package entities

import (
	"time"

	"gorm.io/gorm"
)

// UserRecord is the GORM entity for users (infrastructure layer)
// This represents the database table structure and should NOT be exposed to the application layer
type UserRecord struct {
	ID         string `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Email      string `gorm:"type:varchar(255);not null;uniqueIndex"`
	Name       string `gorm:"type:varchar(255);not null"`
	ExternalID string `gorm:"type:varchar(255);not null;uniqueIndex"` // ID from auth provider (Authentik)
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  gorm.DeletedAt `gorm:"index"`
}

// TableName specifies the table name for the user record
func (UserRecord) TableName() string {
	return "users"
}
