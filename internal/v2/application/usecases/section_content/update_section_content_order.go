package section_content

import (
	"context"
	"fmt"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/contracts"
)

// UpdateSectionContentOrderUseCase handles the business logic for updating a section content's order
type UpdateSectionContentOrderUseCase struct {
	contentRepo   contracts.SectionContentRepository
	sectionRepo   contracts.SectionRepository
	portfolioRepo contracts.PortfolioRepository
	auditLogger   contracts.AuditLogger
}

// NewUpdateSectionContentOrderUseCase creates a new instance of UpdateSectionContentOrderUseCase
func NewUpdateSectionContentOrderUseCase(
	contentRepo contracts.SectionContentRepository,
	sectionRepo contracts.SectionRepository,
	portfolioRepo contracts.PortfolioRepository,
	auditLogger contracts.AuditLogger,
) *UpdateSectionContentOrderUseCase {
	return &UpdateSectionContentOrderUseCase{
		contentRepo:   contentRepo,
		sectionRepo:   sectionRepo,
		portfolioRepo: portfolioRepo,
		auditLogger:   auditLogger,
	}
}

// Execute updates the order of a section content
func (uc *UpdateSectionContentOrderUseCase) Execute(ctx context.Context, id uint, order uint, ownerID string) error {
	// Validate input
	if id == 0 {
		return fmt.Errorf("section content ID is required")
	}
	if ownerID == "" {
		return fmt.Errorf("owner ID is required")
	}

	// Verify section content exists
	content, err := uc.contentRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("section content not found")
	}

	// Verify ownership through section and portfolio
	section, err := uc.sectionRepo.GetByID(ctx, content.SectionID)
	if err != nil {
		return fmt.Errorf("section not found")
	}

	portfolio, err := uc.portfolioRepo.GetByID(ctx, section.PortfolioID)
	if err != nil {
		return fmt.Errorf("portfolio not found")
	}
	if portfolio.OwnerID != ownerID {
		return fmt.Errorf("unauthorized: you don't own this section content")
	}

	// Update the order
	if err := uc.contentRepo.UpdateOrder(ctx, id, order); err != nil {
		return fmt.Errorf("failed to update section content order: %w", err)
	}

	// Audit logging
	if uc.auditLogger != nil {
		uc.auditLogger.LogUpdate(ctx, "section_content", id, map[string]interface{}{
			"section_id": content.SectionID,
			"order":      order,
			"owner_id":   ownerID,
		})
	}

	return nil
}
