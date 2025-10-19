package dto

// PaginationQuery represents query parameters for pagination
type PaginationQuery struct {
	Page  int `form:"page" binding:"omitempty,min=1"`
	Limit int `form:"limit" binding:"omitempty,min=1,max=100"`
}

// GetPageAndLimit returns validated page and limit values with defaults
func (p *PaginationQuery) GetPageAndLimit() (int, int) {
	page := 1
	if p.Page > 0 {
		page = p.Page
	}

	limit := 10
	if p.Limit > 0 && p.Limit <= 100 {
		limit = p.Limit
	}

	return page, limit
}

// GetOffset calculates the offset based on page and limit
func (p *PaginationQuery) GetOffset() int {
	page, limit := p.GetPageAndLimit()
	return (page - 1) * limit
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// SuccessResponse represents a generic success response
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// PaginatedResponse represents a paginated response
type PaginatedResponse struct {
	Data    interface{} `json:"data"`
	Page    int         `json:"page"`
	Limit   int         `json:"limit"`
	Message string      `json:"message"`
}
