package dto

import "time"

// ============================================================================
// Project DTOs (Application Layer)
// ============================================================================

// ProjectDTO represents a project in the application layer
type ProjectDTO struct {
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

// CreateProjectInput is the input for creating a project
type CreateProjectInput struct {
	Title       string
	Description string
	MainImage   *string
	Images      []string
	Skills      []string
	Client      *string
	Link        *string
	CategoryID  uint
	OwnerID     string
}

// UpdateProjectInput is the input for updating a project
type UpdateProjectInput struct {
	ID          uint
	Title       string
	Description string
	MainImage   *string
	Images      []string
	Skills      []string
	Client      *string
	Link        *string
	OwnerID     string // For authorization check
}

// ListProjectsInput is the input for listing projects
type ListProjectsInput struct {
	OwnerID    string
	Pagination PaginationDTO
}

// ListProjectsOutput is the output for listing projects
type ListProjectsOutput struct {
	Projects   []ProjectDTO
	Pagination PaginatedResultDTO
}
