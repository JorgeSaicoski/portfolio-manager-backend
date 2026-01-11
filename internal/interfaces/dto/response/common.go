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

// DataResponse represents a response with data and message (API_OVERVIEW.md format)
// Used for single resource responses
type DataResponse struct {
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

// PaginatedDataResponse represents a paginated response (API_OVERVIEW.md format)
// Used for list/collection responses with pagination
type PaginatedDataResponse struct {
	Data    interface{} `json:"data"`
	Page    int         `json:"page"`
	Limit   int         `json:"limit"`
	Total   int64       `json:"total"`
	Message string      `json:"message"`
}
