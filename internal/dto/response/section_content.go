package response

import (
	"time"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/models"
)

// SectionContentResponse represents a section content block in responses
type SectionContentResponse struct {
	ID        uint       `json:"id"`
	SectionID uint       `json:"section_id"`
	Type      string     `json:"type"`
	Content   string     `json:"content"`
	Order     uint       `json:"order"`
	Metadata  *string    `json:"metadata,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// ToSectionContentResponse converts a model to a response DTO
func ToSectionContentResponse(content *models.SectionContent) SectionContentResponse {
	return SectionContentResponse{
		ID:        content.ID,
		SectionID: content.SectionID,
		Type:      content.Type,
		Content:   content.Content,
		Order:     content.Order,
		Metadata:  content.Metadata,
		CreatedAt: content.CreatedAt,
		UpdatedAt: content.UpdatedAt,
		DeletedAt: nil,
	}
}

// ToSectionContentListResponse converts a slice of models to response DTOs
func ToSectionContentListResponse(contents []models.SectionContent) []SectionContentResponse {
	responses := make([]SectionContentResponse, 0, len(contents))
	for i := range contents {
		responses = append(responses, ToSectionContentResponse(&contents[i]))
	}
	return responses
}
