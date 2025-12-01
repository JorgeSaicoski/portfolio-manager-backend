package repo

import (
	models2 "github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/models"
)

type PortfolioRepository interface {
	Create(portfolio *models2.Portfolio) error
	GetByID(id uint) (*models2.Portfolio, error)
	GetByIDWithRelations(id uint) (*models2.Portfolio, error)
	GetByOwnerIDBasic(ownerID string, limit, offset int) ([]models2.Portfolio, int64, error)
	GetByIDBasic(id uint) (*models2.Portfolio, error)
	Update(portfolio *models2.Portfolio) error
	Delete(id uint) error
	List(limit, offset int) ([]models2.Portfolio, error)
	CheckDuplicate(title string, ownerID string, id uint) (bool, error)
}

type ProjectRepository interface {
	Create(project *models2.Project) error
	GetByID(id uint) (*models2.Project, error)
	GetByOwnerIDBasic(ownerID string, limit, offset int) ([]models2.Project, int64, error)
	GetByCategoryID(categoryID string) ([]models2.Project, error)
	Update(project *models2.Project) error
	UpdatePosition(id uint, position uint) error
	Delete(id uint) error
	List(limit, offset int) ([]models2.Project, error)
	GetBySkills(skills []string) ([]models2.Project, error)
	GetByClient(client string) ([]models2.Project, error)
	CheckDuplicate(title string, categoryID uint, id uint) (bool, error)
}

type SectionRepository interface {
	Create(section *models2.Section) error
	GetByID(id uint) (*models2.Section, error)
	GetByIDWithRelations(id uint) (*models2.Section, error)
	GetByIDs(ids []uint) ([]*models2.Section, error)
	GetByOwnerID(ownerID string, limit, offset int) ([]models2.Section, int64, error)
	GetByPortfolioID(portfolioID string) ([]models2.Section, error)
	GetByPortfolioIDWithRelations(portfolioID string) ([]models2.Section, error)
	GetByType(sectionType string) ([]models2.Section, error)
	Update(section *models2.Section) error
	UpdatePosition(id uint, position uint) error
	BulkUpdatePositions(items []struct {
		ID       uint `json:"id" binding:"required"`
		Position uint `json:"position" binding:"required,min=1"`
	}) error
	Delete(id uint) error
	List(limit, offset int) ([]models2.Section, error)
	CheckDuplicate(title string, portfolioID uint, id uint) (bool, error)
}

type SectionContentRepository interface {
	Create(content *models2.SectionContent) error
	GetByID(id uint) (*models2.SectionContent, error)
	GetBySectionID(sectionID uint) ([]models2.SectionContent, error)
	Update(content *models2.SectionContent) error
	UpdateOrder(id uint, order uint) error
	Delete(id uint) error
	CheckDuplicateOrder(sectionID uint, order uint, id uint) (bool, error)
}

type CategoryRepository interface {
	Create(category *models2.Category) error
	GetByID(id uint) (*models2.Category, error)
	GetByIDBasic(id uint) (*models2.Category, error)
	GetByIDWithRelations(id uint) (*models2.Category, error)
	GetByIDs(ids []uint) ([]*models2.Category, error)
	GetByPortfolioID(portfolioID string) ([]models2.Category, error)
	GetByPortfolioIDWithRelations(portfolioID string) ([]models2.Category, error)
	GetByOwnerIDBasic(ownerID string, limit, offset int) ([]models2.Category, int64, error)
	Update(category *models2.Category) error
	UpdatePosition(id uint, position uint) error
	BulkUpdatePositions(items []struct {
		ID       uint `json:"id" binding:"required"`
		Position uint `json:"position" binding:"required,min=1"`
	}) error
	Delete(id uint) error
	List(limit, offset int) ([]models2.Category, error)
}

type ImageRepository interface {
	Create(image *models2.Image) error
	GetByID(id uint) (*models2.Image, error)
	GetByEntity(entityType string, entityID uint) ([]models2.Image, error)
	GetByOwnerID(ownerID string, limit, offset int) ([]models2.Image, error)
	Update(image *models2.Image) error
	Delete(id uint) error
	CheckOwnership(id uint, ownerID string) (bool, error)
	CheckEntityOwnership(entityID uint, entityType string, ownerID string) (bool, error)
}
