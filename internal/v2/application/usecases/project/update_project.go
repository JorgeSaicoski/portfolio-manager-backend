package project

import (
	"context"
	"fmt"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/contracts"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/dto"
)

// UpdateProjectUseCase handles the business logic for updating a project
type UpdateProjectUseCase struct {
	projectRepo  contracts.ProjectRepository
	categoryRepo contracts.CategoryRepository
	auditLogger  contracts.AuditLogger
}

// NewUpdateProjectUseCase creates a new instance of UpdateProjectUseCase
func NewUpdateProjectUseCase(
	projectRepo contracts.ProjectRepository,
	categoryRepo contracts.CategoryRepository,
	auditLogger contracts.AuditLogger,
) *UpdateProjectUseCase {
	return &UpdateProjectUseCase{
		projectRepo:  projectRepo,
		categoryRepo: categoryRepo,
		auditLogger:  auditLogger,
	}
}

// Execute updates a project with ownership verification
func (uc *UpdateProjectUseCase) Execute(ctx context.Context, input dto.UpdateProjectInput) error {
	// Validate input
	if input.ID == 0 {
		return fmt.Errorf("invalid project ID")
	}
	if input.OwnerID == "" {
		return fmt.Errorf("owner ID is required")
	}
	if input.Title == "" {
		return fmt.Errorf("project title is required")
	}
	if input.Description == "" {
		return fmt.Errorf("project description is required")
	}

	// Verify project exists and user owns it
	project, err := uc.projectRepo.GetByID(ctx, input.ID)
	if err != nil {
		return fmt.Errorf("project not found")
	}

	// Verify ownership through category
	category, err := uc.categoryRepo.GetByID(ctx, project.CategoryID)
	if err != nil {
		return fmt.Errorf("category not found")
	}
	if category.OwnerID != input.OwnerID {
		return fmt.Errorf("unauthorized: you don't own this project")
	}

	// Update the project
	if err := uc.projectRepo.Update(ctx, input); err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}

	// Audit logging
	if uc.auditLogger != nil {
		uc.auditLogger.LogUpdate(ctx, "project", input.ID, map[string]interface{}{
			"title":       input.Title,
			"description": input.Description,
			"category_id": project.CategoryID,
			"owner_id":    input.OwnerID,
		})
	}

	return nil
}
