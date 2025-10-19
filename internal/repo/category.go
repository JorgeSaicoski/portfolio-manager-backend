package repo

import (
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/models"
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
	err := r.db.Select("id, title, description, owner_id, portfolio_id, created_at, updated_at").
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
func (r *categoryRepository) GetByPortfolioID(portfolioID string) ([]*models.Category, error) {
	var categories []*models.Category
	err := r.db.Select("id, title, description, owner_id, portfolio_id, created_at, updated_at").
		Where("portfolio_id = ?", portfolioID).
		Find(&categories).Error
	return categories, err
}

// GetByPortfolioIDWithRelations For detail views - with projects preloaded
func (r *categoryRepository) GetByPortfolioIDWithRelations(portfolioID string) ([]*models.Category, error) {
	var categories []*models.Category
	err := r.db.Preload("Projects").
		Where("portfolio_id = ?", portfolioID).
		Find(&categories).Error
	return categories, err
}

func (r *categoryRepository) Update(category *models.Category) error {
	return r.db.Model(category).Where("id = ?", category.ID).Updates(category).Error
}

func (r *categoryRepository) Delete(id uint) error {
	return r.db.Delete(&models.Category{}, id).Error
}

func (r *categoryRepository) List(limit, offset int) ([]*models.Category, error) {
	var categories []*models.Category
	err := r.db.Limit(limit).Offset(offset).
		Find(&categories).Error
	return categories, err
}

// GetByOwnerIDBasic For list views - categories owned by user
func (r *categoryRepository) GetByOwnerIDBasic(ownerID string, limit, offset int) ([]*models.Category, error) {
	var categories []*models.Category
	err := r.db.Select("id, title, description, owner_id, portfolio_id, created_at, updated_at").
		Where("owner_id = ?", ownerID).
		Limit(limit).Offset(offset).
		Find(&categories).Error
	return categories, err
}
