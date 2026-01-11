package entities

import "gorm.io/gorm"

// SectionContentRecord represents a section content in the database (GORM entity)
type SectionContentRecord struct {
	gorm.Model
	SectionID uint    `gorm:"not null;index"`
	Type      string  `gorm:"type:varchar(50);not null"`
	Content   *string `gorm:"type:text"`
	Order     uint    `gorm:"not null;default:0"`
	ImageID   *uint   `gorm:"index"`
	OwnerID   string  `gorm:"type:varchar(255);not null;index"`

	// Relations
	Section SectionRecord `gorm:"foreignKey:SectionID;constraint:OnDelete:CASCADE"`
}

// TableName specifies the table name for SectionContentRecord
func (SectionContentRecord) TableName() string {
	return "section_contents"
}
