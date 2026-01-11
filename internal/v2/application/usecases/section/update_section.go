package section

import (
	"context"
	"fmt"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/contracts"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/dto"
)

// UpdateSectionUseCase handles the business logic for updating a section
type UpdateSectionUseCase struct {
	sectionRepo   contracts.SectionRepository
	portfolioRepo contracts.PortfolioRepository
	auditLogger   contracts.AuditLogger
	metrics       contracts.MetricsCollector
}

// NewUpdateSectionUseCase creates a new instance of UpdateSectionUseCase
func NewUpdateSectionUseCase(
	sectionRepo contracts.SectionRepository,
	portfolioRepo contracts.PortfolioRepository,
	auditLogger contracts.AuditLogger,
	metrics contracts.MetricsCollector,
) *UpdateSectionUseCase {
	return &UpdateSectionUseCase{
		sectionRepo:   sectionRepo,
		portfolioRepo: portfolioRepo,
		auditLogger:   auditLogger,
		metrics:       metrics,
	}
}

// Execute updates a section with ownership verification
func (uc *UpdateSectionUseCase) Execute(ctx context.Context, input dto.UpdateSectionInput) error {
	// Validate input
	if input.ID == 0 {
		return fmt.Errorf("invalid section ID")
	}
	if input.OwnerID == "" {
		return fmt.Errorf("owner ID is required")
	}
	if input.Title == "" {
		return fmt.Errorf("section title is required")
	}
	if input.Type == "" {
		return fmt.Errorf("section type is required")
	}

	// Verify section exists and user owns it
	section, err := uc.sectionRepo.GetByID(ctx, input.ID)
	if err != nil {
		return fmt.Errorf("section not found")
	}

	// Verify ownership through portfolio
	portfolio, err := uc.portfolioRepo.GetByID(ctx, section.PortfolioID)
	if err != nil {
		return fmt.Errorf("portfolio not found")
	}
	if portfolio.OwnerID != input.OwnerID {
		return fmt.Errorf("unauthorized: you don't own this section")
	}

	// Check for duplicate title if title is being changed
	if input.Title != section.Title {
		isDuplicate, err := uc.sectionRepo.CheckTitleDuplicate(ctx, input.Title, section.PortfolioID, input.ID)
		if err != nil {
			return fmt.Errorf("failed to check for duplicate title: %w", err)
		}
		if isDuplicate {
			return fmt.Errorf("section with title '%s' already exists in this portfolio", input.Title)
		}
	}

	// Update the section
	if err := uc.sectionRepo.Update(ctx, input); err != nil {
		return fmt.Errorf("failed to update section: %w", err)
	}

	// Audit logging
	if uc.auditLogger != nil {
		uc.auditLogger.LogUpdate(ctx, "section", input.ID, map[string]interface{}{
			"title":        input.Title,
			"description":  input.Description,
			"position":     input.Position,
			"type":         input.Type,
			"portfolio_id": section.PortfolioID,
			"owner_id":     input.OwnerID,
		})
	}

	// Metrics
	if uc.metrics != nil {
		uc.metrics.IncrementSectionsUpdated()
	}

	return nil
}
