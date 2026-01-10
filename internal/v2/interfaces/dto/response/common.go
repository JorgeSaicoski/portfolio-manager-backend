package response

// ErrorResponse represents a standard error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// SuccessResponse represents a standard success response
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// PaginationResponse represents pagination metadata in API responses
type PaginationResponse struct {
	Total int64 `json:"total"`
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
}
