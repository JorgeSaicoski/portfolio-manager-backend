package section_content

import (
	"context"
	"fmt"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/contracts"
)

// DeleteSectionContentUseCase handles the business logic for deleting a section content
type DeleteSectionContentUseCase struct {
	contentRepo   contracts.SectionContentRepository
	sectionRepo   contracts.SectionRepository
	portfolioRepo contracts.PortfolioRepository
	auditLogger   contracts.AuditLogger
}

// NewDeleteSectionContentUseCase creates a new instance of DeleteSectionContentUseCase
func NewDeleteSectionContentUseCase(
	contentRepo contracts.SectionContentRepository,
	sectionRepo contracts.SectionRepository,
	portfolioRepo contracts.PortfolioRepository,
	auditLogger contracts.AuditLogger,
) *DeleteSectionContentUseCase {
	return &DeleteSectionContentUseCase{
		contentRepo:   contentRepo,
		sectionRepo:   sectionRepo,
		portfolioRepo: portfolioRepo,
		auditLogger:   auditLogger,
	}
}

// Execute deletes a section content with ownership verification
func (uc *DeleteSectionContentUseCase) Execute(ctx context.Context, id uint, ownerID string) error {
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

	// Delete the section content
	if err := uc.contentRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete section content: %w", err)
	}

	// Audit logging
	if uc.auditLogger != nil {
		uc.auditLogger.LogDelete(ctx, "section_content", id, map[string]interface{}{
			"section_id": content.SectionID,
			"type":       content.Type,
			"owner_id":   ownerID,
		})
	}

	return nil
}
