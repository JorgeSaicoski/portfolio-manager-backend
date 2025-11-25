package repo

import (
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/models"
	"gorm.io/gorm"
)

type portfolioRepository struct {
	db *gorm.DB
}

func NewPortfolioRepository(db *gorm.DB) PortfolioRepository {
	return &portfolioRepository{
		db: db,
	}
}

func (r *portfolioRepository) Create(portfolio *models.Portfolio) error {
	return r.db.Create(portfolio).Error
}

// For list views - only basic portfolio info
func (r *portfolioRepository) GetByOwnerIDBasic(ownerID string, limit, offset int) ([]models.Portfolio, error) {
	var portfolios []models.Portfolio
	err := r.db.Select("id, title, description, owner_id, created_at, updated_at").
		Where("owner_id = ?", ownerID).
		Limit(limit).Offset(offset).
		Find(&portfolios).Error
	return portfolios, err
}

// For detail views - with relationships using JOIN
func (r *portfolioRepository) GetByIDWithRelations(id uint) (*models.Portfolio, error) {
	var portfolio models.Portfolio
	err := r.db.Select("portfolios.*, sections.id as section_id, sections.title as section_title, categories.id as category_id, categories.title as category_title").
		Joins("LEFT JOIN sections ON sections.portfolio_id = portfolios.id AND sections.deleted_at IS NULL").
		Joins("LEFT JOIN categories ON categories.portfolio_id = portfolios.id AND categories.deleted_at IS NULL").
		Where("portfolios.id = ?", id).
		First(&portfolio).Error
	return &portfolio, err
}

func (r *portfolioRepository) GetByIDBasic(id uint) (*models.Portfolio, error) {
	var portfolio models.Portfolio
	err := r.db.Select("id, owner_id").First(&portfolio, id).Error
	if err != nil {
		return nil, err
	}
	return &portfolio, nil
}

func (r *portfolioRepository) Update(portfolio *models.Portfolio) error {
	return r.db.Model(portfolio).Where("id = ?", portfolio.ID).Updates(portfolio).Error
}

func (r *portfolioRepository) Delete(id uint) error {
	return r.db.Delete(&models.Portfolio{}, id).Error
}

func (r *portfolioRepository) List(limit, offset int) ([]models.Portfolio, error) {
	var portfolios []models.Portfolio
	err := r.db.Preload("Sections").Preload("Categories").
		Limit(limit).Offset(offset).
		Find(&portfolios).Error
	return portfolios, err
}

// CheckDuplicate checks if a portfolio with the same title exists for the same owner
// excluding the portfolio with the given id (useful for updates)
func (r *portfolioRepository) CheckDuplicate(title string, ownerID string, id uint) (bool, error) {
	var count int64
	query := r.db.Model(&models.Portfolio{}).Where("title = ? AND owner_id = ?", title, ownerID)

	// Exclude the current portfolio when checking for duplicates (for updates)
	if id != 0 {
		query = query.Where("id != ?", id)
	}

	err := query.Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
