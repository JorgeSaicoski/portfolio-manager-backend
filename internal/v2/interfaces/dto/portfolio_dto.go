package dto

import "time"

// Portfolio Request/Response DTOs for v2 interfaces layer
// CreatePortfolioRequest represents the HTTP request to create a portfolio
type CreatePortfolioRequest struct {
	Title       string `json:"title" binding:"required,min=1,max=200"`
	Description string `json:"description" binding:"max=1000"`
}

// UpdatePortfolioRequest represents the HTTP request to update a portfolio
type UpdatePortfolioRequest struct {
	Title       string `json:"title" binding:"omitempty,min=1,max=200"`
	Description string `json:"description" binding:"omitempty,max=1000"`
}

// PortfolioResponse represents a portfolio in HTTP responses
type PortfolioResponse struct {
	ID          uint      `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	OwnerID     string    `json:"owner_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ListPortfoliosResponse represents a list of portfolios with pagination
type ListPortfoliosResponse struct {
	Data  []PortfolioResponse `json:"data"`
	Total int64               `json:"total"`
	Page  int                 `json:"page"`
	Limit int                 `json:"limit"`
}
