package repo

import (
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/models"
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

func (r *portfolioRepository) GetByID(id uint) (*models.Portfolio, error) {
	var portfolio models.Portfolio
	err := r.db.Preload("Sections").Preload("Categories").First(&portfolio, id).Error
	if err != nil {
		return nil, err
	}
	return &portfolio, nil
}

func (r *portfolioRepository) GetByOwnerID(ownerID string, limit, offset int) ([]*models.Portfolio, error) {
	var portfolios []*models.Portfolio
	err := r.db.Preload("Sections").Preload("Categories").Where("owner_id = ?", ownerID).Find(&portfolios).
		Limit(limit).Offset(offset).
		Find(&portfolios).Error
	return portfolios, err
}

func (r *portfolioRepository) Update(portfolio *models.Portfolio) error {
	return r.db.Model(portfolio).Where("id = ?", portfolio.ID).Updates(portfolio).Error
}

func (r *portfolioRepository) Delete(id uint) error {
	return r.db.Delete(&models.Portfolio{}, id).Error
}

func (r *portfolioRepository) List(limit, offset int) ([]*models.Portfolio, error) {
	var portfolios []*models.Portfolio
	err := r.db.Preload("Sections").Preload("Categories").
		Limit(limit).Offset(offset).
		Find(&portfolios).Error
	return portfolios, err
}
