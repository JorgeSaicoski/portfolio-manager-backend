package request

// CreateSectionRequest represents HTTP request for creating a section
type CreateSectionRequest struct {
	Title       string  `json:"title" binding:"required,min=1,max=255"`
	Description *string `json:"description,omitempty" binding:"omitempty,max=1000"`
	Position    uint    `json:"position" binding:"omitempty"`
	Type        string  `json:"type" binding:"required,min=1,max=50"`
	PortfolioID uint    `json:"portfolio_id" binding:"required,min=1"`
}

// UpdateSectionRequest represents HTTP request for updating a section
type UpdateSectionRequest struct {
	Title       string  `json:"title" binding:"omitempty,min=1,max=255"`
	Description *string `json:"description,omitempty" binding:"omitempty,max=1000"`
	Position    uint    `json:"position" binding:"omitempty"`
	Type        string  `json:"type" binding:"omitempty,min=1,max=50"`
}

// UpdateSectionPositionRequest represents HTTP request for updating a section's position
type UpdateSectionPositionRequest struct {
	Position uint `json:"position" binding:"required"`
}

// BulkReorderSectionsRequest represents HTTP request for bulk reordering sections
type BulkReorderSectionsRequest struct {
	Items []BulkUpdatePositionItemRequest `json:"items" binding:"required,min=1,dive"`
}

// ListSectionsRequest represents HTTP request for listing sections
type ListSectionsRequest struct {
	Page  int `form:"page" binding:"omitempty,min=1"`
	Limit int `form:"limit" binding:"omitempty,min=1,max=100"`
}
