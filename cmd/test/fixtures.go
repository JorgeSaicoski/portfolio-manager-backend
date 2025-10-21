package test

import (
	models2 "github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/models"
	"gorm.io/gorm"
)

// Portfolio fixtures
func CreateTestPortfolio(db *gorm.DB, ownerID string) *models2.Portfolio {
	portfolio := &models2.Portfolio{
		Title:       "Test Portfolio",
		Description: stringPtr("Test description"),
		OwnerID:     ownerID,
	}
	db.Create(portfolio)
	return portfolio
}

func CreateTestPortfolioWithTitle(db *gorm.DB, ownerID string, title string) *models2.Portfolio {
	portfolio := &models2.Portfolio{
		Title:       title,
		Description: stringPtr("Test description"),
		OwnerID:     ownerID,
	}
	db.Create(portfolio)
	return portfolio
}

// Category fixtures
func CreateTestCategory(db *gorm.DB, portfolioID uint, ownerID string) *models2.Category {
	category := &models2.Category{
		Title:       "Test Category",
		Description: stringPtr("Test category description"),
		PortfolioID: portfolioID,
		OwnerID:     ownerID,
	}
	db.Create(category)
	return category
}

func CreateTestCategoryWithTitle(db *gorm.DB, portfolioID uint, ownerID string, title string) *models2.Category {
	category := &models2.Category{
		Title:       title,
		Description: stringPtr("Test category description"),
		PortfolioID: portfolioID,
		OwnerID:     ownerID,
	}
	db.Create(category)
	return category
}

// Project fixtures
func CreateTestProject(db *gorm.DB, categoryID uint, ownerID string) *models2.Project {
	project := &models2.Project{
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

func CreateTestProjectWithTitle(db *gorm.DB, categoryID uint, ownerID string, title string) *models2.Project {
	project := &models2.Project{
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
func CreateTestSection(db *gorm.DB, portfolioID uint, ownerID string) *models2.Section {
	section := &models2.Section{
		Title:       "Test Section",
		Description: stringPtr("Test section description"),
		Type:        "text",
		PortfolioID: portfolioID,
		OwnerID:     ownerID,
	}
	db.Create(section)
	return section
}

func CreateTestSectionWithTitle(db *gorm.DB, portfolioID uint, ownerID string, title string) *models2.Section {
	section := &models2.Section{
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
