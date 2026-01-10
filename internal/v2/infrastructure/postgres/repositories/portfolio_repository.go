package repositories

import (
	"context"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/contracts"
	dto2 "github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/dto"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/infrastructure/postgres/entities"
	"gorm.io/gorm"
)

// portfolioRepository implements the PortfolioRepository contract from application layer
type portfolioRepository struct {
	db *gorm.DB
}

// NewPortfolioRepository creates a new portfolio repository instance
// Returns the application contract interface, not the concrete type
func NewPortfolioRepository(db *gorm.DB) contracts.PortfolioRepository {
	return &portfolioRepository{db: db}
}

// Create creates a new portfolio and returns it with the generated ID
func (r *portfolioRepository) Create(ctx context.Context, input dto2.CreatePortfolioInput) (*dto2.PortfolioDTO, error) {
	// Convert application DTO to infrastructure entity
	record := &entities.PortfolioRecord{
		Title:       input.Title,
		Description: input.Description,
		OwnerID:     input.OwnerID,
	}

	// Persist to database
	if err := r.db.WithContext(ctx).Create(record).Error; err != nil {
		return nil, err
	}

	// Convert back to application DTO
	return r.recordToDTO(record), nil
}

// GetByID retrieves a portfolio by its ID
func (r *portfolioRepository) GetByID(ctx context.Context, id uint) (*dto2.PortfolioDTO, error) {
	var record entities.PortfolioRecord

	if err := r.db.WithContext(ctx).First(&record, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}

	return r.recordToDTO(&record), nil
}

// GetByOwnerID retrieves all portfolios owned by a specific user
func (r *portfolioRepository) GetByOwnerID(ctx context.Context, ownerID string, pagination dto2.PaginationDTO) ([]dto2.PortfolioDTO, int64, error) {
	var records []entities.PortfolioRecord
	var total int64

	// Count total records
	if err := r.db.WithContext(ctx).Model(&entities.PortfolioRecord{}).
		Where("owner_id = ?", ownerID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated records
	offset := (pagination.Page - 1) * pagination.Limit
	if err := r.db.WithContext(ctx).
		Where("owner_id = ?", ownerID).
		Limit(pagination.Limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&records).Error; err != nil {
		return nil, 0, err
	}

	// Convert to DTOs
	dtos := make([]dto2.PortfolioDTO, len(records))
	for i, record := range records {
		dtos[i] = *r.recordToDTO(&record)
	}

	return dtos, total, nil
}

// Update updates an existing portfolio
func (r *portfolioRepository) Update(ctx context.Context, input dto2.UpdatePortfolioInput) error {
	updates := make(map[string]interface{})

	if input.Title != "" {
		updates["title"] = input.Title
	}
	if input.Description != "" {
		updates["description"] = input.Description
	}

	if len(updates) == 0 {
		return nil // Nothing to update
	}

	return r.db.WithContext(ctx).
		Model(&entities.PortfolioRecord{}).
		Where("id = ?", input.ID).
		Updates(updates).Error
}

// Delete deletes a portfolio by ID
func (r *portfolioRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entities.PortfolioRecord{}, id).Error
}

// CheckTitleDuplicate checks if a title already exists for an owner
func (r *portfolioRepository) CheckTitleDuplicate(ctx context.Context, title, ownerID string, excludeID uint) (bool, error) {
	var count int64

	query := r.db.WithContext(ctx).Model(&entities.PortfolioRecord{}).
		Where("title = ? AND owner_id = ?", title, ownerID)

	if excludeID != 0 {
		query = query.Where("id != ?", excludeID)
	}

	if err := query.Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

// recordToDTO converts a GORM record to an application DTO
func (r *portfolioRepository) recordToDTO(record *entities.PortfolioRecord) *dto2.PortfolioDTO {
	return &dto2.PortfolioDTO{
		ID:          record.ID,
		Title:       record.Title,
		Description: record.Description,
		OwnerID:     record.OwnerID,
		CreatedAt:   record.CreatedAt,
		UpdatedAt:   record.UpdatedAt,
	}
}
