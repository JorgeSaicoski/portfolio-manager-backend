package request

// CreatePortfolioRequest represents the request body for creating a portfolio
type CreatePortfolioRequest struct {
	Title       string  `json:"title" binding:"required,min=1,max=255"`
	Description *string `json:"description,omitempty" binding:"omitempty,max=1000"`
}

// UpdatePortfolioRequest represents the request body for updating a portfolio
type UpdatePortfolioRequest struct {
	Title       string  `json:"title" binding:"omitempty,min=1,max=255"`
	Description *string `json:"description,omitempty" binding:"omitempty,max=1000"`
}
