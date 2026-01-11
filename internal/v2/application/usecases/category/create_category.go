package category

import (
	"context"
	"fmt"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/contracts"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/dto"
)

// CreateCategoryUseCase handles the business logic for creating a category
type CreateCategoryUseCase struct {
	categoryRepo  contracts.CategoryRepository
	portfolioRepo contracts.PortfolioRepository
	auditLogger   contracts.AuditLogger
	metrics       contracts.MetricsCollector
}

// NewCreateCategoryUseCase creates a new instance of CreateCategoryUseCase
func NewCreateCategoryUseCase(
	categoryRepo contracts.CategoryRepository,
	portfolioRepo contracts.PortfolioRepository,
	auditLogger contracts.AuditLogger,
	metrics contracts.MetricsCollector,
) *CreateCategoryUseCase {
	return &CreateCategoryUseCase{
		categoryRepo:  categoryRepo,
		portfolioRepo: portfolioRepo,
		auditLogger:   auditLogger,
		metrics:       metrics,
	}
}

// Execute creates a new category
func (uc *CreateCategoryUseCase) Execute(ctx context.Context, input dto.CreateCategoryInput) (*dto.CategoryDTO, error) {
	// Validate input
	if input.Title == "" {
		return nil, fmt.Errorf("category title is required")
	}
	if input.OwnerID == "" {
		return nil, fmt.Errorf("owner ID is required")
	}
	if input.PortfolioID == 0 {
		return nil, fmt.Errorf("portfolio ID is required")
	}

	// Verify portfolio exists and user owns it
	portfolio, err := uc.portfolioRepo.GetByID(ctx, input.PortfolioID)
	if err != nil {
		return nil, fmt.Errorf("portfolio not found")
	}
	if portfolio.OwnerID != input.OwnerID {
		return nil, fmt.Errorf("unauthorized: you don't own this portfolio")
	}

	// Create the category
	category, err := uc.categoryRepo.Create(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}

	// Audit logging
	if uc.auditLogger != nil {
		uc.auditLogger.LogCreate(ctx, "category", category.ID, map[string]interface{}{
			"title":        category.Title,
			"portfolio_id": category.PortfolioID,
			"owner_id":     category.OwnerID,
		})
	}

	// Metrics
	if uc.metrics != nil {
		uc.metrics.IncrementCategoriesCreated()
	}

	return category, nil
}
