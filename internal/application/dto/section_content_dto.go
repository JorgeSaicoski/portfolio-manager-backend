package dto

import "time"

// ============================================================================
// SectionContent DTOs (Application Layer)
// ============================================================================

// SectionContentDTO represents a section content in the application layer
type SectionContentDTO struct {
	ID        uint
	SectionID uint
	Type      string
	Content   *string
	Order     uint
	ImageID   *uint
	OwnerID   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// CreateSectionContentInput is the input for creating a section content
type CreateSectionContentInput struct {
	SectionID uint
	Type      string
	Content   *string
	Order     uint
	ImageID   *uint
	OwnerID   string
}

// UpdateSectionContentInput is the input for updating a section content
type UpdateSectionContentInput struct {
	ID      uint
	Type    string
	Content *string
	Order   uint
	ImageID *uint
	OwnerID string // For authorization check
}
