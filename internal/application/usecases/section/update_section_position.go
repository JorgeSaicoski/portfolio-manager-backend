package section

import (
	"context"
	"fmt"

	contracts2 "github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/contracts"
)

// UpdateSectionPositionUseCase handles the business logic for updating a single section position
type UpdateSectionPositionUseCase struct {
	sectionRepo   contracts2.SectionRepository
	portfolioRepo contracts2.PortfolioRepository
	auditLogger   contracts2.AuditLogger
}

// NewUpdateSectionPositionUseCase creates a new instance of UpdateSectionPositionUseCase
func NewUpdateSectionPositionUseCase(
	sectionRepo contracts2.SectionRepository,
	portfolioRepo contracts2.PortfolioRepository,
	auditLogger contracts2.AuditLogger,
) *UpdateSectionPositionUseCase {
	return &UpdateSectionPositionUseCase{
		sectionRepo:   sectionRepo,
		portfolioRepo: portfolioRepo,
		auditLogger:   auditLogger,
	}
}

// Execute updates a section's position with ownership verification
func (uc *UpdateSectionPositionUseCase) Execute(ctx context.Context, id uint, position uint, ownerID string) error {
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

	// Update position
	if err := uc.sectionRepo.UpdatePosition(ctx, id, position); err != nil {
		return fmt.Errorf("failed to update section position: %w", err)
	}

	// Audit logging
	if uc.auditLogger != nil {
		uc.auditLogger.LogUpdate(ctx, "section", id, map[string]interface{}{
			"position":     position,
			"portfolio_id": section.PortfolioID,
			"owner_id":     ownerID,
		})
	}

	return nil
}
