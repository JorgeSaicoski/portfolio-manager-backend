package section_content

import (
	"context"
	"fmt"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/contracts"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/dto"
)

// GetSectionContentPublicUseCase handles the business logic for getting a section content publicly
type GetSectionContentPublicUseCase struct {
	contentRepo contracts.SectionContentRepository
}

// NewGetSectionContentPublicUseCase creates a new instance of GetSectionContentPublicUseCase
func NewGetSectionContentPublicUseCase(
	contentRepo contracts.SectionContentRepository,
) *GetSectionContentPublicUseCase {
	return &GetSectionContentPublicUseCase{
		contentRepo: contentRepo,
	}
}

// Execute retrieves a section content by ID without ownership verification
func (uc *GetSectionContentPublicUseCase) Execute(ctx context.Context, id uint) (*dto.SectionContentDTO, error) {
	if id == 0 {
		return nil, fmt.Errorf("section content ID is required")
	}

	content, err := uc.contentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("section content not found")
	}

	return content, nil
}
