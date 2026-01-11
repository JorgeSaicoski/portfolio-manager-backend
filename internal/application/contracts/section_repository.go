package contracts

import (
	"context"

	dto2 "github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/dto"
)

// SectionRepository defines the interface for section data persistence
// This is a contract in the application layer that the infrastructure layer must implement
type SectionRepository interface {
	// Create creates a new section and returns the created section DTO
	Create(ctx context.Context, input dto2.CreateSectionInput) (*dto2.SectionDTO, error)

	// GetByID retrieves a section by its ID (basic info only)
	GetByID(ctx context.Context, id uint) (*dto2.SectionDTO, error)

	// GetByIDs retrieves multiple sections by their IDs
	GetByIDs(ctx context.Context, ids []uint) ([]dto2.SectionDTO, error)

	// GetByPortfolioID retrieves all sections for a specific portfolio (ordered by position)
	GetByPortfolioID(ctx context.Context, portfolioID uint) ([]dto2.SectionDTO, error)

	// GetByOwnerID retrieves all sections owned by a specific user with pagination
	// Returns the list of sections, total count, and any error
	GetByOwnerID(ctx context.Context, ownerID string, pagination dto2.PaginationDTO) ([]dto2.SectionDTO, int64, error)

	// GetByType retrieves all sections of a specific type
	GetByType(ctx context.Context, sectionType string) ([]dto2.SectionDTO, error)

	// Update updates an existing section
	Update(ctx context.Context, input dto2.UpdateSectionInput) error

	// UpdatePosition updates only the position field of a section
	UpdatePosition(ctx context.Context, id uint, position uint) error

	// BulkUpdatePositions updates positions for multiple sections in a transaction
	BulkUpdatePositions(ctx context.Context, input dto2.BulkUpdateSectionPositionsInput) error

	// Delete deletes a section by its ID (cascade deletes section contents)
	Delete(ctx context.Context, id uint) error

	// CheckTitleDuplicate checks if a section title already exists for a portfolio
	// excludeID is used when updating to exclude the current section from the check (pass 0 when creating)
	CheckTitleDuplicate(ctx context.Context, title string, portfolioID uint, excludeID uint) (bool, error)
}
