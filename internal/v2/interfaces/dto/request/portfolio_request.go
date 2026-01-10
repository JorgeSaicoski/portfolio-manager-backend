package request

// CreatePortfolioRequest represents the HTTP request body for creating a portfolio
type CreatePortfolioRequest struct {
	Title       string `json:"title" binding:"required,min=1,max=255"`
	Description string `json:"description,omitempty" binding:"omitempty,max=1000"`
}

// UpdatePortfolioRequest represents the HTTP request body for updating a portfolio
type UpdatePortfolioRequest struct {
	Title       string `json:"title,omitempty" binding:"omitempty,min=1,max=255"`
	Description string `json:"description,omitempty" binding:"omitempty,max=1000"`
}

// ListPortfoliosRequest represents the HTTP query parameters for listing portfolios
type ListPortfoliosRequest struct {
	Page  int `form:"page" binding:"omitempty,min=1"`
	Limit int `form:"limit" binding:"omitempty,min=1,max=100"`
}
