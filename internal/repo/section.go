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

// GetByOwnerIDBasic For list views - only basic section info for a specific owner
func (r *sectionRepository) GetByOwnerIDBasic(ownerID string, limit, offset int) ([]*models.Section, error) {
	var sections []*models.Section
	err := r.db.Where("owner_id = ?", ownerID).
		Order("position ASC, created_at ASC").
		Limit(limit).Offset(offset).
		Find(&sections).Error
	return sections, err
}

// For list views - only basic portfolio info
func (r *sectionRepository) GetByPortfolioIDBasic(portfolioID string) ([]*models.Section, error) {
	var sections []*models.Section
	err := r.db.Select("id, title, position, owner_id, created_at, updated_at").
		Where("portfolio_id = ?", portfolioID).
		Order("position ASC, created_at ASC").
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
		Order("position ASC, created_at ASC").
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

// UpdatePosition updates only the position field of a section
func (r *sectionRepository) UpdatePosition(id uint, position uint) error {
	return r.db.Model(&models.Section{}).Where("id = ?", id).Update("position", position).Error
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

// CheckDuplicate checks if a section with the same title exists for the same portfolio
// excluding the section with the given id (useful for updates)
func (r *sectionRepository) CheckDuplicate(title string, portfolioID uint, id uint) (bool, error) {
	var count int64
	query := r.db.Model(&models.Section{}).Where("title = ? AND portfolio_id = ?", title, portfolioID)

	// Exclude the current section when checking for duplicates (for updates)
	if id != 0 {
		query = query.Where("id != ?", id)
	}

	err := query.Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
