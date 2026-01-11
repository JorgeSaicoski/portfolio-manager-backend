package category

import (
	"context"
	"fmt"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/contracts"
)

// UpdateCategoryPositionUseCase handles the business logic for updating a single category position
type UpdateCategoryPositionUseCase struct {
	categoryRepo  contracts.CategoryRepository
	portfolioRepo contracts.PortfolioRepository
	auditLogger   contracts.AuditLogger
}

// NewUpdateCategoryPositionUseCase creates a new instance of UpdateCategoryPositionUseCase
func NewUpdateCategoryPositionUseCase(
	categoryRepo contracts.CategoryRepository,
	portfolioRepo contracts.PortfolioRepository,
	auditLogger contracts.AuditLogger,
) *UpdateCategoryPositionUseCase {
	return &UpdateCategoryPositionUseCase{
		categoryRepo:  categoryRepo,
		portfolioRepo: portfolioRepo,
		auditLogger:   auditLogger,
	}
}

// Execute updates a category's position with ownership verification
func (uc *UpdateCategoryPositionUseCase) Execute(ctx context.Context, id uint, position uint, ownerID string) error {
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

	// Update position
	if err := uc.categoryRepo.UpdatePosition(ctx, id, position); err != nil {
		return fmt.Errorf("failed to update category position: %w", err)
	}

	// Audit logging
	if uc.auditLogger != nil {
		uc.auditLogger.LogUpdate(ctx, "category", id, map[string]interface{}{
			"position":     position,
			"portfolio_id": category.PortfolioID,
			"owner_id":     ownerID,
		})
	}

	return nil
}
