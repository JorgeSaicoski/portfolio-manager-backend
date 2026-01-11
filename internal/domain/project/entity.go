package project

import "time"

// Project represents a project entity in the domain layer
type Project struct {
	ID          uint
	Title       string
	Description string
	MainImage   *string
	Images      []string
	Skills      []string
	Client      *string
	Link        *string
	CategoryID  uint
	OwnerID     string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
