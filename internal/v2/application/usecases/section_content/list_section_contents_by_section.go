package section_content

import (
	"context"
	"fmt"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/contracts"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/dto"
)

// ListSectionContentsBySectionUseCase handles the business logic for listing section contents by section ID
type ListSectionContentsBySectionUseCase struct {
	contentRepo contracts.SectionContentRepository
}

// NewListSectionContentsBySectionUseCase creates a new instance of ListSectionContentsBySectionUseCase
func NewListSectionContentsBySectionUseCase(
	contentRepo contracts.SectionContentRepository,
) *ListSectionContentsBySectionUseCase {
	return &ListSectionContentsBySectionUseCase{
		contentRepo: contentRepo,
	}
}

// Execute retrieves all section contents for a specific section
func (uc *ListSectionContentsBySectionUseCase) Execute(ctx context.Context, sectionID uint) ([]dto.SectionContentDTO, error) {
	if sectionID == 0 {
		return nil, fmt.Errorf("section ID is required")
	}

	contents, err := uc.contentRepo.GetBySectionID(ctx, sectionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get section contents: %w", err)
	}

	return contents, nil
}
