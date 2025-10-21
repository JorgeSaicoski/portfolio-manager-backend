package response

import (
	"time"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/models"
)

// SectionResponse represents a section in responses
type SectionResponse struct {
	ID          uint                     `json:"id"`
	Title       string                   `json:"title"`
	Description *string                  `json:"description,omitempty"`
	Type        string                   `json:"type"`
	OwnerID     string                   `json:"owner_id,omitempty"`
	PortfolioID uint                     `json:"portfolio_id"`
	Contents    []SectionContentResponse `json:"contents,omitempty"`
	CreatedAt   time.Time                `json:"created_at"`
	UpdatedAt   time.Time                `json:"updated_at"`
	DeletedAt   *time.Time               `json:"deleted_at,omitempty"`
}

// ToSectionResponse converts a model to a response DTO
func ToSectionResponse(section *models.Section) SectionResponse {
	var contents []SectionContentResponse
	if len(section.Contents) > 0 {
		contents = ToSectionContentListResponse(section.Contents)
	}

	return SectionResponse{
		ID:          section.ID,
		Title:       section.Title,
		Description: section.Description,
		Type:        section.Type,
		OwnerID:     section.OwnerID,
		PortfolioID: section.PortfolioID,
		Contents:    contents,
		CreatedAt:   section.CreatedAt,
		UpdatedAt:   section.UpdatedAt,
		DeletedAt:   nil,
	}
}
