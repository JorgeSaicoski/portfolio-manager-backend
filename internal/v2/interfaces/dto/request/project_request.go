package request

// CreateProjectRequest represents HTTP request for creating a project
type CreateProjectRequest struct {
	Title       string   `json:"title" binding:"required,min=1,max=255"`
	Description string   `json:"description" binding:"required,min=1"`
	MainImage   *string  `json:"main_image,omitempty" binding:"omitempty,url"`
	Images      []string `json:"images,omitempty" binding:"omitempty,dive,url"`
	Skills      []string `json:"skills,omitempty"`
	Client      *string  `json:"client,omitempty" binding:"omitempty,max=255"`
	Link        *string  `json:"link,omitempty" binding:"omitempty,url"`
	CategoryID  uint     `json:"category_id" binding:"required,min=1"`
}

// UpdateProjectRequest represents HTTP request for updating a project
type UpdateProjectRequest struct {
	Title       string   `json:"title" binding:"omitempty,min=1,max=255"`
	Description string   `json:"description" binding:"omitempty,min=1"`
	MainImage   *string  `json:"main_image,omitempty" binding:"omitempty,url"`
	Images      []string `json:"images,omitempty" binding:"omitempty,dive,url"`
	Skills      []string `json:"skills,omitempty"`
	Client      *string  `json:"client,omitempty" binding:"omitempty,max=255"`
	Link        *string  `json:"link,omitempty" binding:"omitempty,url"`
}

// ListProjectsRequest represents HTTP request for listing projects
type ListProjectsRequest struct {
	Page  int `form:"page" binding:"omitempty,min=1"`
	Limit int `form:"limit" binding:"omitempty,min=1,max=100"`
}

// SearchProjectsBySkillsRequest represents HTTP request for searching projects by skills
type SearchProjectsBySkillsRequest struct {
	Skills []string `form:"skills" binding:"required,min=1"`
}

// SearchProjectsByClientRequest represents HTTP request for searching projects by client
type SearchProjectsByClientRequest struct {
	Client string `form:"client" binding:"required,min=1"`
}
