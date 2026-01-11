package contracts

import (
	"context"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/dto"
)

// ProjectRepository defines the interface for project data persistence
type ProjectRepository interface {
	// Create creates a new project and returns the created project DTO
	Create(ctx context.Context, input dto.CreateProjectInput) (*dto.ProjectDTO, error)

	// GetByID retrieves a project by its ID
	GetByID(ctx context.Context, id uint) (*dto.ProjectDTO, error)

	// GetByIDs retrieves multiple projects by their IDs
	GetByIDs(ctx context.Context, ids []uint) ([]dto.ProjectDTO, error)

	// GetByCategoryID retrieves all projects for a specific category (ordered by ID)
	GetByCategoryID(ctx context.Context, categoryID uint) ([]dto.ProjectDTO, error)

	// GetByOwnerID retrieves all projects owned by a specific user with pagination
	GetByOwnerID(ctx context.Context, ownerID string, pagination dto.PaginationDTO) ([]dto.ProjectDTO, int64, error)

	// SearchBySkills retrieves projects matching ANY of the specified skills
	SearchBySkills(ctx context.Context, skills []string) ([]dto.ProjectDTO, error)

	// SearchByClient retrieves projects by client name (case-insensitive partial match)
	SearchByClient(ctx context.Context, client string) ([]dto.ProjectDTO, error)

	// Update updates an existing project
	Update(ctx context.Context, input dto.UpdateProjectInput) error

	// Delete deletes a project by its ID
	Delete(ctx context.Context, id uint) error
}
