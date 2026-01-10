package contracts

import (
	"context"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/dto"
)

// CategoryRepository defines the interface for category data persistence
// This is a contract in the application layer that the infrastructure layer must implement
type CategoryRepository interface {
	// Create creates a new category and returns the created category DTO
	Create(ctx context.Context, input dto.CreateCategoryInput) (*dto.CategoryDTO, error)

	// GetByID retrieves a category by its ID (basic info only)
	GetByID(ctx context.Context, id uint) (*dto.CategoryDTO, error)

	// GetByIDs retrieves multiple categories by their IDs
	GetByIDs(ctx context.Context, ids []uint) ([]dto.CategoryDTO, error)

	// GetByPortfolioID retrieves all categories for a specific portfolio (ordered by position)
	GetByPortfolioID(ctx context.Context, portfolioID uint) ([]dto.CategoryDTO, error)

	// GetByOwnerID retrieves all categories owned by a specific user with pagination
	// Returns the list of categories, total count, and any error
	GetByOwnerID(ctx context.Context, ownerID string, pagination dto.PaginationDTO) ([]dto.CategoryDTO, int64, error)

	// Update updates an existing category
	Update(ctx context.Context, input dto.UpdateCategoryInput) error

	// UpdatePosition updates only the position field of a category
	UpdatePosition(ctx context.Context, id uint, position uint) error

	// BulkUpdatePositions updates positions for multiple categories in a transaction
	BulkUpdatePositions(ctx context.Context, input dto.BulkUpdateCategoryPositionsInput) error

	// Delete deletes a category by its ID (cascade deletes projects)
	Delete(ctx context.Context, id uint) error
}
