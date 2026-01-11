package category

import (
	"context"
	"fmt"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/contracts"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/dto"
)

// BulkReorderCategoriesUseCase handles the business logic for bulk reordering categories
type BulkReorderCategoriesUseCase struct {
	categoryRepo  contracts.CategoryRepository
	portfolioRepo contracts.PortfolioRepository
	auditLogger   contracts.AuditLogger
}

// NewBulkReorderCategoriesUseCase creates a new instance of BulkReorderCategoriesUseCase
func NewBulkReorderCategoriesUseCase(
	categoryRepo contracts.CategoryRepository,
	portfolioRepo contracts.PortfolioRepository,
	auditLogger contracts.AuditLogger,
) *BulkReorderCategoriesUseCase {
	return &BulkReorderCategoriesUseCase{
		categoryRepo:  categoryRepo,
		portfolioRepo: portfolioRepo,
		auditLogger:   auditLogger,
	}
}

// Execute performs bulk position updates with ownership verification
func (uc *BulkReorderCategoriesUseCase) Execute(ctx context.Context, input dto.BulkUpdateCategoryPositionsInput) error {
	if len(input.Items) == 0 {
		return fmt.Errorf("no categories to reorder")
	}
	if input.OwnerID == "" {
		return fmt.Errorf("owner ID is required")
	}

	// Verify ownership of all categories
	categoryIDs := make([]uint, len(input.Items))
	for i, item := range input.Items {
		categoryIDs[i] = item.ID
	}

	categories, err := uc.categoryRepo.GetByIDs(ctx, categoryIDs)
	if err != nil {
		return fmt.Errorf("failed to retrieve categories: %w", err)
	}
	if len(categories) != len(categoryIDs) {
		return fmt.Errorf("some categories not found")
	}

	// Verify all categories belong to portfolios owned by the user
	portfolioIDSet := make(map[uint]bool)
	for _, cat := range categories {
		portfolioIDSet[cat.PortfolioID] = true
	}

	for portfolioID := range portfolioIDSet {
		portfolio, err := uc.portfolioRepo.GetByID(ctx, portfolioID)
		if err != nil {
			return fmt.Errorf("portfolio not found")
		}
		if portfolio.OwnerID != input.OwnerID {
			return fmt.Errorf("unauthorized: you don't own all categories")
		}
	}

	// Perform bulk update
	if err := uc.categoryRepo.BulkUpdatePositions(ctx, input); err != nil {
		return fmt.Errorf("failed to reorder categories: %w", err)
	}

	// Audit logging
	if uc.auditLogger != nil {
		uc.auditLogger.LogUpdate(ctx, "category", 0, map[string]interface{}{
			"operation": "bulk_reorder",
			"count":     len(input.Items),
			"owner_id":  input.OwnerID,
		})
	}

	return nil
}
