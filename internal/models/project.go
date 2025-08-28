package models

import (
	"gorm.io/gorm"
)

type Project struct {
	gorm.Model
	Title       string   `json:"title"`
	Images      []string `json:"images" gorm:"type:text[]"`
	MainImage   string   `json:"main_image"`
	Description string   `json:"description" gorm:"type:text"`
	Skills      []string `json:"skills" gorm:"type:text[]"`
	Client      string   `json:"client"`
	Link        string   `json:"link"`
	OwnerID     string   `json:"ownerId,omitempty"`
}
