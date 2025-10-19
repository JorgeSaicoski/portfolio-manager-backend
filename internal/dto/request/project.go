package request

// CreateProjectRequest represents the request body for creating a project
type CreateProjectRequest struct {
	Title       string   `json:"title" binding:"required,min=1,max=255"`
	Images      []string `json:"images,omitempty"`
	MainImage   string   `json:"main_image" binding:"omitempty,url"`
	Description string   `json:"description" binding:"required,min=1"`
	Skills      []string `json:"skills,omitempty"`
	Client      string   `json:"client" binding:"omitempty,max=255"`
	Link        string   `json:"link" binding:"omitempty,url"`
	CategoryID  uint     `json:"category_id" binding:"required,min=1"`
}

// UpdateProjectRequest represents the request body for updating a project
type UpdateProjectRequest struct {
	Title       string   `json:"title" binding:"omitempty,min=1,max=255"`
	Images      []string `json:"images,omitempty"`
	MainImage   string   `json:"main_image" binding:"omitempty,url"`
	Description string   `json:"description" binding:"omitempty,min=1"`
	Skills      []string `json:"skills,omitempty"`
	Client      string   `json:"client" binding:"omitempty,max=255"`
	Link        string   `json:"link" binding:"omitempty,url"`
	CategoryID  uint     `json:"category_id" binding:"omitempty,min=1"`
}
