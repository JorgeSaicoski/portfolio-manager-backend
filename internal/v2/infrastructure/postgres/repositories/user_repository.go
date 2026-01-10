package repositories

import (
	"context"
	"fmt"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/contracts"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/dto"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/infrastructure/postgres/entities"
	"gorm.io/gorm"
)

// userRepository implements the UserRepository contract
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) contracts.UserRepository {
	return &userRepository{db: db}
}

// Create creates a new user
func (r *userRepository) Create(ctx context.Context, email, name, externalID string) (*dto.UserDTO, error) {
	record := &entities.UserRecord{
		Email:      email,
		Name:       name,
		ExternalID: externalID,
	}

	if err := r.db.WithContext(ctx).Create(record).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return r.recordToDTO(record), nil
}

// GetByID retrieves a user by ID
func (r *userRepository) GetByID(ctx context.Context, userID string) (*dto.UserDTO, error) {
	var record entities.UserRecord

	if err := r.db.WithContext(ctx).Where("id = ?", userID).First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return r.recordToDTO(&record), nil
}

// GetByExternalID retrieves a user by external ID (from auth provider)
func (r *userRepository) GetByExternalID(ctx context.Context, externalID string) (*dto.UserDTO, error) {
	var record entities.UserRecord

	if err := r.db.WithContext(ctx).Where("external_id = ?", externalID).First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return r.recordToDTO(&record), nil
}

// Update updates user information
func (r *userRepository) Update(ctx context.Context, userID, email, name string) error {
	updates := map[string]interface{}{}

	if email != "" {
		updates["email"] = email
	}
	if name != "" {
		updates["name"] = name
	}

	if len(updates) == 0 {
		return nil
	}

	if err := r.db.WithContext(ctx).Model(&entities.UserRecord{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// Delete deletes a user
func (r *userRepository) Delete(ctx context.Context, userID string) error {
	if err := r.db.WithContext(ctx).Delete(&entities.UserRecord{}, "id = ?", userID).Error; err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

// recordToDTO converts a user record to DTO
func (r *userRepository) recordToDTO(record *entities.UserRecord) *dto.UserDTO {
	return &dto.UserDTO{
		ID:         fmt.Sprintf("%d", record.ID),
		Email:      record.Email,
		Name:       record.Name,
		ExternalID: record.ExternalID,
		CreatedAt:  record.CreatedAt,
		UpdatedAt:  record.UpdatedAt,
	}
}
