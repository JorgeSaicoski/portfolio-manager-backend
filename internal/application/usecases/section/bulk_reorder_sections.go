package section

import (
	"context"
	"fmt"

	contracts2 "github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/contracts"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/dto"
)

// BulkReorderSectionsUseCase handles the business logic for bulk reordering sections
type BulkReorderSectionsUseCase struct {
	sectionRepo   contracts2.SectionRepository
	portfolioRepo contracts2.PortfolioRepository
	auditLogger   contracts2.AuditLogger
}

// NewBulkReorderSectionsUseCase creates a new instance of BulkReorderSectionsUseCase
func NewBulkReorderSectionsUseCase(
	sectionRepo contracts2.SectionRepository,
	portfolioRepo contracts2.PortfolioRepository,
	auditLogger contracts2.AuditLogger,
) *BulkReorderSectionsUseCase {
	return &BulkReorderSectionsUseCase{
		sectionRepo:   sectionRepo,
		portfolioRepo: portfolioRepo,
		auditLogger:   auditLogger,
	}
}

// Execute performs bulk position updates with ownership verification
func (uc *BulkReorderSectionsUseCase) Execute(ctx context.Context, input dto.BulkUpdateSectionPositionsInput) error {
	if len(input.Items) == 0 {
		return fmt.Errorf("no sections to reorder")
	}
	if input.OwnerID == "" {
		return fmt.Errorf("owner ID is required")
	}

	// Verify ownership of all sections
	sectionIDs := make([]uint, len(input.Items))
	for i, item := range input.Items {
		sectionIDs[i] = item.ID
	}

	sections, err := uc.sectionRepo.GetByIDs(ctx, sectionIDs)
	if err != nil {
		return fmt.Errorf("failed to retrieve sections: %w", err)
	}
	if len(sections) != len(sectionIDs) {
		return fmt.Errorf("some sections not found")
	}

	// Verify all sections belong to portfolios owned by the user
	portfolioIDSet := make(map[uint]bool)
	for _, sec := range sections {
		portfolioIDSet[sec.PortfolioID] = true
	}

	for portfolioID := range portfolioIDSet {
		portfolio, err := uc.portfolioRepo.GetByID(ctx, portfolioID)
		if err != nil {
			return fmt.Errorf("portfolio not found")
		}
		if portfolio.OwnerID != input.OwnerID {
			return fmt.Errorf("unauthorized: you don't own all sections")
		}
	}

	// Perform bulk update
	if err := uc.sectionRepo.BulkUpdatePositions(ctx, input); err != nil {
		return fmt.Errorf("failed to reorder sections: %w", err)
	}

	// Audit logging
	if uc.auditLogger != nil {
		uc.auditLogger.LogUpdate(ctx, "section", 0, map[string]interface{}{
			"operation": "bulk_reorder",
			"count":     len(input.Items),
			"owner_id":  input.OwnerID,
		})
	}

	return nil
}
