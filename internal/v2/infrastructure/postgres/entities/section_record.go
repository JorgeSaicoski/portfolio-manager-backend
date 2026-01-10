package entities

import "gorm.io/gorm"

// SectionRecord is the GORM entity for sections (infrastructure layer)
// This represents the database table structure and should NOT be exposed to the application layer
type SectionRecord struct {
	gorm.Model
	Title       string  `gorm:"type:varchar(255);not null"`
	Description *string `gorm:"type:text"`
	Type        string  `gorm:"type:varchar(100)"` // Optional: NavBar, HomePageSection, etc.
	Position    uint    `gorm:"default:0;not null"`
	OwnerID     string  `gorm:"type:varchar(255);not null;index"`
	PortfolioID uint    `gorm:"not null;index"`

	// Foreign key relationship
	Portfolio PortfolioRecord `gorm:"foreignKey:PortfolioID;constraint:OnDelete:CASCADE"`
}

// TableName specifies the table name for the section record
func (SectionRecord) TableName() string {
	return "sections"
}
