package entities

import "gorm.io/gorm"

// SectionRecord represents the database structure for sections
type SectionRecord struct {
	gorm.Model
	Title       string `gorm:"type:varchar(100);not null"`
	Description string `gorm:"type:varchar(500)"`
	Type        string `gorm:"type:varchar(50);not null"`
	Position    int    `gorm:"not null;default:0"`
	OwnerID     string `gorm:"type:varchar(255);not null;index"`
	PortfolioID uint   `gorm:"not null;index"`
}

// TableName specifies the table name for SectionRecord
func (SectionRecord) TableName() string {
	return "sections"
}
