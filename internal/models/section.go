package models

import "gorm.io/gorm"

type Section struct {
	gorm.Model
	Title       string           `json:"title"`
	Description *string          `json:"description,omitempty"`
	Type        string           `json:"type"`
	Position    uint             `json:"position" gorm:"default:0"`
	OwnerID     string           `json:"ownerId,omitempty"`
	PortfolioID uint             `json:"portfolio_id"`
	Contents    []SectionContent `json:"contents,omitempty" gorm:"foreignKey:SectionID"`
}
