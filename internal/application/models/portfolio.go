package models

import "gorm.io/gorm"

type Portfolio struct {
	gorm.Model
	Title       string     `json:"title"`
	Description *string    `json:"description,omitempty"`
	Sections    []Section  `json:"sections" gorm:"foreignKey:PortfolioID"`
	Categories  []Category `json:"categories" gorm:"foreignKey:PortfolioID"`
	OwnerID     string     `json:"ownerId,omitempty"`
}
