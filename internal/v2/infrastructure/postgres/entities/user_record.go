package entities

import "gorm.io/gorm"

// UserRecord represents the database structure for users
type UserRecord struct {
	gorm.Model
	Email      string `gorm:"type:varchar(255);uniqueIndex:idx_email;not null"`
	Name       string `gorm:"type:varchar(255)"`
	ExternalID string `gorm:"type:varchar(255);uniqueIndex:idx_external_id"` // ID from auth provider

	// Relations
	Portfolios []PortfolioRecord `gorm:"foreignKey:OwnerID;references:ID;constraint:OnDelete:CASCADE"`
}

// TableName specifies the table name
func (UserRecord) TableName() string {
	return "users"
}
