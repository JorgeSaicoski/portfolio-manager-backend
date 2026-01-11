package category

import (
	"context"
	"fmt"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/contracts"
)

// DeleteCategoryUseCase handles the business logic for deleting a category
type DeleteCategoryUseCase struct {
	categoryRepo  contracts.CategoryRepository
	portfolioRepo contracts.PortfolioRepository
	auditLogger   contracts.AuditLogger
	metrics       contracts.MetricsCollector
}

// NewDeleteCategoryUseCase creates a new instance of DeleteCategoryUseCase
func NewDeleteCategoryUseCase(
	categoryRepo contracts.CategoryRepository,
	portfolioRepo contracts.PortfolioRepository,
	auditLogger contracts.AuditLogger,
	metrics contracts.MetricsCollector,
) *DeleteCategoryUseCase {
	return &DeleteCategoryUseCase{
		categoryRepo:  categoryRepo,
		portfolioRepo: portfolioRepo,
		auditLogger:   auditLogger,
		metrics:       metrics,
	}
}

// Execute deletes a category with ownership verification
func (uc *DeleteCategoryUseCase) Execute(ctx context.Context, id uint, ownerID string) error {
	if id == 0 {
		return fmt.Errorf("invalid category ID")
	}
	if ownerID == "" {
		return fmt.Errorf("owner ID is required")
	}

	// Verify category exists and user owns it
	category, err := uc.categoryRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("category not found")
	}

	// Verify ownership through portfolio
	portfolio, err := uc.portfolioRepo.GetByID(ctx, category.PortfolioID)
	if err != nil {
		return fmt.Errorf("portfolio not found")
	}
	if portfolio.OwnerID != ownerID {
		return fmt.Errorf("unauthorized: you don't own this category")
	}

	// Delete the category
	if err := uc.categoryRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	// Audit logging
	if uc.auditLogger != nil {
		uc.auditLogger.LogDelete(ctx, "category", id, map[string]interface{}{
			"title":        category.Title,
			"portfolio_id": category.PortfolioID,
			"owner_id":     ownerID,
		})
	}

	// Metrics
	if uc.metrics != nil {
		uc.metrics.IncrementCategoriesDeleted()
	}

	return nil
}
