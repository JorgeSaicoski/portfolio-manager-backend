package category

import (
	"context"
	"fmt"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/contracts"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/dto"
)

// GetCategoryPublicUseCase handles the business logic for retrieving a category publicly (no auth)
type GetCategoryPublicUseCase struct {
	categoryRepo contracts.CategoryRepository
}

// NewGetCategoryPublicUseCase creates a new instance of GetCategoryPublicUseCase
func NewGetCategoryPublicUseCase(
	categoryRepo contracts.CategoryRepository,
) *GetCategoryPublicUseCase {
	return &GetCategoryPublicUseCase{
		categoryRepo: categoryRepo,
	}
}

// Execute retrieves a category by ID without ownership verification (public access)
func (uc *GetCategoryPublicUseCase) Execute(ctx context.Context, id uint) (*dto.CategoryDTO, error) {
	if id == 0 {
		return nil, fmt.Errorf("invalid category ID")
	}

	// Get category (no ownership check for public access)
	category, err := uc.categoryRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("category not found")
	}

	return category, nil
}
