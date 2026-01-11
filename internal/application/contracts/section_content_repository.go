package contracts

import (
	"context"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/dto"
)

// SectionContentRepository defines the interface for section content data persistence
type SectionContentRepository interface {
	// Create creates a new section content and returns the created content DTO
	Create(ctx context.Context, input dto.CreateSectionContentInput) (*dto.SectionContentDTO, error)

	// GetByID retrieves a section content by its ID
	GetByID(ctx context.Context, id uint) (*dto.SectionContentDTO, error)

	// GetBySectionID retrieves all section contents for a specific section (ordered by order field)
	GetBySectionID(ctx context.Context, sectionID uint) ([]dto.SectionContentDTO, error)

	// Update updates an existing section content
	Update(ctx context.Context, input dto.UpdateSectionContentInput) error

	// UpdateOrder updates only the order field of a section content
	UpdateOrder(ctx context.Context, id uint, order uint) error

	// Delete deletes a section content by its ID
	Delete(ctx context.Context, id uint) error
}
