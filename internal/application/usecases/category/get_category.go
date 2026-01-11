package category

import (
	"context"
	"fmt"

	contracts2 "github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/contracts"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/dto"
)

// GetCategoryUseCase handles the business logic for retrieving a category by ID
type GetCategoryUseCase struct {
	categoryRepo  contracts2.CategoryRepository
	portfolioRepo contracts2.PortfolioRepository
	auditLogger   contracts2.AuditLogger
}

// NewGetCategoryUseCase creates a new instance of GetCategoryUseCase
func NewGetCategoryUseCase(
	categoryRepo contracts2.CategoryRepository,
	portfolioRepo contracts2.PortfolioRepository,
	auditLogger contracts2.AuditLogger,
) *GetCategoryUseCase {
	return &GetCategoryUseCase{
		categoryRepo:  categoryRepo,
		portfolioRepo: portfolioRepo,
		auditLogger:   auditLogger,
	}
}

// Execute retrieves a category by ID with ownership verification
func (uc *GetCategoryUseCase) Execute(ctx context.Context, id uint, ownerID string) (*dto.CategoryDTO, error) {
	if id == 0 {
		return nil, fmt.Errorf("invalid category ID")
	}
	if ownerID == "" {
		return nil, fmt.Errorf("owner ID is required")
	}

	// Get category
	category, err := uc.categoryRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("category not found")
	}

	// Verify ownership through portfolio
	portfolio, err := uc.portfolioRepo.GetByID(ctx, category.PortfolioID)
	if err != nil {
		return nil, fmt.Errorf("portfolio not found")
	}
	if portfolio.OwnerID != ownerID {
		// Log unauthorized access attempt
		if uc.auditLogger != nil {
			uc.auditLogger.LogAccess(ctx, "category", id, ownerID, false)
		}
		return nil, fmt.Errorf("unauthorized: you don't own this category")
	}

	// Log authorized access
	if uc.auditLogger != nil {
		uc.auditLogger.LogAccess(ctx, "category", id, ownerID, true)
	}

	return category, nil
}
