package models

import "gorm.io/gorm"

type Portfolio struct {
	gorm.Model
	Title       string     `json:"title"`
	Description *string    `json:"description,omitempty"`
	Sections    []Section  `json:"sections" gorm:"foreignKey:PortfolioID;constraint:OnDelete:CASCADE"`
	Categories  []Category `json:"categories" gorm:"foreignKey:PortfolioID;constraint:OnDelete:CASCADE"`
	OwnerID     string     `json:"ownerId,omitempty"`
}
