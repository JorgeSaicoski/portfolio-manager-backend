package repositories

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/contracts"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/dto"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/infrastructure/postgres/entities"
)

// sectionContentRepository implements the SectionContentRepository interface using GORM
type sectionContentRepository struct {
	db *gorm.DB
}

// NewSectionContentRepository creates a new section content repository instance
func NewSectionContentRepository(db *gorm.DB) contracts.SectionContentRepository {
	return &sectionContentRepository{db: db}
}

// Create creates a new section content
func (r *sectionContentRepository) Create(ctx context.Context, input dto.CreateSectionContentInput) (*dto.SectionContentDTO, error) {
	record := &entities.SectionContentRecord{
		SectionID: input.SectionID,
		Type:      input.Type,
		Content:   input.Content,
		Order:     input.Order,
		ImageID:   input.ImageID,
		OwnerID:   input.OwnerID,
	}

	if err := r.db.WithContext(ctx).Create(record).Error; err != nil {
		return nil, fmt.Errorf("failed to create section content: %w", err)
	}

	return r.recordToDTO(record), nil
}

// GetByID retrieves a section content by ID
func (r *sectionContentRepository) GetByID(ctx context.Context, id uint) (*dto.SectionContentDTO, error) {
	var record entities.SectionContentRecord
	if err := r.db.WithContext(ctx).First(&record, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("section content not found")
		}
		return nil, fmt.Errorf("failed to get section content: %w", err)
	}

	return r.recordToDTO(&record), nil
}

// GetBySectionID retrieves all section contents for a specific section
func (r *sectionContentRepository) GetBySectionID(ctx context.Context, sectionID uint) ([]dto.SectionContentDTO, error) {
	var records []entities.SectionContentRecord
	if err := r.db.WithContext(ctx).
		Where("section_id = ?", sectionID).
		Order("\"order\" ASC, id ASC").
		Find(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to get section contents: %w", err)
	}

	dtos := make([]dto.SectionContentDTO, len(records))
	for i, record := range records {
		dtos[i] = *r.recordToDTO(&record)
	}

	return dtos, nil
}

// Update updates an existing section content
func (r *sectionContentRepository) Update(ctx context.Context, input dto.UpdateSectionContentInput) error {
	updates := map[string]interface{}{
		"type":     input.Type,
		"content":  input.Content,
		"order":    input.Order,
		"image_id": input.ImageID,
	}

	if err := r.db.WithContext(ctx).
		Model(&entities.SectionContentRecord{}).
		Where("id = ?", input.ID).
		Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update section content: %w", err)
	}

	return nil
}

// UpdateOrder updates only the order field of a section content
func (r *sectionContentRepository) UpdateOrder(ctx context.Context, id uint, order uint) error {
	if err := r.db.WithContext(ctx).
		Model(&entities.SectionContentRecord{}).
		Where("id = ?", id).
		Update("order", order).Error; err != nil {
		return fmt.Errorf("failed to update section content order: %w", err)
	}

	return nil
}

// Delete deletes a section content by ID
func (r *sectionContentRepository) Delete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&entities.SectionContentRecord{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete section content: %w", err)
	}

	return nil
}

// recordToDTO converts a SectionContentRecord to SectionContentDTO
func (r *sectionContentRepository) recordToDTO(record *entities.SectionContentRecord) *dto.SectionContentDTO {
	return &dto.SectionContentDTO{
		ID:        record.ID,
		SectionID: record.SectionID,
		Type:      record.Type,
		Content:   record.Content,
		Order:     record.Order,
		ImageID:   record.ImageID,
		OwnerID:   record.OwnerID,
		CreatedAt: record.CreatedAt,
		UpdatedAt: record.UpdatedAt,
	}
}
