package request

// CreateSectionContentRequest represents the request body for creating a section content block
type CreateSectionContentRequest struct {
	Type      string  `json:"type" binding:"required,oneof=text image"`
	Content   string  `json:"content" binding:"required,min=1,max=5000"`
	Order     *uint   `json:"order,omitempty" binding:"omitempty"`
	Metadata  *string `json:"metadata,omitempty" binding:"omitempty"`
	SectionID uint    `json:"section_id" binding:"required,min=1"`
}

// UpdateSectionContentRequest represents the request body for updating a section content block
type UpdateSectionContentRequest struct {
	Type     string  `json:"type" binding:"omitempty,oneof=text image"`
	Content  string  `json:"content" binding:"omitempty,min=1,max=5000"`
	Order    *uint   `json:"order,omitempty" binding:"omitempty"`
	Metadata *string `json:"metadata,omitempty" binding:"omitempty"`
}

// UpdateSectionContentOrderRequest represents the request body for updating just the order
type UpdateSectionContentOrderRequest struct {
	Order uint `json:"order"`
}
