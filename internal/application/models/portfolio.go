package models

import "gorm.io/gorm"

type Portfolio struct {
	gorm.Model
	Title         string     `json:"title"`
	Description   *string    `json:"description,omitempty"`
	CategoryCount uint       `json:"category_count" gorm:"default:0"`
	Sections      []Section  `json:"sections" gorm:"foreignKey:PortfolioID"`
	Categories    []Category `json:"categories" gorm:"foreignKey:PortfolioID"`
	OwnerID       string     `json:"ownerId,omitempty"`
}
