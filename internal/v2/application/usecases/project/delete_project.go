package project

import (
	"context"
	"fmt"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/contracts"
)

// DeleteProjectUseCase handles the business logic for deleting a project
type DeleteProjectUseCase struct {
	projectRepo  contracts.ProjectRepository
	categoryRepo contracts.CategoryRepository
	auditLogger  contracts.AuditLogger
}

// NewDeleteProjectUseCase creates a new instance of DeleteProjectUseCase
func NewDeleteProjectUseCase(
	projectRepo contracts.ProjectRepository,
	categoryRepo contracts.CategoryRepository,
	auditLogger contracts.AuditLogger,
) *DeleteProjectUseCase {
	return &DeleteProjectUseCase{
		projectRepo:  projectRepo,
		categoryRepo: categoryRepo,
		auditLogger:  auditLogger,
	}
}

// Execute deletes a project with ownership verification
func (uc *DeleteProjectUseCase) Execute(ctx context.Context, id uint, ownerID string) error {
	if id == 0 {
		return fmt.Errorf("invalid project ID")
	}
	if ownerID == "" {
		return fmt.Errorf("owner ID is required")
	}

	// Verify project exists and user owns it
	project, err := uc.projectRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("project not found")
	}

	// Verify ownership through category
	category, err := uc.categoryRepo.GetByID(ctx, project.CategoryID)
	if err != nil {
		return fmt.Errorf("category not found")
	}
	if category.OwnerID != ownerID {
		return fmt.Errorf("unauthorized: you don't own this project")
	}

	// Delete the project
	if err := uc.projectRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	// Audit logging
	if uc.auditLogger != nil {
		uc.auditLogger.LogDelete(ctx, "project", id, map[string]interface{}{
			"title":       project.Title,
			"category_id": project.CategoryID,
			"owner_id":    ownerID,
		})
	}

	return nil
}
