package repositories

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/contracts"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/dto"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/infrastructure/postgres/entities"
)

// portfolioRepository is the GORM implementation of PortfolioRepository
// It implements the contract defined in the application layer
type portfolioRepository struct {
	db *gorm.DB
}

// NewPortfolioRepository creates a new portfolio repository instance
// Returns the interface type (contracts.PortfolioRepository), not the concrete type
func NewPortfolioRepository(db *gorm.DB) contracts.PortfolioRepository {
	return &portfolioRepository{db: db}
}

// Create creates a new portfolio in the database
func (r *portfolioRepository) Create(ctx context.Context, input dto.CreatePortfolioInput) (*dto.PortfolioDTO, error) {
	// Convert application DTO to infrastructure entity
	record := &entities.PortfolioRecord{
		Title:       input.Title,
		Description: input.Description,
		OwnerID:     input.OwnerID,
	}

	// Persist to database
	if err := r.db.WithContext(ctx).Create(record).Error; err != nil {
		return nil, fmt.Errorf("failed to create portfolio: %w", err)
	}

	// Convert infrastructure entity back to application DTO
	return r.recordToDTO(record), nil
}

// GetByID retrieves a portfolio by its ID
func (r *portfolioRepository) GetByID(ctx context.Context, id uint) (*dto.PortfolioDTO, error) {
	var record entities.PortfolioRecord

	if err := r.db.WithContext(ctx).First(&record, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("portfolio with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to get portfolio: %w", err)
	}

	return r.recordToDTO(&record), nil
}

// GetByOwnerID retrieves all portfolios owned by a user with pagination
func (r *portfolioRepository) GetByOwnerID(ctx context.Context, ownerID string, pagination dto.PaginationDTO) ([]dto.PortfolioDTO, int64, error) {
	var records []entities.PortfolioRecord
	var total int64

	// Count total portfolios for this owner
	if err := r.db.WithContext(ctx).
		Model(&entities.PortfolioRecord{}).
		Where("owner_id = ?", ownerID).
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count portfolios: %w", err)
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
		return nil, 0, fmt.Errorf("failed to list portfolios: %w", err)
	}

	// Convert records to DTOs
	dtos := make([]dto.PortfolioDTO, len(records))
	for i, record := range records {
		dtos[i] = *r.recordToDTO(&record)
	}

	return dtos, total, nil
}

// Update updates an existing portfolio
func (r *portfolioRepository) Update(ctx context.Context, input dto.UpdatePortfolioInput) error {
	updates := map[string]interface{}{}

	// Only update non-empty fields
	if input.Title != "" {
		updates["title"] = input.Title
	}
	if input.Description != "" {
		updates["description"] = input.Description
	}

	if len(updates) == 0 {
		return nil // Nothing to update
	}

	result := r.db.WithContext(ctx).
		Model(&entities.PortfolioRecord{}).
		Where("id = ?", input.ID).
		Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("failed to update portfolio: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("portfolio with ID %d not found", input.ID)
	}

	return nil
}

// Delete deletes a portfolio by its ID (soft delete with GORM)
func (r *portfolioRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&entities.PortfolioRecord{}, id)

	if result.Error != nil {
		return fmt.Errorf("failed to delete portfolio: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("portfolio with ID %d not found", id)
	}

	return nil
}

// CheckTitleDuplicate checks if a portfolio title already exists for a user
func (r *portfolioRepository) CheckTitleDuplicate(ctx context.Context, title, ownerID string, excludeID uint) (bool, error) {
	var count int64

	query := r.db.WithContext(ctx).
		Model(&entities.PortfolioRecord{}).
		Where("title = ? AND owner_id = ?", title, ownerID)

	// Exclude the current portfolio when updating
	if excludeID > 0 {
		query = query.Where("id != ?", excludeID)
	}

	if err := query.Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check duplicate title: %w", err)
	}

	return count > 0, nil
}

// recordToDTO converts a PortfolioRecord (infrastructure) to PortfolioDTO (application)
func (r *portfolioRepository) recordToDTO(record *entities.PortfolioRecord) *dto.PortfolioDTO {
	return &dto.PortfolioDTO{
		ID:          record.ID,
		Title:       record.Title,
		Description: record.Description,
		OwnerID:     record.OwnerID,
		CreatedAt:   record.CreatedAt,
		UpdatedAt:   record.UpdatedAt,
	}
}
