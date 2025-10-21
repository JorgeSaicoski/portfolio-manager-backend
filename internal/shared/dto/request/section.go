package request

// CreateSectionRequest represents the request body for creating a section
type CreateSectionRequest struct {
	Title       string  `json:"title" binding:"required,min=1,max=255"`
	Description *string `json:"description,omitempty" binding:"omitempty,max=1000"`
	Type        string  `json:"type" binding:"required,min=1,max=100"`
	PortfolioID uint    `json:"portfolio_id" binding:"required,min=1"`
}

// UpdateSectionRequest represents the request body for updating a section
type UpdateSectionRequest struct {
	Title       string  `json:"title" binding:"omitempty,min=1,max=255"`
	Description *string `json:"description,omitempty" binding:"omitempty,max=1000"`
	Type        string  `json:"type" binding:"omitempty,min=1,max=100"`
	PortfolioID uint    `json:"portfolio_id" binding:"omitempty,min=1"`
}
