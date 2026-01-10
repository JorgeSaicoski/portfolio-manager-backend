package entities

import "gorm.io/gorm"

// CategoryRecord represents the database structure for categories
type CategoryRecord struct {
	gorm.Model
	Title       string `gorm:"type:varchar(100);not null"`
	Description string `gorm:"type:varchar(500)"`
	Position    int    `gorm:"not null;default:0"`
	OwnerID     string `gorm:"type:varchar(255);not null;index"`
	PortfolioID uint   `gorm:"not null;index"`
}

// TableName specifies the table name for CategoryRecord
func (CategoryRecord) TableName() string {
	return "categories"
}
