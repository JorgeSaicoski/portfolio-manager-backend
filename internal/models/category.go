package models

import "gorm.io/gorm"

type Category struct {
	gorm.Model
	Title       string  `json:"title"`
	Description *string `json:"description,omitempty"`
	OwnerID     string  `json:"ownerId,omitempty"`
}
