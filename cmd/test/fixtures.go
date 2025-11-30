package test

import (
	"fmt"

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
		Skills:      []string{"Go", "React"},
		CategoryID:  categoryID,
		OwnerID:     ownerID,
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
		Skills:      []string{"Go", "React"},
		CategoryID:  categoryID,
		OwnerID:     ownerID,
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

// Image fixtures
func CreateTestImage(db *gorm.DB, entityID uint, entityType string, ownerID string) *models2.Image {
	image := &models2.Image{
		URL:          "/uploads/images/original/test.png",
		ThumbnailURL: "/uploads/images/thumbnail/test.png",
		FileName:     "test.png",
		FileSize:     1024,
		MimeType:     "image/png",
		Alt:          "Test image",
		OwnerID:      ownerID,
		Type:         "image",
		EntityID:     entityID,
		EntityType:   entityType,
		IsMain:       false,
	}
	db.Create(image)
	return image
}

func CreateTestImageWithAlt(db *gorm.DB, entityID uint, entityType string, ownerID string, alt string) *models2.Image {
	image := &models2.Image{
		URL:          "/uploads/images/original/test.png",
		ThumbnailURL: "/uploads/images/thumbnail/test.png",
		FileName:     "test.png",
		FileSize:     1024,
		MimeType:     "image/png",
		Alt:          alt,
		OwnerID:      ownerID,
		Type:         "image",
		EntityID:     entityID,
		EntityType:   entityType,
		IsMain:       false,
	}
	db.Create(image)
	return image
}

// SectionContent fixtures
func CreateTestSectionContent(db *gorm.DB, sectionID uint, ownerID string) *models2.SectionContent {
	content := &models2.SectionContent{
		SectionID: sectionID,
		Type:      "text",
		Content:   "Test content",
		Order:     0,
		OwnerID:   ownerID,
	}
	db.Create(content)
	return content
}

func CreateTestSectionContentWithOrder(db *gorm.DB, sectionID uint, ownerID string, order uint) *models2.SectionContent {
	content := &models2.SectionContent{
		SectionID: sectionID,
		Type:      "text",
		Content:   fmt.Sprintf("Content order %d", order),
		Order:     order,
		OwnerID:   ownerID,
	}
	db.Create(content)
	return content
}

func CreateTestSectionContentWithImage(db *gorm.DB, sectionID uint, ownerID string) *models2.SectionContent {
	// Create an image first
	image := CreateTestImage(db, sectionID, "section", ownerID)

	metadata := fmt.Sprintf(`{"image_id": %d}`, image.ID)
	content := &models2.SectionContent{
		SectionID: sectionID,
		Type:      "image",
		Content:   "Image description",
		Order:     0,
		Metadata:  &metadata,
		OwnerID:   ownerID,
	}
	db.Create(content)
	return content
}

func stringPtr(s string) *string {
	return &s
}
