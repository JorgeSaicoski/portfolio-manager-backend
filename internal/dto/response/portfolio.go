package response

import (
	"time"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/models"
)

// PortfolioResponse represents a basic portfolio in responses
type PortfolioResponse struct {
	ID          uint       `json:"id"`
	Title       string     `json:"title"`
	Description *string    `json:"description,omitempty"`
	OwnerID     string     `json:"owner_id,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

// PortfolioDetailResponse represents a detailed portfolio with relationships
type PortfolioDetailResponse struct {
	ID          uint               `json:"id"`
	Title       string             `json:"title"`
	Description *string            `json:"description,omitempty"`
	OwnerID     string             `json:"owner_id,omitempty"`
	Sections    []SectionResponse  `json:"sections,omitempty"`
	Categories  []CategoryResponse `json:"categories,omitempty"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
	DeletedAt   *time.Time         `json:"deleted_at,omitempty"`
}

// ToPortfolioResponse converts a model to a basic response DTO
func ToPortfolioResponse(portfolio *models.Portfolio) PortfolioResponse {
	return PortfolioResponse{
		ID:          portfolio.ID,
		Title:       portfolio.Title,
		Description: portfolio.Description,
		OwnerID:     portfolio.OwnerID,
		CreatedAt:   portfolio.CreatedAt,
		UpdatedAt:   portfolio.UpdatedAt,
		DeletedAt:   nil,
	}
}

// ToPortfolioDetailResponse converts a model with relations to a detailed response DTO
func ToPortfolioDetailResponse(portfolio *models.Portfolio) PortfolioDetailResponse {
	sections := make([]SectionResponse, len(portfolio.Sections))
	for i, section := range portfolio.Sections {
		sections[i] = ToSectionResponse(&section)
	}

	categories := make([]CategoryResponse, len(portfolio.Categories))
	for i, category := range portfolio.Categories {
		categories[i] = ToCategoryResponse(&category)
	}

	return PortfolioDetailResponse{
		ID:          portfolio.ID,
		Title:       portfolio.Title,
		Description: portfolio.Description,
		OwnerID:     portfolio.OwnerID,
		Sections:    sections,
		Categories:  categories,
		CreatedAt:   portfolio.CreatedAt,
		UpdatedAt:   portfolio.UpdatedAt,
		DeletedAt:   nil,
	}
}

// ToPortfolioListResponse converts a slice of models to response DTOs
func ToPortfolioListResponse(portfolios []*models.Portfolio) []PortfolioResponse {
	responses := make([]PortfolioResponse, len(portfolios))
	for i, portfolio := range portfolios {
		responses[i] = ToPortfolioResponse(portfolio)
	}
	return responses
}
