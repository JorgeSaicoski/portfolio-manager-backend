package category

import (
	"context"
	"fmt"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/contracts"
	dto2 "github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/dto"
)

// ListCategoriesUseCase handles the business logic for listing categories
type ListCategoriesUseCase struct {
	categoryRepo contracts.CategoryRepository
}

// NewListCategoriesUseCase creates a new instance of ListCategoriesUseCase
func NewListCategoriesUseCase(
	categoryRepo contracts.CategoryRepository,
) *ListCategoriesUseCase {
	return &ListCategoriesUseCase{
		categoryRepo: categoryRepo,
	}
}

// Execute retrieves all categories owned by a user with pagination
func (uc *ListCategoriesUseCase) Execute(ctx context.Context, input dto2.ListCategoriesInput) (*dto2.ListCategoriesOutput, error) {
	// Get categories with pagination
	categories, total, err := uc.categoryRepo.GetByOwnerID(ctx, input.PortfolioID, input.Pagination)
	if err != nil {
		return nil, fmt.Errorf("failed to list categories: %w", err)
	}

	return &dto2.ListCategoriesOutput{
		Categories: categories,
		Pagination: dto2.PaginatedResultDTO{
			Total: total,
			Page:  input.Pagination.Page,
			Limit: input.Pagination.Limit,
		},
	}, nil
}
