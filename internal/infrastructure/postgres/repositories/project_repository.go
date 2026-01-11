package repositories

import (
	"context"
	"fmt"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/contracts"
	dto2 "github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/dto"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/infrastructure/postgres/entities"
	"gorm.io/gorm"
)

// projectRepository implements the ProjectRepository interface using GORM
type projectRepository struct {
	db *gorm.DB
}

// NewProjectRepository creates a new project repository instance
func NewProjectRepository(db *gorm.DB) contracts.ProjectRepository {
	return &projectRepository{db: db}
}

// Create creates a new project
func (r *projectRepository) Create(ctx context.Context, input dto2.CreateProjectInput) (*dto2.ProjectDTO, error) {
	record := &entities.ProjectRecord{
		Title:       input.Title,
		Description: input.Description,
		MainImage:   input.MainImage,
		Images:      input.Images,
		Skills:      input.Skills,
		Client:      input.Client,
		Link:        input.Link,
		CategoryID:  input.CategoryID,
		OwnerID:     input.OwnerID,
	}

	if err := r.db.WithContext(ctx).Create(record).Error; err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	return r.recordToDTO(record), nil
}

// GetByID retrieves a project by ID
func (r *projectRepository) GetByID(ctx context.Context, id uint) (*dto2.ProjectDTO, error) {
	var record entities.ProjectRecord
	if err := r.db.WithContext(ctx).First(&record, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("project not found")
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	return r.recordToDTO(&record), nil
}

// GetByIDs retrieves multiple projects by their IDs
func (r *projectRepository) GetByIDs(ctx context.Context, ids []uint) ([]dto2.ProjectDTO, error) {
	var records []entities.ProjectRecord
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to get projects: %w", err)
	}

	dtos := make([]dto2.ProjectDTO, len(records))
	for i, record := range records {
		dtos[i] = *r.recordToDTO(&record)
	}

	return dtos, nil
}

// GetByCategoryID retrieves all projects for a specific category
func (r *projectRepository) GetByCategoryID(ctx context.Context, categoryID uint) ([]dto2.ProjectDTO, error) {
	var records []entities.ProjectRecord
	if err := r.db.WithContext(ctx).
		Where("category_id = ?", categoryID).
		Order("id ASC").
		Find(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to get projects by category: %w", err)
	}

	dtos := make([]dto2.ProjectDTO, len(records))
	for i, record := range records {
		dtos[i] = *r.recordToDTO(&record)
	}

	return dtos, nil
}

// GetByOwnerID retrieves all projects owned by a specific user with pagination
func (r *projectRepository) GetByOwnerID(ctx context.Context, ownerID string, pagination dto2.PaginationDTO) ([]dto2.ProjectDTO, int64, error) {
	var records []entities.ProjectRecord
	var total int64

	// Count total
	if err := r.db.WithContext(ctx).Model(&entities.ProjectRecord{}).
		Where("owner_id = ?", ownerID).
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count projects: %w", err)
	}

	// Get paginated results
	offset := (pagination.Page - 1) * pagination.Limit
	if err := r.db.WithContext(ctx).
		Where("owner_id = ?", ownerID).
		Order("id DESC").
		Limit(pagination.Limit).
		Offset(offset).
		Find(&records).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get projects: %w", err)
	}

	dtos := make([]dto2.ProjectDTO, len(records))
	for i, record := range records {
		dtos[i] = *r.recordToDTO(&record)
	}

	return dtos, total, nil
}

// SearchBySkills retrieves projects matching ANY of the specified skills
func (r *projectRepository) SearchBySkills(ctx context.Context, skills []string) ([]dto2.ProjectDTO, error) {
	var records []entities.ProjectRecord

	// Use PostgreSQL array overlap operator (&&)
	if err := r.db.WithContext(ctx).
		Where("skills && ?", skills).
		Order("id DESC").
		Find(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to search projects by skills: %w", err)
	}

	dtos := make([]dto2.ProjectDTO, len(records))
	for i, record := range records {
		dtos[i] = *r.recordToDTO(&record)
	}

	return dtos, nil
}

// SearchByClient retrieves projects by client name (case-insensitive partial match)
func (r *projectRepository) SearchByClient(ctx context.Context, client string) ([]dto2.ProjectDTO, error) {
	var records []entities.ProjectRecord

	if err := r.db.WithContext(ctx).
		Where("client ILIKE ?", "%"+client+"%").
		Order("id DESC").
		Find(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to search projects by client: %w", err)
	}

	dtos := make([]dto2.ProjectDTO, len(records))
	for i, record := range records {
		dtos[i] = *r.recordToDTO(&record)
	}

	return dtos, nil
}

// Update updates an existing project
func (r *projectRepository) Update(ctx context.Context, input dto2.UpdateProjectInput) error {
	updates := map[string]interface{}{
		"title":       input.Title,
		"description": input.Description,
		"main_image":  input.MainImage,
		"images":      input.Images,
		"skills":      input.Skills,
		"client":      input.Client,
		"link":        input.Link,
	}

	if err := r.db.WithContext(ctx).
		Model(&entities.ProjectRecord{}).
		Where("id = ?", input.ID).
		Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}

	return nil
}

// Delete deletes a project by ID
func (r *projectRepository) Delete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&entities.ProjectRecord{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	return nil
}

// recordToDTO converts a ProjectRecord to ProjectDTO
func (r *projectRepository) recordToDTO(record *entities.ProjectRecord) *dto2.ProjectDTO {
	return &dto2.ProjectDTO{
		ID:          record.ID,
		Title:       record.Title,
		Description: record.Description,
		MainImage:   record.MainImage,
		Images:      record.Images,
		Skills:      record.Skills,
		Client:      record.Client,
		Link:        record.Link,
		CategoryID:  record.CategoryID,
		OwnerID:     record.OwnerID,
		CreatedAt:   record.CreatedAt,
		UpdatedAt:   record.UpdatedAt,
	}
}
