package repo

import (
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/models"
	"gorm.io/gorm"
)

type sectionContentRepository struct {
	db *gorm.DB
}

func NewSectionContentRepository(db *gorm.DB) SectionContentRepository {
	return &sectionContentRepository{
		db: db,
	}
}

func (r *sectionContentRepository) Create(content *models.SectionContent) error {
	return r.db.Create(content).Error
}

// GetByID retrieves a single content block by ID
func (r *sectionContentRepository) GetByID(id uint) (*models.SectionContent, error) {
	var content models.SectionContent
	err := r.db.Where("id = ?", id).
		First(&content).Error
	return &content, err
}

// GetBySectionID retrieves all content blocks for a section, ordered by position
func (r *sectionContentRepository) GetBySectionID(sectionID uint) ([]*models.SectionContent, error) {
	var contents []*models.SectionContent
	err := r.db.Where("section_id = ?", sectionID).
		Order("\"order\" ASC, created_at ASC").
		Find(&contents).Error
	return contents, err
}

func (r *sectionContentRepository) Update(content *models.SectionContent) error {
	return r.db.Model(content).Where("id = ?", content.ID).Updates(content).Error
}

// UpdateOrder updates only the order field of a content block
func (r *sectionContentRepository) UpdateOrder(id uint, order uint) error {
	return r.db.Model(&models.SectionContent{}).Where("id = ?", id).Update("order", order).Error
}

func (r *sectionContentRepository) Delete(id uint) error {
	return r.db.Delete(&models.SectionContent{}, id).Error
}

// CheckDuplicateOrder checks if another content block has the same order in the section
// Useful to prevent order conflicts, though not strictly enforced
func (r *sectionContentRepository) CheckDuplicateOrder(sectionID uint, order uint, id uint) (bool, error) {
	var count int64
	query := r.db.Model(&models.SectionContent{}).
		Where("section_id = ? AND \"order\" = ?", sectionID, order)

	// Exclude the current content when checking for duplicates (for updates)
	if id != 0 {
		query = query.Where("id != ?", id)
	}

	err := query.Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
