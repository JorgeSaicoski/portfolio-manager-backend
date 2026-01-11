package contracts

import (
	"context"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/dto"
)

// PortfolioRepository defines the interface for portfolio data persistence
// This is a contract in the application layer that the infrastructure layer must implement
type PortfolioRepository interface {
	// Create creates a new portfolio and returns the created portfolio DTO
	Create(ctx context.Context, input dto.CreatePortfolioInput) (*dto.PortfolioDTO, error)

	// GetByID retrieves a portfolio by its ID
	GetByID(ctx context.Context, id uint) (*dto.PortfolioDTO, error)

	// GetByOwnerID retrieves all portfolios owned by a specific user with pagination
	// Returns the list of portfolios, total count, and any error
	GetByOwnerID(ctx context.Context, ownerID string, pagination dto.PaginationDTO) ([]dto.PortfolioDTO, int64, error)

	// Update updates an existing portfolio
	Update(ctx context.Context, input dto.UpdatePortfolioInput) error

	// Delete deletes a portfolio by its ID
	Delete(ctx context.Context, id uint) error

	// CheckTitleDuplicate checks if a portfolio title already exists for a user
	// excludeID is used when updating to exclude the current portfolio from the check (pass 0 when creating)
	CheckTitleDuplicate(ctx context.Context, title, ownerID string, excludeID uint) (bool, error)
}
