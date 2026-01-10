package entities

import "gorm.io/gorm"

// PortfolioRecord represents the database structure for portfolios
// This is internal to the infrastructure layer and uses GORM tags
type PortfolioRecord struct {
	gorm.Model
	Title       string `gorm:"type:varchar(100);not null;index:idx_owner_title"`
	Description string `gorm:"type:varchar(500)"`
	OwnerID     string `gorm:"type:varchar(255);not null;index:idx_owner_title"`

	// Relations (for GORM cascade handling)
	Sections   []SectionRecord  `gorm:"foreignKey:PortfolioID;constraint:OnDelete:CASCADE"`
	Categories []CategoryRecord `gorm:"foreignKey:PortfolioID;constraint:OnDelete:CASCADE"`
}

// TableName specifies the table name for PortfolioRecord
func (PortfolioRecord) TableName() string {
	return "portfolios"
}
