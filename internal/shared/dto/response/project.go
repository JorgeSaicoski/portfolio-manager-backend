package response

import (
	"time"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/models"
)

// ProjectResponse represents a project in responses
type ProjectResponse struct {
	ID          uint       `json:"id"`
	Title       string     `json:"title"`
	Images      []string   `json:"images,omitempty"`
	MainImage   string     `json:"main_image,omitempty"`
	Description string     `json:"description"`
	Skills      []string   `json:"skills,omitempty"`
	Client      string     `json:"client,omitempty"`
	Link        string     `json:"link,omitempty"`
	OwnerID     string     `json:"owner_id,omitempty"`
	CategoryID  uint       `json:"category_id"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

// ToProjectResponse converts a model to a response DTO
func ToProjectResponse(project *models.Project) ProjectResponse {
	return ProjectResponse{
		ID:          project.ID,
		Title:       project.Title,
		Images:      project.Images,
		MainImage:   project.MainImage,
		Description: project.Description,
		Skills:      project.Skills,
		Client:      project.Client,
		Link:        project.Link,
		OwnerID:     project.OwnerID,
		CategoryID:  project.CategoryID,
		CreatedAt:   project.CreatedAt,
		UpdatedAt:   project.UpdatedAt,
		DeletedAt:   nil,
	}
}

// ToProjectListResponse converts a slice of models to response DTOs
func ToProjectListResponse(projects []*models.Project) []ProjectResponse {
	responses := make([]ProjectResponse, len(projects))
	for i, project := range projects {
		responses[i] = ToProjectResponse(project)
	}
	return responses
}
