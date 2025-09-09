package repo

import (
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/models"
)

type PortfolioRepository interface {
	Create(portfolio *models.Portfolio) error
	GetByIDWithRelations(id uint) (*models.Portfolio, error)
	GetByOwnerIDBasic(ownerID string, limit, offset int) ([]*models.Portfolio, error)
	GetByIDBasic(id uint) (*models.Portfolio, error)
	Update(portfolio *models.Portfolio) error
	Delete(id uint) error
	List(limit, offset int) ([]*models.Portfolio, error)
}

type ProjectRepository interface {
	Create(project *models.Project) error
	GetByID(id uint) (*models.Project, error)
	GetByCategoryID(categoryID string) ([]*models.Project, error)
	Update(project *models.Project) error
	Delete(id uint) error
	List(limit, offset int) ([]*models.Project, error)
	GetBySkills(skills []string) ([]*models.Project, error)
	GetByClient(client string) ([]*models.Project, error)
}

type SectionRepository interface {
	Create(section *models.Section) error
	GetByIDWithRelations(id uint) (*models.Section, error)
	GetByPortfolioIDBasic(portfolioID string) ([]*models.Section, error)
	GetByPortfolioIDWithRelations(portfolioID string) ([]*models.Section, error)
	GetByType(sectionType string) ([]*models.Section, error)
	Update(section *models.Section) error
	Delete(id uint) error
	List(limit, offset int) ([]*models.Section, error)
}

type CategoryRepository interface {
	Create(category *models.Category) error
	GetByID(id uint) (*models.Category, error)
	GetByPortfolioID(portfolioID string) ([]*models.Category, error)
	Update(category *models.Category) error
	Delete(id uint) error
	List(limit, offset int) ([]*models.Category, error)
}
