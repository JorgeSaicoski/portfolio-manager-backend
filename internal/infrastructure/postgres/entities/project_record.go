package entities

import (
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// ProjectRecord represents a project in the database (GORM entity)
type ProjectRecord struct {
	gorm.Model
	Title       string         `gorm:"type:varchar(255);not null"`
	Description string         `gorm:"type:text;not null"`
	MainImage   *string        `gorm:"type:varchar(500)"`
	Images      pq.StringArray `gorm:"type:text[]"`
	Skills      pq.StringArray `gorm:"type:text[]"`
	Client      *string        `gorm:"type:varchar(255)"`
	Link        *string        `gorm:"type:varchar(500)"`
	CategoryID  uint           `gorm:"not null;index"`
	OwnerID     string         `gorm:"type:varchar(255);not null;index"`

	// Relations
	Category CategoryRecord `gorm:"foreignKey:CategoryID;constraint:OnDelete:CASCADE"`
}

// TableName specifies the table name for ProjectRecord
func (ProjectRecord) TableName() string {
	return "projects"
}
