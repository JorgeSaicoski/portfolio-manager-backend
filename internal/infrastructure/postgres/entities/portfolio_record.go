package entities

import "gorm.io/gorm"

// PortfolioRecord is the GORM entity for portfolios (infrastructure layer)
// This represents the database table structure and should NOT be exposed to the application layer
type PortfolioRecord struct {
	gorm.Model
	Title       string `gorm:"type:varchar(255);not null"`
	Description string `gorm:"type:text"`
	OwnerID     string `gorm:"type:varchar(255);not null;index"`

	// Add index for checking duplicate titles per owner
	// Composite index on (owner_id, title) for efficient duplicate checking
}

// TableName specifies the table name for the portfolio record
func (PortfolioRecord) TableName() string {
	return "portfolios"
}
