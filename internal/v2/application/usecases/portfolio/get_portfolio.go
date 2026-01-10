package portfolio

import (
	"context"
	"fmt"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/contracts"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/dto"
)

// GetPortfolioUseCase handles the business logic for retrieving a portfolio
type GetPortfolioUseCase struct {
	portfolioRepo contracts.PortfolioRepository
}

// NewGetPortfolioUseCase creates a new instance of GetPortfolioUseCase
func NewGetPortfolioUseCase(portfolioRepo contracts.PortfolioRepository) *GetPortfolioUseCase {
	return &GetPortfolioUseCase{
		portfolioRepo: portfolioRepo,
	}
}

// Execute retrieves a portfolio by ID
func (uc *GetPortfolioUseCase) Execute(ctx context.Context, id uint) (*dto.PortfolioDTO, error) {
	// Validate input
	if id == 0 {
		return nil, fmt.Errorf("portfolio ID is required")
	}

	// Get portfolio from repository
	portfolio, err := uc.portfolioRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("portfolio not found: %w", err)
	}

	return portfolio, nil
}
