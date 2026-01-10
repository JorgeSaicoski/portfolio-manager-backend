package response

import "time"

// PortfolioResponse represents a portfolio in API responses
type PortfolioResponse struct {
	ID          uint      `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	OwnerID     string    `json:"owner_id,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ListPortfoliosResponse represents the response for listing portfolios
type ListPortfoliosResponse struct {
	Portfolios []PortfolioResponse `json:"portfolios"`
	Pagination PaginationResponse  `json:"pagination"`
}
