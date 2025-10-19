package response

import (
	"time"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/models"
)

// CategoryResponse represents a basic category in responses
type CategoryResponse struct {
	ID          uint       `json:"id"`
	Title       string     `json:"title"`
	Description *string    `json:"description,omitempty"`
	OwnerID     string     `json:"owner_id,omitempty"`
	PortfolioID uint       `json:"portfolio_id"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

// CategoryDetailResponse represents a detailed category with projects
type CategoryDetailResponse struct {
	ID          uint              `json:"id"`
	Title       string            `json:"title"`
	Description *string           `json:"description,omitempty"`
	OwnerID     string            `json:"owner_id,omitempty"`
	PortfolioID uint              `json:"portfolio_id"`
	Projects    []ProjectResponse `json:"projects,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	DeletedAt   *time.Time        `json:"deleted_at,omitempty"`
}

// ToCategoryResponse converts a model to a basic response DTO
func ToCategoryResponse(category *models.Category) CategoryResponse {
	return CategoryResponse{
		ID:          category.ID,
		Title:       category.Title,
		Description: category.Description,
		OwnerID:     category.OwnerID,
		PortfolioID: category.PortfolioID,
		CreatedAt:   category.CreatedAt,
		UpdatedAt:   category.UpdatedAt,
		DeletedAt:   nil,
	}
}

// ToCategoryDetailResponse converts a model with relations to a detailed response DTO
func ToCategoryDetailResponse(category *models.Category) CategoryDetailResponse {
	projects := make([]ProjectResponse, len(category.Projects))
	for i, project := range category.Projects {
		projects[i] = ToProjectResponse(&project)
	}

	return CategoryDetailResponse{
		ID:          category.ID,
		Title:       category.Title,
		Description: category.Description,
		OwnerID:     category.OwnerID,
		PortfolioID: category.PortfolioID,
		Projects:    projects,
		CreatedAt:   category.CreatedAt,
		UpdatedAt:   category.UpdatedAt,
		DeletedAt:   nil,
	}
}

// ToCategoryListResponse converts a slice of models to response DTOs
func ToCategoryListResponse(categories []*models.Category) []CategoryResponse {
	responses := make([]CategoryResponse, len(categories))
	for i, category := range categories {
		responses[i] = ToCategoryResponse(category)
	}
	return responses
}
