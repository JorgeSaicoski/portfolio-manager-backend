package dto

import "time"

// ============================================================================
// Category DTOs (Application Layer)
// ============================================================================

// CategoryDTO represents a category in the application layer
type CategoryDTO struct {
	ID          uint
	Title       string
	Description *string
	Position    uint
	OwnerID     string
	PortfolioID uint
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// CreateCategoryInput is the input for creating a category
type CreateCategoryInput struct {
	Title       string
	Description *string
	Position    uint
	OwnerID     string
	PortfolioID uint
}

// UpdateCategoryInput is the input for updating a category
type UpdateCategoryInput struct {
	ID          uint
	Title       string
	Description *string
	Position    uint
	OwnerID     string // For authorization check
}

// ListCategoriesInput is the input for listing categories by portfolio
type ListCategoriesInput struct {
	PortfolioID uint
	Pagination  PaginationDTO
}

// ListCategoriesOutput is the output for listing categories
type ListCategoriesOutput struct {
	Categories []CategoryDTO
	Pagination PaginatedResultDTO
}

// BulkUpdatePositionItem represents a single position update
type BulkUpdatePositionItem struct {
	ID       uint
	Position uint
}

// BulkUpdateCategoryPositionsInput is the input for bulk updating category positions
type BulkUpdateCategoryPositionsInput struct {
	Items   []BulkUpdatePositionItem
	OwnerID string // For authorization check
}
