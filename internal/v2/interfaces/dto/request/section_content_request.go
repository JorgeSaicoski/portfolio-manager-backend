package request

// CreateSectionContentRequest represents HTTP request for creating a section content
type CreateSectionContentRequest struct {
	SectionID uint    `json:"section_id" binding:"required,min=1"`
	Type      string  `json:"type" binding:"required,min=1,max=50"`
	Content   *string `json:"content,omitempty"`
	Order     uint    `json:"order" binding:"omitempty,min=0"`
	ImageID   *uint   `json:"image_id,omitempty" binding:"omitempty,min=1"`
}

// UpdateSectionContentRequest represents HTTP request for updating a section content
type UpdateSectionContentRequest struct {
	Type    string  `json:"type" binding:"omitempty,min=1,max=50"`
	Content *string `json:"content,omitempty"`
	Order   uint    `json:"order" binding:"omitempty,min=0"`
	ImageID *uint   `json:"image_id,omitempty" binding:"omitempty,min=1"`
}

// UpdateSectionContentOrderRequest represents HTTP request for updating section content order
type UpdateSectionContentOrderRequest struct {
	Order uint `json:"order" binding:"required,min=0"`
}
