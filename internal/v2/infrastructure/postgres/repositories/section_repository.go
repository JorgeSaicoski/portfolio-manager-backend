package repositories

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/contracts"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/dto"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/infrastructure/postgres/entities"
)

// sectionRepository is the GORM implementation of SectionRepository
// It implements the contract defined in the application layer
type sectionRepository struct {
	db *gorm.DB
}

// NewSectionRepository creates a new section repository instance
// Returns the interface type (contracts.SectionRepository), not the concrete type
func NewSectionRepository(db *gorm.DB) contracts.SectionRepository {
	return &sectionRepository{db: db}
}

// Create creates a new section in the database
func (r *sectionRepository) Create(ctx context.Context, input dto.CreateSectionInput) (*dto.SectionDTO, error) {
	// Convert application DTO to infrastructure entity
	record := &entities.SectionRecord{
		Title:       input.Title,
		Description: input.Description,
		Type:        input.Type,
		Position:    input.Position,
		OwnerID:     input.OwnerID,
		PortfolioID: input.PortfolioID,
	}

	// Persist to database
	if err := r.db.WithContext(ctx).Create(record).Error; err != nil {
		return nil, fmt.Errorf("failed to create section: %w", err)
	}

	// Convert infrastructure entity back to application DTO
	return r.recordToDTO(record), nil
}

// GetByID retrieves a section by its ID
func (r *sectionRepository) GetByID(ctx context.Context, id uint) (*dto.SectionDTO, error) {
	var record entities.SectionRecord

	if err := r.db.WithContext(ctx).First(&record, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("section with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to get section: %w", err)
	}

	return r.recordToDTO(&record), nil
}

// GetByIDs retrieves multiple sections by their IDs
func (r *sectionRepository) GetByIDs(ctx context.Context, ids []uint) ([]dto.SectionDTO, error) {
	var records []entities.SectionRecord

	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to get sections by IDs: %w", err)
	}

	// Convert records to DTOs
	dtos := make([]dto.SectionDTO, len(records))
	for i, record := range records {
		dtos[i] = *r.recordToDTO(&record)
	}

	return dtos, nil
}

// GetByPortfolioID retrieves all sections for a specific portfolio (ordered by position)
func (r *sectionRepository) GetByPortfolioID(ctx context.Context, portfolioID uint) ([]dto.SectionDTO, error) {
	var records []entities.SectionRecord

	if err := r.db.WithContext(ctx).
		Where("portfolio_id = ?", portfolioID).
		Order("position ASC, created_at ASC").
		Find(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to get sections by portfolio ID: %w", err)
	}

	// Convert records to DTOs
	dtos := make([]dto.SectionDTO, len(records))
	for i, record := range records {
		dtos[i] = *r.recordToDTO(&record)
	}

	return dtos, nil
}

// GetByOwnerID retrieves all sections owned by a user with pagination
func (r *sectionRepository) GetByOwnerID(ctx context.Context, ownerID string, pagination dto.PaginationDTO) ([]dto.SectionDTO, int64, error) {
	var records []entities.SectionRecord
	var total int64

	// Count total sections for this owner
	if err := r.db.WithContext(ctx).
		Model(&entities.SectionRecord{}).
		Where("owner_id = ?", ownerID).
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count sections: %w", err)
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
		return nil, 0, fmt.Errorf("failed to list sections: %w", err)
	}

	// Convert records to DTOs
	dtos := make([]dto.SectionDTO, len(records))
	for i, record := range records {
		dtos[i] = *r.recordToDTO(&record)
	}

	return dtos, total, nil
}

// GetByType retrieves all sections of a specific type
func (r *sectionRepository) GetByType(ctx context.Context, sectionType string) ([]dto.SectionDTO, error) {
	var records []entities.SectionRecord

	if err := r.db.WithContext(ctx).
		Where("type = ?", sectionType).
		Order("position ASC, created_at ASC").
		Find(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to get sections by type: %w", err)
	}

	// Convert records to DTOs
	dtos := make([]dto.SectionDTO, len(records))
	for i, record := range records {
		dtos[i] = *r.recordToDTO(&record)
	}

	return dtos, nil
}

// Update updates an existing section
func (r *sectionRepository) Update(ctx context.Context, input dto.UpdateSectionInput) error {
	updates := map[string]interface{}{}

	// Only update non-empty fields
	if input.Title != "" {
		updates["title"] = input.Title
	}
	if input.Description != nil {
		updates["description"] = input.Description
	}
	if input.Type != "" {
		updates["type"] = input.Type
	}
	// Position is always updated (even if 0)
	updates["position"] = input.Position

	if len(updates) == 0 {
		return nil // Nothing to update
	}

	result := r.db.WithContext(ctx).
		Model(&entities.SectionRecord{}).
		Where("id = ?", input.ID).
		Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("failed to update section: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("section with ID %d not found", input.ID)
	}

	return nil
}

// UpdatePosition updates only the position field of a section
func (r *sectionRepository) UpdatePosition(ctx context.Context, id uint, position uint) error {
	result := r.db.WithContext(ctx).
		Model(&entities.SectionRecord{}).
		Where("id = ?", id).
		Update("position", position)

	if result.Error != nil {
		return fmt.Errorf("failed to update section position: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("section with ID %d not found", id)
	}

	return nil
}

// BulkUpdatePositions updates positions for multiple sections in a transaction
func (r *sectionRepository) BulkUpdatePositions(ctx context.Context, input dto.BulkUpdateSectionPositionsInput) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, item := range input.Items {
			if err := tx.Model(&entities.SectionRecord{}).
				Where("id = ?", item.ID).
				Update("position", item.Position).Error; err != nil {
				return fmt.Errorf("failed to update position for section %d: %w", item.ID, err)
			}
		}
		return nil
	})
}

// Delete deletes a section by its ID (soft delete with GORM, cascade handled by GORM)
func (r *sectionRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&entities.SectionRecord{}, id)

	if result.Error != nil {
		return fmt.Errorf("failed to delete section: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("section with ID %d not found", id)
	}

	return nil
}

// CheckTitleDuplicate checks if a section title already exists for a portfolio
func (r *sectionRepository) CheckTitleDuplicate(ctx context.Context, title string, portfolioID uint, excludeID uint) (bool, error) {
	var count int64

	query := r.db.WithContext(ctx).
		Model(&entities.SectionRecord{}).
		Where("title = ? AND portfolio_id = ?", title, portfolioID)

	// Exclude the current section when updating
	if excludeID > 0 {
		query = query.Where("id != ?", excludeID)
	}

	if err := query.Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check duplicate title: %w", err)
	}

	return count > 0, nil
}

// recordToDTO converts a SectionRecord (infrastructure) to SectionDTO (application)
func (r *sectionRepository) recordToDTO(record *entities.SectionRecord) *dto.SectionDTO {
	return &dto.SectionDTO{
		ID:          record.ID,
		Title:       record.Title,
		Description: record.Description,
		Type:        record.Type,
		Position:    record.Position,
		OwnerID:     record.OwnerID,
		PortfolioID: record.PortfolioID,
		CreatedAt:   record.CreatedAt,
		UpdatedAt:   record.UpdatedAt,
	}
}
