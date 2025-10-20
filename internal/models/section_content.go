package models

import "gorm.io/gorm"

// SectionContent represents a content block within a section
// Can be either text or image type, allowing flexible section composition
type SectionContent struct {
	gorm.Model
	SectionID uint    `json:"section_id" gorm:"not null;index"`
	Type      string  `json:"type" gorm:"type:varchar(10);not null"` // "text" or "image"
	Content   string  `json:"content" gorm:"type:text;not null"`
	Order     uint    `json:"order" gorm:"default:0;index"`
	Metadata  *string `json:"metadata,omitempty" gorm:"type:jsonb"` // JSON for additional properties
	OwnerID   string  `json:"owner_id,omitempty" gorm:"type:varchar(255);index"`

	// Relationship back to Section
	Section Section `json:"-" gorm:"foreignKey:SectionID;constraint:OnDelete:CASCADE"`
}
