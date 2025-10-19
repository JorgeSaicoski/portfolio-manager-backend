package test

import (
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/models"
	"gorm.io/gorm"
)

// Portfolio fixtures
func CreateTestPortfolio(db *gorm.DB, ownerID string) *models.Portfolio {
	portfolio := &models.Portfolio{
		Title:       "Test Portfolio",
		Description: stringPtr("Test description"),
		OwnerID:     ownerID,
	}
	db.Create(portfolio)
	return portfolio
}

func CreateTestPortfolioWithTitle(db *gorm.DB, ownerID string, title string) *models.Portfolio {
	portfolio := &models.Portfolio{
		Title:       title,
		Description: stringPtr("Test description"),
		OwnerID:     ownerID,
	}
	db.Create(portfolio)
	return portfolio
}

// Category fixtures
func CreateTestCategory(db *gorm.DB, portfolioID uint, ownerID string) *models.Category {
	category := &models.Category{
		Title:       "Test Category",
		Description: stringPtr("Test category description"),
		PortfolioID: portfolioID,
		OwnerID:     ownerID,
	}
	db.Create(category)
	return category
}

func CreateTestCategoryWithTitle(db *gorm.DB, portfolioID uint, ownerID string, title string) *models.Category {
	category := &models.Category{
		Title:       title,
		Description: stringPtr("Test category description"),
		PortfolioID: portfolioID,
		OwnerID:     ownerID,
	}
	db.Create(category)
	return category
}

// Project fixtures
func CreateTestProject(db *gorm.DB, categoryID uint, ownerID string) *models.Project {
	project := &models.Project{
		Title:       "Test Project",
		Description: "Test project description",
		Images:      []string{},
		Skills:      []string{"Go", "React"},
		CategoryID:  categoryID,
		OwnerID:     ownerID,
		MainImage:   "https://example.com/image.png",
		Client:      "Test Client",
		Link:        "https://example.com",
	}
	db.Create(project)
	return project
}

func CreateTestProjectWithTitle(db *gorm.DB, categoryID uint, ownerID string, title string) *models.Project {
	project := &models.Project{
		Title:       title,
		Description: "Test project description",
		Images:      []string{},
		Skills:      []string{"Go", "React"},
		CategoryID:  categoryID,
		OwnerID:     ownerID,
		MainImage:   "https://example.com/image.png",
		Client:      "Test Client",
		Link:        "https://example.com",
	}
	db.Create(project)
	return project
}

// Section fixtures
func CreateTestSection(db *gorm.DB, portfolioID uint, ownerID string) *models.Section {
	section := &models.Section{
		Title:       "Test Section",
		Description: stringPtr("Test section description"),
		Type:        "text",
		PortfolioID: portfolioID,
		OwnerID:     ownerID,
	}
	db.Create(section)
	return section
}

func CreateTestSectionWithTitle(db *gorm.DB, portfolioID uint, ownerID string, title string) *models.Section {
	section := &models.Section{
		Title:       title,
		Description: stringPtr("Test section description"),
		Type:        "text",
		PortfolioID: portfolioID,
		OwnerID:     ownerID,
	}
	db.Create(section)
	return section
}

func stringPtr(s string) *string {
	return &s
}
