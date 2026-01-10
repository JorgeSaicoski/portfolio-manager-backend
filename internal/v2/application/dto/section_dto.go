package dto

import "time"

// ============================================================================
// Section DTOs (Application Layer)
// ============================================================================

// SectionDTO represents a section in the application layer
type SectionDTO struct {
	ID          uint
	Title       string
	Description *string
	Type        string // Optional: could be NavBar, HomePageSection, etc.
	Position    uint
	OwnerID     string
	PortfolioID uint
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// CreateSectionInput is the input for creating a section
type CreateSectionInput struct {
	Title       string
	Description *string
	Type        string
	Position    uint
	OwnerID     string
	PortfolioID uint
}

// UpdateSectionInput is the input for updating a section
type UpdateSectionInput struct {
	ID          uint
	Title       string
	Description *string
	Type        string
	Position    uint
	OwnerID     string // For authorization check
}

// ListSectionsInput is the input for listing sections by portfolio
type ListSectionsInput struct {
	PortfolioID uint
	Pagination  PaginationDTO
}

// ListSectionsOutput is the output for listing sections
type ListSectionsOutput struct {
	Sections   []SectionDTO
	Pagination PaginatedResultDTO
}

// BulkUpdateSectionPositionsInput is the input for bulk updating section positions
type BulkUpdateSectionPositionsInput struct {
	Items   []BulkUpdatePositionItem
	OwnerID string // For authorization check
}
