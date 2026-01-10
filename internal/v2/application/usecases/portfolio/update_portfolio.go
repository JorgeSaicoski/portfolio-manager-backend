package portfolio

import (
	"context"
	"fmt"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/contracts"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/dto"
)

// UpdatePortfolioUseCase handles the business logic for updating a portfolio
type UpdatePortfolioUseCase struct {
	portfolioRepo contracts.PortfolioRepository
	auditLogger   contracts.AuditLogger
	metrics       contracts.MetricsCollector
}

// NewUpdatePortfolioUseCase creates a new instance of UpdatePortfolioUseCase
func NewUpdatePortfolioUseCase(
	portfolioRepo contracts.PortfolioRepository,
	auditLogger contracts.AuditLogger,
	metrics contracts.MetricsCollector,
) *UpdatePortfolioUseCase {
	return &UpdatePortfolioUseCase{
		portfolioRepo: portfolioRepo,
		auditLogger:   auditLogger,
		metrics:       metrics,
	}
}

// Execute updates an existing portfolio
func (uc *UpdatePortfolioUseCase) Execute(ctx context.Context, input dto.UpdatePortfolioInput) error {
	// 1. Validate input
	if input.ID == 0 {
		return fmt.Errorf("portfolio ID is required")
	}
	if input.OwnerID == "" {
		return fmt.Errorf("owner ID is required")
	}

	// 2. Get existing portfolio
	existing, err := uc.portfolioRepo.GetByID(ctx, input.ID)
	if err != nil {
		return fmt.Errorf("portfolio not found: %w", err)
	}

	// 3. Authorization check - verify ownership
	if existing.OwnerID != input.OwnerID {
		if uc.auditLogger != nil {
			uc.auditLogger.LogAccess(ctx, "portfolio", input.ID, input.OwnerID, false)
		}
		return fmt.Errorf("unauthorized: you don't own this portfolio")
	}

	// 4. Check for duplicate title if title is being changed
	if input.Title != "" && input.Title != existing.Title {
		isDuplicate, err := uc.portfolioRepo.CheckTitleDuplicate(ctx, input.Title, input.OwnerID, input.ID)
		if err != nil {
			return fmt.Errorf("failed to check duplicate title: %w", err)
		}
		if isDuplicate {
			return fmt.Errorf("portfolio with title '%s' already exists for this user", input.Title)
		}
	}

	// 5. Update portfolio via repository
	if err := uc.portfolioRepo.Update(ctx, input); err != nil {
		return fmt.Errorf("failed to update portfolio: %w", err)
	}

	// 6. Audit log
	if uc.auditLogger != nil {
		uc.auditLogger.LogUpdate(ctx, "portfolio", input.ID, map[string]interface{}{
			"title":       input.Title,
			"description": input.Description,
		})
	}

	// 7. Update metrics
	if uc.metrics != nil {
		uc.metrics.IncrementPortfoliosUpdated()
	}

	return nil
}
