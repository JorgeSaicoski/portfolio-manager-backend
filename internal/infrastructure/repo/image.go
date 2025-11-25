package repo

import (
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/models"
	"gorm.io/gorm"
)

type imageRepository struct {
	db *gorm.DB
}

func NewImageRepository(db *gorm.DB) ImageRepository {
	return &imageRepository{
		db: db,
	}
}

func (r *imageRepository) Create(image *models.Image) error {
	return r.db.Create(image).Error
}

func (r *imageRepository) GetByID(id uint) (*models.Image, error) {
	var image models.Image
	err := r.db.Where("id = ?", id).First(&image).Error
	return &image, err
}

// GetByEntity retrieves all images for a specific entity (e.g., project, portfolio, section)
func (r *imageRepository) GetByEntity(entityType string, entityID uint) ([]models.Image, error) {
	var images []models.Image
	err := r.db.Where("entity_type = ? AND entity_id = ?", entityType, entityID).
		Order("is_main DESC, created_at ASC").
		Find(&images).Error
	return images, err
}

// GetByOwnerID retrieves all images owned by a specific user
func (r *imageRepository) GetByOwnerID(ownerID string, limit, offset int) ([]models.Image, error) {
	var images []models.Image
	err := r.db.Where("owner_id = ?", ownerID).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&images).Error
	return images, err
}

func (r *imageRepository) Update(image *models.Image) error {
	return r.db.Model(image).Where("id = ?", image.ID).Updates(image).Error
}

func (r *imageRepository) Delete(id uint) error {
	return r.db.Delete(&models.Image{}, id).Error
}

// CheckOwnership verifies if a user owns the specified image
func (r *imageRepository) CheckOwnership(id uint, ownerID string) (bool, error) {
	var count int64
	err := r.db.Model(&models.Image{}).
		Where("id = ? AND owner_id = ?", id, ownerID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// CheckEntityOwnership verifies if a user owns the entity (project, portfolio, section)
func (r *imageRepository) CheckEntityOwnership(entityID uint, entityType string, ownerID string) (bool, error) {
	var count int64
	var err error

	switch entityType {
	case "project":
		err = r.db.Model(&models.Project{}).
			Where("id = ? AND owner_id = ?", entityID, ownerID).
			Count(&count).Error
	case "portfolio":
		err = r.db.Model(&models.Portfolio{}).
			Where("id = ? AND owner_id = ?", entityID, ownerID).
			Count(&count).Error
	case "section":
		err = r.db.Model(&models.Section{}).
			Where("id = ? AND owner_id = ?", entityID, ownerID).
			Count(&count).Error
	default:
		return false, nil
	}

	if err != nil {
		return false, err
	}
	return count > 0, nil
}
