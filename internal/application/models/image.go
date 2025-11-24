package models

import "gorm.io/gorm"

type Image struct {
	gorm.Model
	URL          string `json:"url"`                                 // e.g., "/uploads/images/original/abc123.jpg"
	ThumbnailURL string `json:"thumbnail_url"`                       // e.g., "/uploads/images/thumbnail/abc123.jpg"
	FileName     string `json:"file_name"`                           // Original filename
	FileSize     int64  `json:"file_size"`                           // File size in bytes
	MimeType     string `json:"mime_type"`                           // e.g., "image/jpeg"
	Alt          string `json:"alt" gorm:"type:varchar(255)"`        // Alt text for accessibility
	OwnerID      string `json:"owner_id" gorm:"index"`               // User who owns this image
	Type         string `json:"type" gorm:"type:varchar(50)"`        // "photo" | "image" | "icon" | "logo" | "banner" | "avatar" | "background"
	EntityID     uint   `json:"entity_id" gorm:"index"`              // ID of the related entity
	EntityType   string `json:"entity_type" gorm:"type:varchar(50)"` // "project" | "portfolio" | "section"
	IsMain       bool   `json:"is_main" gorm:"default:false"`        // Is this the main image for the entity?
}

// TableName specifies the table name for the Image model
func (Image) TableName() string {
	return "images"
}
