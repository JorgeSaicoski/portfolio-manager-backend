package section

import (
	"context"
	"fmt"

	contracts2 "github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/contracts"
)

// DeleteSectionUseCase handles the business logic for deleting a section
type DeleteSectionUseCase struct {
	sectionRepo   contracts2.SectionRepository
	portfolioRepo contracts2.PortfolioRepository
	auditLogger   contracts2.AuditLogger
	metrics       contracts2.MetricsCollector
}

// NewDeleteSectionUseCase creates a new instance of DeleteSectionUseCase
func NewDeleteSectionUseCase(
	sectionRepo contracts2.SectionRepository,
	portfolioRepo contracts2.PortfolioRepository,
	auditLogger contracts2.AuditLogger,
	metrics contracts2.MetricsCollector,
) *DeleteSectionUseCase {
	return &DeleteSectionUseCase{
		sectionRepo:   sectionRepo,
		portfolioRepo: portfolioRepo,
		auditLogger:   auditLogger,
		metrics:       metrics,
	}
}

// Execute deletes a section with ownership verification
func (uc *DeleteSectionUseCase) Execute(ctx context.Context, id uint, ownerID string) error {
	if id == 0 {
		return fmt.Errorf("invalid section ID")
	}
	if ownerID == "" {
		return fmt.Errorf("owner ID is required")
	}

	// Verify section exists and user owns it
	section, err := uc.sectionRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("section not found")
	}

	// Verify ownership through portfolio
	portfolio, err := uc.portfolioRepo.GetByID(ctx, section.PortfolioID)
	if err != nil {
		return fmt.Errorf("portfolio not found")
	}
	if portfolio.OwnerID != ownerID {
		return fmt.Errorf("unauthorized: you don't own this section")
	}

	// Delete the section
	if err := uc.sectionRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete section: %w", err)
	}

	// Audit logging
	if uc.auditLogger != nil {
		uc.auditLogger.LogDelete(ctx, "section", id, map[string]interface{}{
			"title":        section.Title,
			"type":         section.Type,
			"portfolio_id": section.PortfolioID,
			"owner_id":     ownerID,
		})
	}

	// Metrics
	if uc.metrics != nil {
		uc.metrics.IncrementSectionsDeleted()
	}

	return nil
}
