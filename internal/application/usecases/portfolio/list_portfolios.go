package portfolio

import (
	"context"
	"fmt"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/contracts"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/dto"
)

// ListPortfoliosUseCase handles the business logic for listing portfolios
type ListPortfoliosUseCase struct {
	portfolioRepo contracts.PortfolioRepository
}

// NewListPortfoliosUseCase creates a new instance of ListPortfoliosUseCase
func NewListPortfoliosUseCase(portfolioRepo contracts.PortfolioRepository) *ListPortfoliosUseCase {
	return &ListPortfoliosUseCase{
		portfolioRepo: portfolioRepo,
	}
}

// Execute retrieves portfolios owned by a user with pagination
func (uc *ListPortfoliosUseCase) Execute(ctx context.Context, input dto.ListPortfoliosInput) (*dto.ListPortfoliosOutput, error) {
	// Validate input
	if input.OwnerID == "" {
		return nil, fmt.Errorf("owner ID is required")
	}

	// Set default pagination if not provided
	if input.Pagination.Limit == 0 {
		input.Pagination.Limit = 10
	}
	if input.Pagination.Page == 0 {
		input.Pagination.Page = 1
	}

	// Get portfolios from repository
	portfolios, total, err := uc.portfolioRepo.GetByOwnerID(ctx, input.OwnerID, input.Pagination)
	if err != nil {
		return nil, fmt.Errorf("failed to list portfolios: %w", err)
	}

	return &dto.ListPortfoliosOutput{
		Portfolios: portfolios,
		Pagination: dto.PaginatedResultDTO{
			Total: total,
			Page:  input.Pagination.Page,
			Limit: input.Pagination.Limit,
		},
	}, nil
}
