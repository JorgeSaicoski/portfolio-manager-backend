package portfolio

import (
	"context"
	"fmt"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/contracts"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/dto"
)

// CreatePortfolioUseCase handles the business logic for creating a portfolio
type CreatePortfolioUseCase struct {
	portfolioRepo contracts.PortfolioRepository
	auditLogger   contracts.AuditLogger
	metrics       contracts.MetricsCollector
}

// NewCreatePortfolioUseCase creates a new instance of CreatePortfolioUseCase
func NewCreatePortfolioUseCase(
	portfolioRepo contracts.PortfolioRepository,
	auditLogger contracts.AuditLogger,
	metrics contracts.MetricsCollector,
) *CreatePortfolioUseCase {
	return &CreatePortfolioUseCase{
		portfolioRepo: portfolioRepo,
		auditLogger:   auditLogger,
		metrics:       metrics,
	}
}

// Execute creates a new portfolio
func (uc *CreatePortfolioUseCase) Execute(ctx context.Context, input dto.CreatePortfolioInput) (*dto.PortfolioDTO, error) {
	// 1. Validate input
	if input.Title == "" {
		return nil, fmt.Errorf("title is required")
	}
	if input.OwnerID == "" {
		return nil, fmt.Errorf("owner ID is required")
	}

	// 2. Check for duplicate title
	isDuplicate, err := uc.portfolioRepo.CheckTitleDuplicate(ctx, input.Title, input.OwnerID, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to check duplicate title: %w", err)
	}
	if isDuplicate {
		return nil, fmt.Errorf("portfolio with title '%s' already exists for this user", input.Title)
	}

	// 3. Create portfolio via repository
	portfolio, err := uc.portfolioRepo.Create(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create portfolio: %w", err)
	}

	// 4. Audit log
	if uc.auditLogger != nil {
		uc.auditLogger.LogCreate(ctx, "portfolio", portfolio.ID, map[string]interface{}{
			"title":   portfolio.Title,
			"ownerID": portfolio.OwnerID,
		})
	}

	// 5. Update metrics
	if uc.metrics != nil {
		uc.metrics.IncrementPortfoliosCreated()
	}

	return portfolio, nil
}
