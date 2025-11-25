package request

// CreateImageRequest represents the request to upload a new image
// Note: The actual file is handled via multipart.FileHeader in the handler
type CreateImageRequest struct {
	EntityType string `form:"entity_type" binding:"required,oneof=project portfolio section"`
	EntityID   uint   `form:"entity_id" binding:"required,min=1"`
	Type       string `form:"type" binding:"required,oneof=photo image icon logo banner avatar background"`
	Alt        string `form:"alt" binding:"max=255"`
	IsMain     bool   `form:"is_main"`
}

// UpdateImageRequest represents the request to update image metadata
type UpdateImageRequest struct {
	Alt    string `json:"alt" binding:"omitempty,max=255"`
	IsMain *bool  `json:"is_main"` // pointer to distinguish between false and not provided
}
