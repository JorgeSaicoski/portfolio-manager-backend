package models

import (
	"time"

	"gorm.io/gorm"
)

type Project struct {
	gorm.Model
	Title       string    `json:"title"`
	Images      []string  `json:"images" gorm:"type:text[]"` // ["/uploads/projects/1/img1.jpg", "/uploads/projects/1/img2.jpg"]
	MainImage   string    `json:"main_image"`                // "/uploads/projects/1/main.jpg"
	Description string    `json:"description" gorm:"type:text"`
	Skills      []string  `json:"skills" gorm:"type:text[]"`
	Client      string    `json:"client"`
	Link        string    `json:"link"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
