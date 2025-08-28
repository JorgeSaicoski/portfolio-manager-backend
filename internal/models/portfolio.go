package models

import "gorm.io/gorm"

type Portfolio struct {
	gorm.Model
	Title       string     `json:"title"`
	Description *string    `json:"description,omitempty"`
	Sections    []Section  `json:"sections" gorm:"foreignKey:SectionID"`
	Categories  []Category `json:"categories" gorm:"foreignKey:CategoryID"`
	OwnerID     string     `json:"ownerId,omitempty"`
}
