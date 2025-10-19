package response

import (
	"time"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/models"
)

// SectionResponse represents a section in responses
type SectionResponse struct {
	ID          uint       `json:"id"`
	Title       string     `json:"title"`
	Description *string    `json:"description,omitempty"`
	Type        string     `json:"type"`
	OwnerID     string     `json:"owner_id,omitempty"`
	PortfolioID uint       `json:"portfolio_id"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

// ToSectionResponse converts a model to a response DTO
func ToSectionResponse(section *models.Section) SectionResponse {
	return SectionResponse{
		ID:          section.ID,
		Title:       section.Title,
		Description: section.Description,
		Type:        section.Type,
		OwnerID:     section.OwnerID,
		PortfolioID: section.PortfolioID,
		CreatedAt:   section.CreatedAt,
		UpdatedAt:   section.UpdatedAt,
		DeletedAt:   nil,
	}
}

// ToSectionListResponse converts a slice of models to response DTOs
func ToSectionListResponse(sections []*models.Section) []SectionResponse {
	responses := make([]SectionResponse, len(sections))
	for i, section := range sections {
		responses[i] = ToSectionResponse(section)
	}
	return responses
}
