package portfolio

import (
	"context"
	"fmt"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/contracts"
)

// DeletePortfolioUseCase handles the business logic for deleting a portfolio
type DeletePortfolioUseCase struct {
	portfolioRepo contracts.PortfolioRepository
	auditLogger   contracts.AuditLogger
	metrics       contracts.MetricsCollector
}

// NewDeletePortfolioUseCase creates a new instance of DeletePortfolioUseCase
func NewDeletePortfolioUseCase(
	portfolioRepo contracts.PortfolioRepository,
	auditLogger contracts.AuditLogger,
	metrics contracts.MetricsCollector,
) *DeletePortfolioUseCase {
	return &DeletePortfolioUseCase{
		portfolioRepo: portfolioRepo,
		auditLogger:   auditLogger,
		metrics:       metrics,
	}
}

// Execute deletes a portfolio
func (uc *DeletePortfolioUseCase) Execute(ctx context.Context, id uint, ownerID string) error {
	// 1. Validate input
	if id == 0 {
		return fmt.Errorf("portfolio ID is required")
	}
	if ownerID == "" {
		return fmt.Errorf("owner ID is required")
	}

	// 2. Get portfolio to verify it exists and check ownership
	existing, err := uc.portfolioRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("portfolio not found: %w", err)
	}

	// 3. Authorization check - verify ownership
	if existing.OwnerID != ownerID {
		if uc.auditLogger != nil {
			uc.auditLogger.LogAccess(ctx, "portfolio", id, ownerID, false)
		}
		return fmt.Errorf("unauthorized: you don't own this portfolio")
	}

	// 4. Delete portfolio via repository
	if err := uc.portfolioRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete portfolio: %w", err)
	}

	// 5. Audit log
	if uc.auditLogger != nil {
		uc.auditLogger.LogDelete(ctx, "portfolio", id, map[string]interface{}{
			"title":   existing.Title,
			"ownerID": existing.OwnerID,
		})
	}

	// 6. Update metrics
	if uc.metrics != nil {
		uc.metrics.IncrementPortfoliosDeleted()
	}

	return nil
}
