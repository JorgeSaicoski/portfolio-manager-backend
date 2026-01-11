package project

import (
	"context"
	"fmt"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/contracts"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/dto"
)

// GetProjectUseCase handles the business logic for retrieving a project by ID
type GetProjectUseCase struct {
	projectRepo  contracts.ProjectRepository
	categoryRepo contracts.CategoryRepository
	auditLogger  contracts.AuditLogger
}

// NewGetProjectUseCase creates a new instance of GetProjectUseCase
func NewGetProjectUseCase(
	projectRepo contracts.ProjectRepository,
	categoryRepo contracts.CategoryRepository,
	auditLogger contracts.AuditLogger,
) *GetProjectUseCase {
	return &GetProjectUseCase{
		projectRepo:  projectRepo,
		categoryRepo: categoryRepo,
		auditLogger:  auditLogger,
	}
}

// Execute retrieves a project by ID with ownership verification
func (uc *GetProjectUseCase) Execute(ctx context.Context, id uint, ownerID string) (*dto.ProjectDTO, error) {
	if id == 0 {
		return nil, fmt.Errorf("invalid project ID")
	}
	if ownerID == "" {
		return nil, fmt.Errorf("owner ID is required")
	}

	// Get project
	project, err := uc.projectRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("project not found")
	}

	// Verify ownership through category
	category, err := uc.categoryRepo.GetByID(ctx, project.CategoryID)
	if err != nil {
		return nil, fmt.Errorf("category not found")
	}
	if category.OwnerID != ownerID {
		// Log unauthorized access attempt
		if uc.auditLogger != nil {
			uc.auditLogger.LogAccess(ctx, "project", id, ownerID, false)
		}
		return nil, fmt.Errorf("unauthorized: you don't own this project")
	}

	// Log authorized access
	if uc.auditLogger != nil {
		uc.auditLogger.LogAccess(ctx, "project", id, ownerID, true)
	}

	return project, nil
}
