package project

import (
	"context"
	"fmt"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/contracts"
	dto2 "github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/dto"
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
func (uc *ListProjectsUseCase) Execute(ctx context.Context, input dto2.ListProjectsInput) (*dto2.ListProjectsOutput, error) {
	// Get projects with pagination
	projects, total, err := uc.projectRepo.GetByOwnerID(ctx, input.OwnerID, input.Pagination)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	return &dto2.ListProjectsOutput{
		Projects: projects,
		Pagination: dto2.PaginatedResultDTO{
			Total: total,
			Page:  input.Pagination.Page,
			Limit: input.Pagination.Limit,
		},
	}, nil
}
