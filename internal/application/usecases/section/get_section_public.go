package section

import (
	"context"
	"fmt"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/contracts"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/dto"
)

// GetSectionPublicUseCase handles the business logic for retrieving a section publicly (no auth)
type GetSectionPublicUseCase struct {
	sectionRepo contracts.SectionRepository
}

// NewGetSectionPublicUseCase creates a new instance of GetSectionPublicUseCase
func NewGetSectionPublicUseCase(
	sectionRepo contracts.SectionRepository,
) *GetSectionPublicUseCase {
	return &GetSectionPublicUseCase{
		sectionRepo: sectionRepo,
	}
}

// Execute retrieves a section by ID without ownership verification (public access)
func (uc *GetSectionPublicUseCase) Execute(ctx context.Context, id uint) (*dto.SectionDTO, error) {
	if id == 0 {
		return nil, fmt.Errorf("invalid section ID")
	}

	// Get section (no ownership check for public access)
	section, err := uc.sectionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("section not found")
	}

	return section, nil
}
