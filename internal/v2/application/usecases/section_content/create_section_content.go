package section_content

import (
	"context"
	"fmt"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/contracts"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/dto"
)

// CreateSectionContentUseCase handles the business logic for creating a section content
type CreateSectionContentUseCase struct {
	contentRepo   contracts.SectionContentRepository
	sectionRepo   contracts.SectionRepository
	portfolioRepo contracts.PortfolioRepository
	auditLogger   contracts.AuditLogger
}

// NewCreateSectionContentUseCase creates a new instance of CreateSectionContentUseCase
func NewCreateSectionContentUseCase(
	contentRepo contracts.SectionContentRepository,
	sectionRepo contracts.SectionRepository,
	portfolioRepo contracts.PortfolioRepository,
	auditLogger contracts.AuditLogger,
) *CreateSectionContentUseCase {
	return &CreateSectionContentUseCase{
		contentRepo:   contentRepo,
		sectionRepo:   sectionRepo,
		portfolioRepo: portfolioRepo,
		auditLogger:   auditLogger,
	}
}

// Execute creates a new section content
func (uc *CreateSectionContentUseCase) Execute(ctx context.Context, input dto.CreateSectionContentInput) (*dto.SectionContentDTO, error) {
	// Validate input
	if input.SectionID == 0 {
		return nil, fmt.Errorf("section ID is required")
	}
	if input.Type == "" {
		return nil, fmt.Errorf("content type is required")
	}
	if input.OwnerID == "" {
		return nil, fmt.Errorf("owner ID is required")
	}

	// Verify section exists and user owns it (through portfolio)
	section, err := uc.sectionRepo.GetByID(ctx, input.SectionID)
	if err != nil {
		return nil, fmt.Errorf("section not found")
	}

	// Verify ownership through portfolio
	portfolio, err := uc.portfolioRepo.GetByID(ctx, section.PortfolioID)
	if err != nil {
		return nil, fmt.Errorf("portfolio not found")
	}
	if portfolio.OwnerID != input.OwnerID {
		return nil, fmt.Errorf("unauthorized: you don't own this section")
	}

	// Create the section content
	content, err := uc.contentRepo.Create(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create section content: %w", err)
	}

	// Audit logging
	if uc.auditLogger != nil {
		uc.auditLogger.LogCreate(ctx, "section_content", content.ID, map[string]interface{}{
			"section_id": content.SectionID,
			"type":       content.Type,
			"owner_id":   content.OwnerID,
		})
	}

	return content, nil
}
