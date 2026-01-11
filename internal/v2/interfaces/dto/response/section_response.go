package response

import "time"

// SectionResponse represents a section in HTTP responses
type SectionResponse struct {
	ID          uint      `json:"id"`
	Title       string    `json:"title"`
	Description *string   `json:"description,omitempty"`
	Position    uint      `json:"position"`
	Type        string    `json:"type"`
	OwnerID     string    `json:"owner_id,omitempty"`
	PortfolioID uint      `json:"portfolio_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ListSectionsResponse represents the response for listing sections
type ListSectionsResponse struct {
	Sections   []SectionResponse  `json:"sections"`
	Pagination PaginationResponse `json:"pagination"`
}
