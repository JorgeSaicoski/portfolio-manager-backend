package entities

import "gorm.io/gorm"

// CategoryRecord is the GORM entity for categories (infrastructure layer)
// This represents the database table structure and should NOT be exposed to the application layer
type CategoryRecord struct {
	gorm.Model
	Title       string  `gorm:"type:varchar(255);not null"`
	Description *string `gorm:"type:text"`
	Position    uint    `gorm:"default:0;not null"`
	OwnerID     string  `gorm:"type:varchar(255);not null;index"`
	PortfolioID uint    `gorm:"not null;index"`

	// Foreign key relationship
	Portfolio PortfolioRecord `gorm:"foreignKey:PortfolioID;constraint:OnDelete:CASCADE"`
}

// TableName specifies the table name for the category record
func (CategoryRecord) TableName() string {
	return "categories"
}
