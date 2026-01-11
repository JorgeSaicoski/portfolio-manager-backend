package section

import (
	"context"
	"fmt"

	contracts2 "github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/contracts"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/dto"
)

// GetSectionUseCase handles the business logic for retrieving a section by ID
type GetSectionUseCase struct {
	sectionRepo   contracts2.SectionRepository
	portfolioRepo contracts2.PortfolioRepository
	auditLogger   contracts2.AuditLogger
}

// NewGetSectionUseCase creates a new instance of GetSectionUseCase
func NewGetSectionUseCase(
	sectionRepo contracts2.SectionRepository,
	portfolioRepo contracts2.PortfolioRepository,
	auditLogger contracts2.AuditLogger,
) *GetSectionUseCase {
	return &GetSectionUseCase{
		sectionRepo:   sectionRepo,
		portfolioRepo: portfolioRepo,
		auditLogger:   auditLogger,
	}
}

// Execute retrieves a section by ID with ownership verification
func (uc *GetSectionUseCase) Execute(ctx context.Context, id uint, ownerID string) (*dto.SectionDTO, error) {
	if id == 0 {
		return nil, fmt.Errorf("invalid section ID")
	}
	if ownerID == "" {
		return nil, fmt.Errorf("owner ID is required")
	}

	// Get section
	section, err := uc.sectionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("section not found")
	}

	// Verify ownership through portfolio
	portfolio, err := uc.portfolioRepo.GetByID(ctx, section.PortfolioID)
	if err != nil {
		return nil, fmt.Errorf("portfolio not found")
	}
	if portfolio.OwnerID != ownerID {
		// Log unauthorized access attempt
		if uc.auditLogger != nil {
			uc.auditLogger.LogAccess(ctx, "section", id, ownerID, false)
		}
		return nil, fmt.Errorf("unauthorized: you don't own this section")
	}

	// Log authorized access
	if uc.auditLogger != nil {
		uc.auditLogger.LogAccess(ctx, "section", id, ownerID, true)
	}

	return section, nil
}
