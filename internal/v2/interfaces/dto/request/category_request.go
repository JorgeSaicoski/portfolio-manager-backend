package request

// CreateCategoryRequest represents HTTP request for creating a category
type CreateCategoryRequest struct {
	Title       string  `json:"title" binding:"required,min=1,max=255"`
	Description *string `json:"description,omitempty" binding:"omitempty,max=1000"`
	Position    uint    `json:"position" binding:"omitempty"`
	PortfolioID uint    `json:"portfolio_id" binding:"required,min=1"`
}

// UpdateCategoryRequest represents HTTP request for updating a category
type UpdateCategoryRequest struct {
	Title       string  `json:"title" binding:"omitempty,min=1,max=255"`
	Description *string `json:"description,omitempty" binding:"omitempty,max=1000"`
	Position    uint    `json:"position" binding:"omitempty"`
}

// UpdateCategoryPositionRequest represents HTTP request for updating a category's position
type UpdateCategoryPositionRequest struct {
	Position uint `json:"position" binding:"required"`
}

// BulkUpdatePositionItemRequest represents a single position update in bulk operation
type BulkUpdatePositionItemRequest struct {
	ID       uint `json:"id" binding:"required"`
	Position uint `json:"position" binding:"required"`
}

// BulkReorderCategoriesRequest represents HTTP request for bulk reordering categories
type BulkReorderCategoriesRequest struct {
	Items []BulkUpdatePositionItemRequest `json:"items" binding:"required,min=1,dive"`
}

// ListCategoriesRequest represents HTTP request for listing categories
type ListCategoriesRequest struct {
	Page  int `form:"page" binding:"omitempty,min=1"`
	Limit int `form:"limit" binding:"omitempty,min=1,max=100"`
}
