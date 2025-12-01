package handler

import (
	"fmt"
	"net/http"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/infrastructure/audit"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/infrastructure/repo"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type UserHandler struct {
	portfolioRepo      repo.PortfolioRepository
	categoryRepo       repo.CategoryRepository
	sectionRepo        repo.SectionRepository
	projectRepo        repo.ProjectRepository
	imageRepo          repo.ImageRepository
	sectionContentRepo repo.SectionContentRepository
}

func NewUserHandler(
	portfolioRepo repo.PortfolioRepository,
	categoryRepo repo.CategoryRepository,
	sectionRepo repo.SectionRepository,
	projectRepo repo.ProjectRepository,
	imageRepo repo.ImageRepository,
	sectionContentRepo repo.SectionContentRepository,
) *UserHandler {
	return &UserHandler{
		portfolioRepo:      portfolioRepo,
		categoryRepo:       categoryRepo,
		sectionRepo:        sectionRepo,
		projectRepo:        projectRepo,
		imageRepo:          imageRepo,
		sectionContentRepo: sectionContentRepo,
	}
}

// CleanupUserData deletes all data associated with a user
// This endpoint should be called when a user is deleted from Authentik
// Thanks to CASCADE DELETE constraints, deleting portfolios will automatically
// delete all related categories, sections, projects, and section_contents
func (h *UserHandler) CleanupUserData(c *gin.Context) {
	// This endpoint should only be accessible by admin users or webhook
	// For now, we'll use the authenticated user ID
	userID := c.GetString("userID")

	if userID == "" {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "CLEANUP_USER_DATA_MISSING_USER_ID",
			"where":     "backend/internal/application/handler/user.go",
			"function":  "CleanupUserData",
		}).Warn("User ID is required")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "User ID is required",
		})
		return
	}

	logrus.WithFields(logrus.Fields{
		"userID": userID,
	}).Info("Starting user data cleanup")

	// Get all portfolios for this user
	portfolios, _, err := h.portfolioRepo.GetByOwnerIDBasic(userID, 1000, 0)
	if err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "CLEANUP_USER_DATA_DB_ERROR",
			"where":     "backend/internal/application/handler/user.go",
			"function":  "CleanupUserData",
			"userID":    userID,
			"error":     err.Error(),
		}).Error("Failed to retrieve user portfolios for cleanup")

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve user data",
		})
		return
	}

	portfolioCount := len(portfolios)
	var totalImagesDeleted, totalSectionContentDeleted int

	// Delete all portfolios (CASCADE will delete categories, sections, and projects)
	// But we need to manually delete images and section contents due to polymorphic associations
	for _, portfolio := range portfolios {
		// Get and delete all sections for this portfolio
		sections, err := h.sectionRepo.GetByPortfolioID(fmt.Sprintf("%d", portfolio.ID))
		if err == nil {
			for _, section := range sections {
				// Delete section content (not covered by CASCADE due to owner_id check)
				sectionContents, err := h.sectionContentRepo.GetBySectionID(section.ID)
				if err == nil {
					for _, content := range sectionContents {
						if err := h.sectionContentRepo.Delete(content.ID); err != nil {
							logrus.WithFields(logrus.Fields{
								"userID":    userID,
								"contentID": content.ID,
								"error":     err.Error(),
							}).Warn("Failed to delete section content during cleanup")
						} else {
							totalSectionContentDeleted++
						}
					}
				}

				// Delete images associated with sections (polymorphic, not covered by CASCADE)
				sectionImages, err := h.imageRepo.GetByEntity("section", section.ID)
				if err == nil {
					for _, img := range sectionImages {
						if err := h.imageRepo.Delete(img.ID); err != nil {
							logrus.WithFields(logrus.Fields{
								"userID":  userID,
								"imageID": img.ID,
								"error":   err.Error(),
							}).Warn("Failed to delete section image during cleanup")
						} else {
							totalImagesDeleted++
						}
					}
				}
			}
		}

		// Get and delete all categories and their projects
		categories, err := h.categoryRepo.GetByPortfolioID(fmt.Sprintf("%d", portfolio.ID))
		if err == nil {
			for _, category := range categories {
				projects, err := h.projectRepo.GetByCategoryID(fmt.Sprintf("%d", category.ID))
				if err == nil {
					for _, project := range projects {
						// Delete images associated with projects (polymorphic, not covered by CASCADE)
						projectImages, err := h.imageRepo.GetByEntity("project", project.ID)
						if err == nil {
							for _, img := range projectImages {
								if err := h.imageRepo.Delete(img.ID); err != nil {
									logrus.WithFields(logrus.Fields{
										"userID":  userID,
										"imageID": img.ID,
										"error":   err.Error(),
									}).Warn("Failed to delete project image during cleanup")
								} else {
									totalImagesDeleted++
								}
							}
						}
					}
				}
			}
		}

		// Delete images associated with portfolios (polymorphic, not covered by CASCADE)
		portfolioImages, err := h.imageRepo.GetByEntity("portfolio", portfolio.ID)
		if err == nil {
			for _, img := range portfolioImages {
				if err := h.imageRepo.Delete(img.ID); err != nil {
					logrus.WithFields(logrus.Fields{
						"userID":  userID,
						"imageID": img.ID,
						"error":   err.Error(),
					}).Warn("Failed to delete portfolio image during cleanup")
				} else {
					totalImagesDeleted++
				}
			}
		}

		// Now delete the portfolio (CASCADE will handle categories, sections, and projects)
		if err := h.portfolioRepo.Delete(portfolio.ID); err != nil {
			audit.GetErrorLogger().WithFields(logrus.Fields{
				"operation":   "CLEANUP_USER_DATA_DELETE_ERROR",
				"where":       "backend/internal/application/handler/user.go",
				"function":    "CleanupUserData",
				"userID":      userID,
				"portfolioID": portfolio.ID,
				"error":       err.Error(),
			}).Error("Failed to delete portfolio during user cleanup")

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to delete user data",
			})
			return
		}

		logrus.WithFields(logrus.Fields{
			"userID":      userID,
			"portfolioID": portfolio.ID,
			"title":       portfolio.Title,
		}).Info("Portfolio deleted as part of user cleanup (CASCADE: categories, sections, projects)")
	}

	logrus.WithFields(logrus.Fields{
		"userID":                userID,
		"portfolioCount":        portfolioCount,
		"imagesDeleted":         totalImagesDeleted,
		"sectionContentDeleted": totalSectionContentDeleted,
	}).Info("User data cleanup completed successfully")

	c.JSON(http.StatusOK, gin.H{
		"message":               "User data cleaned up successfully",
		"portfoliosDeleted":     portfolioCount,
		"imagesDeleted":         totalImagesDeleted,
		"sectionContentDeleted": totalSectionContentDeleted,
	})
}

