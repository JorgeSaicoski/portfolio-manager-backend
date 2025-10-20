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
	return pq.Array(a).Scan(value)
}

// Value implements the driver.Valuer interface
func (a StringArray) Value() (driver.Value, error) {
	if a == nil {
		return pq.Array([]string{}).Value()
	}
	return pq.Array(a).Value()
}

type Project struct {
	gorm.Model
	Title       string      `json:"title"`
	Images      StringArray `json:"images" gorm:"type:text[]"`
	MainImage   string      `json:"main_image"`
	Description string      `json:"description" gorm:"type:text"`
	Skills      StringArray `json:"skills" gorm:"type:text[]"`
	Client      string      `json:"client"`
	Link        string      `json:"link"`
	Position    uint        `json:"position"` // Todo Implement this
	OwnerID     string      `json:"ownerId,omitempty"`
	CategoryID  uint        `json:"category_id"`
}
