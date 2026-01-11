package project

import (
	"context"
	"fmt"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/contracts"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/dto"
)

// CreateProjectUseCase handles the business logic for creating a project
type CreateProjectUseCase struct {
	projectRepo  contracts.ProjectRepository
	categoryRepo contracts.CategoryRepository
	auditLogger  contracts.AuditLogger
	metrics      contracts.MetricsCollector
}

// NewCreateProjectUseCase creates a new instance of CreateProjectUseCase
func NewCreateProjectUseCase(
	projectRepo contracts.ProjectRepository,
	categoryRepo contracts.CategoryRepository,
	auditLogger contracts.AuditLogger,
	metrics contracts.MetricsCollector,
) *CreateProjectUseCase {
	return &CreateProjectUseCase{
		projectRepo:  projectRepo,
		categoryRepo: categoryRepo,
		auditLogger:  auditLogger,
		metrics:      metrics,
	}
}

// Execute creates a new project
func (uc *CreateProjectUseCase) Execute(ctx context.Context, input dto.CreateProjectInput) (*dto.ProjectDTO, error) {
	// Validate input
	if input.Title == "" {
		return nil, fmt.Errorf("project title is required")
	}
	if input.Description == "" {
		return nil, fmt.Errorf("project description is required")
	}
	if input.OwnerID == "" {
		return nil, fmt.Errorf("owner ID is required")
	}
	if input.CategoryID == 0 {
		return nil, fmt.Errorf("category ID is required")
	}

	// Verify category exists and user owns it
	category, err := uc.categoryRepo.GetByID(ctx, input.CategoryID)
	if err != nil {
		return nil, fmt.Errorf("category not found")
	}
	if category.OwnerID != input.OwnerID {
		return nil, fmt.Errorf("unauthorized: you don't own this category")
	}

	// Create the project
	project, err := uc.projectRepo.Create(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	// Audit logging
	if uc.auditLogger != nil {
		uc.auditLogger.LogCreate(ctx, "project", project.ID, map[string]interface{}{
			"title":       project.Title,
			"category_id": project.CategoryID,
			"owner_id":    project.OwnerID,
		})
	}

	// Metrics
	if uc.metrics != nil {
		// Note: Projects don't have dedicated metrics in MetricsCollector yet
		// This can be added if needed
	}

	return project, nil
}