// GetUserDataSummary returns a summary of all data owned by a user
// Useful for showing users what will be deleted before cleanup
func (h *UserHandler) GetUserDataSummary(c *gin.Context) {
	userID := c.GetString("userID")

	if userID == "" {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "GET_USER_DATA_SUMMARY_MISSING_USER_ID",
			"where":     "backend/internal/application/handler/user.go",
			"function":  "GetUserDataSummary",
		}).Warn("User ID is required")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "User ID is required",
		})
		return
	}

	// Get all portfolios for this user
	portfolios, _, err := h.portfolioRepo.GetByOwnerIDBasic(userID, 1000, 0)
	if err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "GET_USER_DATA_SUMMARY_DB_ERROR",
			"where":     "backend/internal/application/handler/user.go",
			"function":  "GetUserDataSummary",
			"userID":    userID,
			"error":     err.Error(),
		}).Error("Failed to retrieve user data")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve user data",
		})
		return
	}

	// Count all related data
	var totalCategories, totalSections, totalProjects int

	for _, portfolio := range portfolios {
		categories, err := h.categoryRepo.GetByPortfolioID(fmt.Sprintf("%d", portfolio.ID))
		if err == nil {
			totalCategories += len(categories)

			// Count projects in each category
			for _, category := range categories {
				projects, err := h.projectRepo.GetByCategoryID(fmt.Sprintf("%d", category.ID))
				if err == nil {
					totalProjects += len(projects)
				}
			}
		}

		sections, err := h.sectionRepo.GetByPortfolioID(fmt.Sprintf("%d", portfolio.ID))
		if err == nil {
			totalSections += len(sections)
		}
	}

	summary := map[string]interface{}{
		"userID":     userID,
		"portfolios": len(portfolios),
		"categories": totalCategories,
		"sections":   totalSections,
		"projects":   totalProjects,
		"totalItems": len(portfolios) + totalCategories + totalSections + totalProjects,
	}

	logrus.WithFields(logrus.Fields{
		"userID":  userID,
		"summary": summary,
	}).Info("User data summary requested")

	c.JSON(http.StatusOK, gin.H{
		"message": "User data summary",
		"data":    summary,
	})
}
