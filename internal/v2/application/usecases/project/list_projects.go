package project

import (
	"context"
	"fmt"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/contracts"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/dto"
)

// ListProjectsUseCase handles the business logic for listing projects
type ListProjectsUseCase struct {
	projectRepo contracts.ProjectRepository
}

// NewListProjectsUseCase creates a new instance of ListProjectsUseCase
func NewListProjectsUseCase(
	projectRepo contracts.ProjectRepository,
) *ListProjectsUseCase {
	return &ListProjectsUseCase{
		projectRepo: projectRepo,
	}
}

// Execute retrieves all projects owned by a user with pagination
func (uc *ListProjectsUseCase) Execute(ctx context.Context, input dto.ListProjectsInput) (*dto.ListProjectsOutput, error) {
	// Get projects with pagination
	projects, total, err := uc.projectRepo.GetByOwnerID(ctx, input.OwnerID, input.Pagination)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	return &dto.ListProjectsOutput{
		Projects: projects,
		Pagination: dto.PaginatedResultDTO{
			Total: total,
			Page:  input.Pagination.Page,
			Limit: input.Pagination.Limit,
		},
	}, nil
}
