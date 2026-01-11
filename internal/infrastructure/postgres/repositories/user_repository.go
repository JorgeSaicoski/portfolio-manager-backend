package repositories

import (
	"context"
	"fmt"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/contracts"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/dto"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/infrastructure/postgres/entities"
	"gorm.io/gorm"
)

// userRepository is the GORM implementation of UserRepository
// It implements the contract defined in the application layer
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository instance
// Returns the interface type (contracts.UserRepository), not the concrete type
func NewUserRepository(db *gorm.DB) contracts.UserRepository {
	return &userRepository{db: db}
}

// Create creates a new user in the database
func (r *userRepository) Create(ctx context.Context, email, name, externalID string) (*dto.UserDTO, error) {
	// Convert application parameters to infrastructure entity
	record := &entities.UserRecord{
		Email:      email,
		Name:       name,
		ExternalID: externalID,
	}

	// Persist to database
	if err := r.db.WithContext(ctx).Create(record).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Convert infrastructure entity back to application DTO
	return r.recordToDTO(record), nil
}

// GetByID retrieves a user by their ID
func (r *userRepository) GetByID(ctx context.Context, userID string) (*dto.UserDTO, error) {
	var record entities.UserRecord

	if err := r.db.WithContext(ctx).First(&record, "id = ?", userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user with ID %s not found", userID)
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return r.recordToDTO(&record), nil
}

// GetByExternalID retrieves a user by their external ID (from auth provider)
func (r *userRepository) GetByExternalID(ctx context.Context, externalID string) (*dto.UserDTO, error) {
	var record entities.UserRecord

	if err := r.db.WithContext(ctx).Where("external_id = ?", externalID).First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user with external ID %s not found", externalID)
		}
		return nil, fmt.Errorf("failed to get user by external ID: %w", err)
	}

	return r.recordToDTO(&record), nil
}

// Update updates an existing user
func (r *userRepository) Update(ctx context.Context, userID, email, name string) error {
	updates := map[string]interface{}{}

	// Only update non-empty fields
	if email != "" {
		updates["email"] = email
	}
	if name != "" {
		updates["name"] = name
	}

	if len(updates) == 0 {
		return nil // Nothing to update
	}

	result := r.db.WithContext(ctx).
		Model(&entities.UserRecord{}).
		Where("id = ?", userID).
		Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("failed to update user: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("user with ID %s not found", userID)
	}

	return nil
}

// Delete deletes a user by their ID (soft delete with GORM)
func (r *userRepository) Delete(ctx context.Context, userID string) error {
	result := r.db.WithContext(ctx).Delete(&entities.UserRecord{}, "id = ?", userID)

	if result.Error != nil {
		return fmt.Errorf("failed to delete user: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("user with ID %s not found", userID)
	}

	return nil
}

// recordToDTO converts a UserRecord (infrastructure) to UserDTO (application)
func (r *userRepository) recordToDTO(record *entities.UserRecord) *dto.UserDTO {
	return &dto.UserDTO{
		ID:         record.ID,
		Email:      record.Email,
		Name:       record.Name,
		ExternalID: record.ExternalID,
		CreatedAt:  record.CreatedAt,
		UpdatedAt:  record.UpdatedAt,
	}
}
