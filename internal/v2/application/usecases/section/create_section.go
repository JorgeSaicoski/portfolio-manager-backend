package section

import (
	"context"
	"fmt"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/contracts"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/dto"
)

// CreateSectionUseCase handles the business logic for creating a section
type CreateSectionUseCase struct {
	sectionRepo   contracts.SectionRepository
	portfolioRepo contracts.PortfolioRepository
	auditLogger   contracts.AuditLogger
	metrics       contracts.MetricsCollector
}

// NewCreateSectionUseCase creates a new instance of CreateSectionUseCase
func NewCreateSectionUseCase(
	sectionRepo contracts.SectionRepository,
	portfolioRepo contracts.PortfolioRepository,
	auditLogger contracts.AuditLogger,
	metrics contracts.MetricsCollector,
) *CreateSectionUseCase {
	return &CreateSectionUseCase{
		sectionRepo:   sectionRepo,
		portfolioRepo: portfolioRepo,
		auditLogger:   auditLogger,
		metrics:       metrics,
	}
}

// Execute creates a new section
func (uc *CreateSectionUseCase) Execute(ctx context.Context, input dto.CreateSectionInput) (*dto.SectionDTO, error) {
	// Validate input
	if input.Title == "" {
		return nil, fmt.Errorf("section title is required")
	}
	if input.OwnerID == "" {
		return nil, fmt.Errorf("owner ID is required")
	}
	if input.PortfolioID == 0 {
		return nil, fmt.Errorf("portfolio ID is required")
	}
	if input.Type == "" {
		return nil, fmt.Errorf("section type is required")
	}

	// Verify portfolio exists and user owns it
	portfolio, err := uc.portfolioRepo.GetByID(ctx, input.PortfolioID)
	if err != nil {
		return nil, fmt.Errorf("portfolio not found")
	}
	if portfolio.OwnerID != input.OwnerID {
		return nil, fmt.Errorf("unauthorized: you don't own this portfolio")
	}

	// Check for duplicate title in the same portfolio
	isDuplicate, err := uc.sectionRepo.CheckTitleDuplicate(ctx, input.Title, input.PortfolioID, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to check for duplicate title: %w", err)
	}
	if isDuplicate {
		return nil, fmt.Errorf("section with title '%s' already exists in this portfolio", input.Title)
	}

	// Create the section
	section, err := uc.sectionRepo.Create(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create section: %w", err)
	}

	// Audit logging
	if uc.auditLogger != nil {
		uc.auditLogger.LogCreate(ctx, "section", section.ID, map[string]interface{}{
			"title":        section.Title,
			"type":         section.Type,
			"portfolio_id": section.PortfolioID,
			"owner_id":     section.OwnerID,
		})
	}

	// Metrics
	if uc.metrics != nil {
		uc.metrics.IncrementSectionsCreated()
	}

	return section, nil
}
