package db

import (
	"log"

	"gorm.io/gorm"
)

// RemoveImageFeature removes all image-related tables and columns
func RemoveImageFeature(db *gorm.DB) error {
	log.Println("Starting image feature removal migration...")

	// Drop the images table entirely
	if err := db.Exec("DROP TABLE IF EXISTS images CASCADE").Error; err != nil {
		log.Printf("Error dropping images table: %v", err)
		return err
	}
	log.Println("Dropped images table")

	// Remove image-related columns from projects table (if they exist)
	db.Exec("ALTER TABLE projects DROP COLUMN IF EXISTS main_image")
	db.Exec("ALTER TABLE projects DROP COLUMN IF EXISTS images")

	log.Println("Image feature removal migration completed successfully")
	return nil
}
