package response

import "time"

// ProjectResponse represents a project in HTTP responses
type ProjectResponse struct {
	ID          uint      `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	MainImage   *string   `json:"main_image,omitempty"`
	Images      []string  `json:"images,omitempty"`
	Skills      []string  `json:"skills,omitempty"`
	Client      *string   `json:"client,omitempty"`
	Link        *string   `json:"link,omitempty"`
	CategoryID  uint      `json:"category_id"`
	OwnerID     string    `json:"owner_id,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ListProjectsResponse represents the response for listing projects
type ListProjectsResponse struct {
	Projects   []ProjectResponse  `json:"projects"`
	Pagination PaginationResponse `json:"pagination"`
}
