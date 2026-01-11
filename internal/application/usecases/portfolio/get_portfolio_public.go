package portfolio

import (
	"context"
	"fmt"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/contracts"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/dto"
)

// GetPortfolioPublicUseCase handles the business logic for retrieving a portfolio publicly (no auth)
type GetPortfolioPublicUseCase struct {
	portfolioRepo contracts.PortfolioRepository
}

// NewGetPortfolioPublicUseCase creates a new instance of GetPortfolioPublicUseCase
func NewGetPortfolioPublicUseCase(
	portfolioRepo contracts.PortfolioRepository,
) *GetPortfolioPublicUseCase {
	return &GetPortfolioPublicUseCase{
		portfolioRepo: portfolioRepo,
	}
}

// Execute retrieves a portfolio by ID without ownership verification (public access)
func (uc *GetPortfolioPublicUseCase) Execute(ctx context.Context, id uint) (*dto.PortfolioDTO, error) {
	if id == 0 {
		return nil, fmt.Errorf("invalid portfolio ID")
	}

	// Get portfolio (no ownership check for public access)
	portfolio, err := uc.portfolioRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("portfolio not found")
	}

	return portfolio, nil
}
