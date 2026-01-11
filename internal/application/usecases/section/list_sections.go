package section

import (
	"context"
	"fmt"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/contracts"
	dto2 "github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/dto"
)

// ListSectionsUseCase handles the business logic for listing sections
type ListSectionsUseCase struct {
	sectionRepo contracts.SectionRepository
}

// NewListSectionsUseCase creates a new instance of ListSectionsUseCase
func NewListSectionsUseCase(
	sectionRepo contracts.SectionRepository,
) *ListSectionsUseCase {
	return &ListSectionsUseCase{
		sectionRepo: sectionRepo,
	}
}

// Execute retrieves all sections owned by a user with pagination
func (uc *ListSectionsUseCase) Execute(ctx context.Context, input dto2.ListSectionsInput) (*dto2.ListSectionsOutput, error) {
	// Get sections with pagination
	sections, total, err := uc.sectionRepo.GetByOwnerID(ctx, input.PortfolioID, input.Pagination)
	if err != nil {
		return nil, fmt.Errorf("failed to list sections: %w", err)
	}

	return &dto2.ListSectionsOutput{
		Sections: sections,
		Pagination: dto2.PaginatedResultDTO{
			Total: total,
			Page:  input.Pagination.Page,
			Limit: input.Pagination.Limit,
		},
	}, nil
}
