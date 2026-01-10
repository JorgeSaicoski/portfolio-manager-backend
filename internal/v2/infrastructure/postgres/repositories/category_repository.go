package repositories

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/contracts"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/dto"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/infrastructure/postgres/entities"
)

// categoryRepository is the GORM implementation of CategoryRepository
// It implements the contract defined in the application layer
type categoryRepository struct {
	db *gorm.DB
}

// NewCategoryRepository creates a new category repository instance
// Returns the interface type (contracts.CategoryRepository), not the concrete type
func NewCategoryRepository(db *gorm.DB) contracts.CategoryRepository {
	return &categoryRepository{db: db}
}

// Create creates a new category in the database
func (r *categoryRepository) Create(ctx context.Context, input dto.CreateCategoryInput) (*dto.CategoryDTO, error) {
	// Convert application DTO to infrastructure entity
	record := &entities.CategoryRecord{
		Title:       input.Title,
		Description: input.Description,
		Position:    input.Position,
		OwnerID:     input.OwnerID,
		PortfolioID: input.PortfolioID,
	}

	// Persist to database
	if err := r.db.WithContext(ctx).Create(record).Error; err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}

	// Convert infrastructure entity back to application DTO
	return r.recordToDTO(record), nil
}

// GetByID retrieves a category by its ID
func (r *categoryRepository) GetByID(ctx context.Context, id uint) (*dto.CategoryDTO, error) {
	var record entities.CategoryRecord

	if err := r.db.WithContext(ctx).First(&record, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("category with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	return r.recordToDTO(&record), nil
}

// GetByIDs retrieves multiple categories by their IDs
func (r *categoryRepository) GetByIDs(ctx context.Context, ids []uint) ([]dto.CategoryDTO, error) {
	var records []entities.CategoryRecord

	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to get categories by IDs: %w", err)
	}

	// Convert records to DTOs
	dtos := make([]dto.CategoryDTO, len(records))
	for i, record := range records {
		dtos[i] = *r.recordToDTO(&record)
	}

	return dtos, nil
}

// GetByPortfolioID retrieves all categories for a specific portfolio (ordered by position)
func (r *categoryRepository) GetByPortfolioID(ctx context.Context, portfolioID uint) ([]dto.CategoryDTO, error) {
	var records []entities.CategoryRecord

	if err := r.db.WithContext(ctx).
		Where("portfolio_id = ?", portfolioID).
		Order("position ASC, created_at ASC").
		Find(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to get categories by portfolio ID: %w", err)
	}

	// Convert records to DTOs
	dtos := make([]dto.CategoryDTO, len(records))
	for i, record := range records {
		dtos[i] = *r.recordToDTO(&record)
	}

	return dtos, nil
}

// GetByOwnerID retrieves all categories owned by a user with pagination
func (r *categoryRepository) GetByOwnerID(ctx context.Context, ownerID string, pagination dto.PaginationDTO) ([]dto.CategoryDTO, int64, error) {
	var records []entities.CategoryRecord
	var total int64

	// Count total categories for this owner
	if err := r.db.WithContext(ctx).
		Model(&entities.CategoryRecord{}).
		Where("owner_id = ?", ownerID).
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count categories: %w", err)
	}

	// Calculate offset
	offset := (pagination.Page - 1) * pagination.Limit

	// Get paginated results
	if err := r.db.WithContext(ctx).
		Where("owner_id = ?", ownerID).
		Order("created_at DESC").
		Limit(pagination.Limit).
		Offset(offset).
		Find(&records).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list categories: %w", err)
	}

	// Convert records to DTOs
	dtos := make([]dto.CategoryDTO, len(records))
	for i, record := range records {
		dtos[i] = *r.recordToDTO(&record)
	}

	return dtos, total, nil
}

// Update updates an existing category
func (r *categoryRepository) Update(ctx context.Context, input dto.UpdateCategoryInput) error {
	updates := map[string]interface{}{}

	// Only update non-empty fields
	if input.Title != "" {
		updates["title"] = input.Title
	}
	if input.Description != nil {
		updates["description"] = input.Description
	}
	// Position is always updated (even if 0)
	updates["position"] = input.Position

	if len(updates) == 0 {
		return nil // Nothing to update
	}

	result := r.db.WithContext(ctx).
		Model(&entities.CategoryRecord{}).
		Where("id = ?", input.ID).
		Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("failed to update category: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("category with ID %d not found", input.ID)
	}

	return nil
}

// UpdatePosition updates only the position field of a category
func (r *categoryRepository) UpdatePosition(ctx context.Context, id uint, position uint) error {
	result := r.db.WithContext(ctx).
		Model(&entities.CategoryRecord{}).
		Where("id = ?", id).
		Update("position", position)

	if result.Error != nil {
		return fmt.Errorf("failed to update category position: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("category with ID %d not found", id)
	}

	return nil
}

// BulkUpdatePositions updates positions for multiple categories in a transaction
func (r *categoryRepository) BulkUpdatePositions(ctx context.Context, input dto.BulkUpdateCategoryPositionsInput) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, item := range input.Items {
			if err := tx.Model(&entities.CategoryRecord{}).
				Where("id = ?", item.ID).
				Update("position", item.Position).Error; err != nil {
				return fmt.Errorf("failed to update position for category %d: %w", item.ID, err)
			}
		}
		return nil
	})
}

// Delete deletes a category by its ID (soft delete with GORM, cascade handled by GORM)
func (r *categoryRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&entities.CategoryRecord{}, id)

	if result.Error != nil {
		return fmt.Errorf("failed to delete category: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("category with ID %d not found", id)
	}

	return nil
}

// recordToDTO converts a CategoryRecord (infrastructure) to CategoryDTO (application)
func (r *categoryRepository) recordToDTO(record *entities.CategoryRecord) *dto.CategoryDTO {
	return &dto.CategoryDTO{
		ID:          record.ID,
		Title:       record.Title,
		Description: record.Description,
		Position:    record.Position,
		OwnerID:     record.OwnerID,
		PortfolioID: record.PortfolioID,
		CreatedAt:   record.CreatedAt,
		UpdatedAt:   record.UpdatedAt,
	}
}
