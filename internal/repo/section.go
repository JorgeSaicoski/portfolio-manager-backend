package repo

import (
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/models"
	"gorm.io/gorm"
)

type sectionRepository struct {
	db *gorm.DB
}

func NewSectionRepository(db *gorm.DB) SectionRepository {
	return &sectionRepository{
		db: db,
	}
}

func (r *sectionRepository) Create(section *models.Section) error {
	return r.db.Create(section).Error
}

// For list views - only basic portfolio info
func (r *sectionRepository) GetByPortfolioIDBasic(portfolioID string) ([]*models.Section, error) {
	var sections []*models.Section
	err := r.db.Select("id, title, owner_id, created_at, updated_at").
		Where("portfolio_id = ?", portfolioID).
		Find(&sections).Error
	return sections, err
}

// For detail views - with relationships using JOIN
func (r *sectionRepository) GetByIDWithRelations(id uint) (*models.Section, error) {
	var section models.Section
	err := r.db.Preload("Portfolio").
		Where("id = ?", id).
		First(&section).Error
	return &section, err
}

func (r *sectionRepository) GetByPortfolioIDWithRelations(portfolioID string) ([]*models.Section, error) {
	var sections []*models.Section
	err := r.db.Preload("Portfolio").
		Where("portfolio_id = ?", portfolioID).
		Find(&sections).Error
	return sections, err
}

func (r *sectionRepository) GetByType(sectionType string) ([]*models.Section, error) {
	var sections []*models.Section
	err := r.db.Where("type = ?", sectionType).
		Find(&sections).Error
	return sections, err
}

func (r *sectionRepository) Update(section *models.Section) error {
	return r.db.Model(section).Where("id = ?", section.ID).Updates(section).Error
}

func (r *sectionRepository) Delete(id uint) error {
	return r.db.Delete(&models.Section{}, id).Error
}

func (r *sectionRepository) List(limit, offset int) ([]*models.Section, error) {
	var sections []*models.Section
	err := r.db.Limit(limit).Offset(offset).
		Find(&sections).Error
	return sections, err
}
