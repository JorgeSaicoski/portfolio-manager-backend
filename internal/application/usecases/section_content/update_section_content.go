package section_content

import (
	"context"
	"fmt"

	contracts2 "github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/contracts"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/dto"
)

// UpdateSectionContentUseCase handles the business logic for updating a section content
type UpdateSectionContentUseCase struct {
	contentRepo   contracts2.SectionContentRepository
	sectionRepo   contracts2.SectionRepository
	portfolioRepo contracts2.PortfolioRepository
	auditLogger   contracts2.AuditLogger
}

// NewUpdateSectionContentUseCase creates a new instance of UpdateSectionContentUseCase
func NewUpdateSectionContentUseCase(
	contentRepo contracts2.SectionContentRepository,
	sectionRepo contracts2.SectionRepository,
	portfolioRepo contracts2.PortfolioRepository,
	auditLogger contracts2.AuditLogger,
) *UpdateSectionContentUseCase {
	return &UpdateSectionContentUseCase{
		contentRepo:   contentRepo,
		sectionRepo:   sectionRepo,
		portfolioRepo: portfolioRepo,
		auditLogger:   auditLogger,
	}
}

// Execute updates an existing section content
func (uc *UpdateSectionContentUseCase) Execute(ctx context.Context, input dto.UpdateSectionContentInput) error {
	// Validate input
	if input.ID == 0 {
		return fmt.Errorf("section content ID is required")
	}
	if input.OwnerID == "" {
		return fmt.Errorf("owner ID is required")
	}

	// Verify section content exists
	content, err := uc.contentRepo.GetByID(ctx, input.ID)
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
	if portfolio.OwnerID != input.OwnerID {
		return fmt.Errorf("unauthorized: you don't own this section content")
	}

	// Update the section content
	if err := uc.contentRepo.Update(ctx, input); err != nil {
		return fmt.Errorf("failed to update section content: %w", err)
	}

	// Audit logging
	if uc.auditLogger != nil {
		uc.auditLogger.LogUpdate(ctx, "section_content", content.ID, map[string]interface{}{
			"section_id": content.SectionID,
			"type":       input.Type,
			"owner_id":   input.OwnerID,
		})
	}

	return nil
}
