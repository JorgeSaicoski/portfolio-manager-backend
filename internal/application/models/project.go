package models

import (
	"database/sql/driver"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

// StringArray is a custom type for PostgreSQL text[] arrays
type StringArray []string

// Scan implements the sql.Scanner interface
func (a *StringArray) Scan(value interface{}) error {
	if value == nil {
		*a = []string{}
		return nil
	}
	// Convert *StringArray to *[]string explicitly for pq.Array
	arr := (*[]string)(a)
	return pq.Array(arr).Scan(value)
}

// Value implements the driver.Valuer interface
func (a StringArray) Value() (driver.Value, error) {
	if a == nil {
		return pq.Array([]string{}).Value()
	}
	// Convert StringArray to []string explicitly for clarity
	arr := ([]string)(a)
	return pq.Array(arr).Value()
}

type Project struct {
	gorm.Model
	Title       string      `json:"title"`
	Images      []Image     `json:"images,omitempty" gorm:"polymorphic:Entity;polymorphicValue:project"`
	Description string      `json:"description" gorm:"type:text"`
	Skills      StringArray `json:"skills" gorm:"type:text[]"`
	Client      string      `json:"client"`
	Link        string      `json:"link"`
	Position    uint        `json:"position" gorm:"default:0"`
	OwnerID     string      `json:"ownerId,omitempty"`
	CategoryID  uint        `json:"category_id"`
}
