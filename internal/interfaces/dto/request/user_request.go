package request

// UpdateUserRequest represents the HTTP request body for updating user profile
type UpdateUserRequest struct {
	Name string `json:"name" binding:"required,min=1,max=255"`
}
