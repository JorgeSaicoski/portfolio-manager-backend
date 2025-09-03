package models

import "gorm.io/gorm"

type Section struct {
	gorm.Model
	Title       string  `json:"title"`
	Description *string `json:"description,omitempty"`
	Type        string  `json:"type"`
	OwnerID     string  `json:"ownerId,omitempty"`
	PortfolioID uint    `json:"portfolio_id"`
}
