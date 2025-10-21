package request

// CreateCategoryRequest represents the request body for creating a category
type CreateCategoryRequest struct {
	Title       string  `json:"title" binding:"required,min=1,max=255"`
	Description *string `json:"description,omitempty" binding:"omitempty,max=1000"`
	PortfolioID uint    `json:"portfolio_id" binding:"required,min=1"`
}

// UpdateCategoryRequest represents the request body for updating a category
type UpdateCategoryRequest struct {
	Title       string  `json:"title" binding:"omitempty,min=1,max=255"`
	Description *string `json:"description,omitempty" binding:"omitempty,max=1000"`
	PortfolioID uint    `json:"portfolio_id" binding:"omitempty,min=1"`
}
