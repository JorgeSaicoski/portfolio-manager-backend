package response

import "time"

// CategoryResponse represents a category in HTTP responses
type CategoryResponse struct {
	ID          uint      `json:"id"`
	Title       string    `json:"title"`
	Description *string   `json:"description,omitempty"`
	Position    uint      `json:"position"`
	OwnerID     string    `json:"owner_id,omitempty"`
	PortfolioID uint      `json:"portfolio_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ListCategoriesResponse represents the response for listing categories
type ListCategoriesResponse struct {
	Categories []CategoryResponse `json:"categories"`
	Pagination PaginationResponse `json:"pagination"`
}
