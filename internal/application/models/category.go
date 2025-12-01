package models

import "gorm.io/gorm"

type Category struct {
	gorm.Model
	Title       string    `json:"title"`
	Description *string   `json:"description,omitempty"`
	Position    uint      `json:"position" gorm:"default:0"`
	OwnerID     string    `json:"ownerId,omitempty"`
	PortfolioID uint      `json:"portfolio_id"`
	Projects    []Project `json:"projects" gorm:"foreignKey:CategoryID;constraint:OnDelete:CASCADE"`
}
