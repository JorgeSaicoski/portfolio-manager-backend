package project

import (
	"context"
	"fmt"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/contracts"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/dto"
)

// GetProjectPublicUseCase handles the business logic for retrieving a project publicly (no auth)
type GetProjectPublicUseCase struct {
	projectRepo contracts.ProjectRepository
}

// NewGetProjectPublicUseCase creates a new instance of GetProjectPublicUseCase
func NewGetProjectPublicUseCase(
	projectRepo contracts.ProjectRepository,
) *GetProjectPublicUseCase {
	return &GetProjectPublicUseCase{
		projectRepo: projectRepo,
	}
}

// Execute retrieves a project by ID without ownership verification (public access)
func (uc *GetProjectPublicUseCase) Execute(ctx context.Context, id uint) (*dto.ProjectDTO, error) {
	if id == 0 {
		return nil, fmt.Errorf("invalid project ID")
	}

	// Get project (no ownership check for public access)
	project, err := uc.projectRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("project not found")
	}

	return project, nil
}
