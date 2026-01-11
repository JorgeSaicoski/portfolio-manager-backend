package category

import (
	"context"
	"fmt"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/contracts"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/dto"
)

// UpdateCategoryUseCase handles the business logic for updating a category
type UpdateCategoryUseCase struct {
	categoryRepo  contracts.CategoryRepository
	portfolioRepo contracts.PortfolioRepository
	auditLogger   contracts.AuditLogger
	metrics       contracts.MetricsCollector
}

// NewUpdateCategoryUseCase creates a new instance of UpdateCategoryUseCase
func NewUpdateCategoryUseCase(
	categoryRepo contracts.CategoryRepository,
	portfolioRepo contracts.PortfolioRepository,
	auditLogger contracts.AuditLogger,
	metrics contracts.MetricsCollector,
) *UpdateCategoryUseCase {
	return &UpdateCategoryUseCase{
		categoryRepo:  categoryRepo,
		portfolioRepo: portfolioRepo,
		auditLogger:   auditLogger,
		metrics:       metrics,
	}
}

// Execute updates a category with ownership verification
func (uc *UpdateCategoryUseCase) Execute(ctx context.Context, input dto.UpdateCategoryInput) error {
	// Validate input
	if input.ID == 0 {
		return fmt.Errorf("invalid category ID")
	}
	if input.OwnerID == "" {
		return fmt.Errorf("owner ID is required")
	}
	if input.Title == "" {
		return fmt.Errorf("category title is required")
	}

	// Verify category exists and user owns it
	category, err := uc.categoryRepo.GetByID(ctx, input.ID)
	if err != nil {
		return fmt.Errorf("category not found")
	}

	// Verify ownership through portfolio
	portfolio, err := uc.portfolioRepo.GetByID(ctx, category.PortfolioID)
	if err != nil {
		return fmt.Errorf("portfolio not found")
	}
	if portfolio.OwnerID != input.OwnerID {
		return fmt.Errorf("unauthorized: you don't own this category")
	}

	// Update the category
	if err := uc.categoryRepo.Update(ctx, input); err != nil {
		return fmt.Errorf("failed to update category: %w", err)
	}

	// Audit logging
	if uc.auditLogger != nil {
		uc.auditLogger.LogUpdate(ctx, "category", input.ID, map[string]interface{}{
			"title":        input.Title,
			"description":  input.Description,
			"position":     input.Position,
			"portfolio_id": category.PortfolioID,
			"owner_id":     input.OwnerID,
		})
	}

	// Metrics
	if uc.metrics != nil {
		uc.metrics.IncrementCategoriesUpdated()
	}

	return nil
}
