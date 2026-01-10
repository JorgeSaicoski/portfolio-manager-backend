package dto

import "time"

// ============================================================================
// Portfolio DTOs (Application Layer)
// ============================================================================

// PortfolioDTO represents a portfolio in the application layer
type PortfolioDTO struct {
	ID          uint
	Title       string
	Description string
	OwnerID     string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// CreatePortfolioInput is the input for creating a portfolio
type CreatePortfolioInput struct {
	Title       string
	Description string
	OwnerID     string
}

// UpdatePortfolioInput is the input for updating a portfolio
type UpdatePortfolioInput struct {
	ID          uint
	Title       string
	Description string
	OwnerID     string // For authorization check
}

// ListPortfoliosInput is the input for listing portfolios
type ListPortfoliosInput struct {
	OwnerID    string
	Pagination PaginationDTO
}

// ListPortfoliosOutput is the output for listing portfolios
type ListPortfoliosOutput struct {
	Portfolios []PortfolioDTO
	Pagination PaginatedResultDTO
}

// ============================================================================
// Common DTOs (Pagination)
// ============================================================================

// PaginationDTO represents pagination parameters
type PaginationDTO struct {
	Page  int
	Limit int
}

// PaginatedResultDTO represents paginated result metadata
type PaginatedResultDTO struct {
	Total int64
	Page  int
	Limit int
}
