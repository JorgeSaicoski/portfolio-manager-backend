package section_content

import (
	"context"
	"fmt"

	contracts2 "github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/contracts"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/dto"
)

// CreateSectionContentUseCase handles the business logic for creating a section content
type CreateSectionContentUseCase struct {
	contentRepo   contracts2.SectionContentRepository
	sectionRepo   contracts2.SectionRepository
	portfolioRepo contracts2.PortfolioRepository
	auditLogger   contracts2.AuditLogger
}

// NewCreateSectionContentUseCase creates a new instance of CreateSectionContentUseCase
func NewCreateSectionContentUseCase(
	contentRepo contracts2.SectionContentRepository,
	sectionRepo contracts2.SectionRepository,
	portfolioRepo contracts2.PortfolioRepository,
	auditLogger contracts2.AuditLogger,
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
