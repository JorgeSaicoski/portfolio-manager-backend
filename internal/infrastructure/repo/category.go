package repo

import (
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/models"
	"gorm.io/gorm"
)

type categoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepository{
		db: db,
	}
}

func (r *categoryRepository) Create(category *models.Category) error {
	return r.db.Create(category).Error
}

// GetByID For basic category info
func (r *categoryRepository) GetByID(id uint) (*models.Category, error) {
	var category models.Category
	err := r.db.Select("id, title, description, position, owner_id, portfolio_id, created_at, updated_at").
		Where("id = ?", id).
		First(&category).Error
	return &category, err
}

// GetByIDBasic For authorization checks - only id and owner_id
func (r *categoryRepository) GetByIDBasic(id uint) (*models.Category, error) {
	var category models.Category
	err := r.db.Select("id, owner_id").
		Where("id = ?", id).
		First(&category).Error
	return &category, err
}

// GetByIDWithRelations For detail views - with projects preloaded
func (r *categoryRepository) GetByIDWithRelations(id uint) (*models.Category, error) {
	var category models.Category
	err := r.db.Preload("Projects").
		Where("id = ?", id).
		First(&category).Error
	return &category, err
}

// GetByPortfolioID For list views - only basic category info
func (r *categoryRepository) GetByPortfolioID(portfolioID string) ([]models.Category, error) {
	var categories []models.Category
	err := r.db.Select("id, title, description, position, owner_id, portfolio_id, created_at, updated_at").
		Where("portfolio_id = ?", portfolioID).
		Order("position ASC, created_at ASC").
		Find(&categories).Error
	return categories, err
}

// GetByPortfolioIDWithRelations For detail views - with projects preloaded
func (r *categoryRepository) GetByPortfolioIDWithRelations(portfolioID string) ([]models.Category, error) {
	var categories []models.Category
	err := r.db.Preload("Projects", func(db *gorm.DB) *gorm.DB {
		return db.Order("projects.position ASC, projects.created_at ASC")
	}).
		Where("portfolio_id = ?", portfolioID).
		Order("position ASC, created_at ASC").
		Find(&categories).Error
	return categories, err
}

func (r *categoryRepository) Update(category *models.Category) error {
	return r.db.Model(category).Where("id = ?", category.ID).Updates(category).Error
}

// UpdatePosition updates only the position field of a category
func (r *categoryRepository) UpdatePosition(id uint, position uint) error {
	return r.db.Model(&models.Category{}).Where("id = ?", id).Update("position", position).Error
}

// GetByIDs fetches multiple categories by their IDs
func (r *categoryRepository) GetByIDs(ids []uint) ([]*models.Category, error) {
	var categories []*models.Category
	if err := r.db.Where("id IN ?", ids).Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

// BulkUpdatePositions updates positions for multiple categories in a transaction
func (r *categoryRepository) BulkUpdatePositions(items []struct {
	ID       uint
	Position uint
}) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, item := range items {
			if err := tx.Model(&models.Category{}).
				Where("id = ?", item.ID).
				Update("position", item.Position).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *categoryRepository) Delete(id uint) error {
	return r.db.Delete(&models.Category{}, id).Error
}

func (r *categoryRepository) List(limit, offset int) ([]models.Category, error) {
	var categories []models.Category
	err := r.db.Select("id, title, description, position, owner_id, portfolio_id, created_at, updated_at").
		Limit(limit).Offset(offset).
		Find(&categories).Error
	return categories, err
}

// GetByOwnerIDBasic For list views - categories owned by user
func (r *categoryRepository) GetByOwnerIDBasic(ownerID string, limit, offset int) ([]models.Category, error) {
	var categories []models.Category
	err := r.db.Select("id, title, description, position, owner_id, portfolio_id, created_at, updated_at").
		Where("owner_id = ?", ownerID).
		Limit(limit).Offset(offset).
		Find(&categories).Error
	return categories, err
}
