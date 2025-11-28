package handler

import (
	"net/http"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/infrastructure/audit"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/infrastructure/repo"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type UserHandler struct {
	portfolioRepo repo.PortfolioRepository
	categoryRepo  repo.CategoryRepository
	sectionRepo   repo.SectionRepository
	projectRepo   repo.ProjectRepository
}

func NewUserHandler(
	portfolioRepo repo.PortfolioRepository,
	categoryRepo repo.CategoryRepository,
	sectionRepo repo.SectionRepository,
	projectRepo repo.ProjectRepository,
) *UserHandler {
	return &UserHandler{
		portfolioRepo: portfolioRepo,
		categoryRepo:  categoryRepo,
		sectionRepo:   sectionRepo,
		projectRepo:   projectRepo,
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
	portfolios, err := h.portfolioRepo.GetByOwnerIDBasic(userID, 1000, 0)
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

	// Delete all portfolios (CASCADE will delete all related data)
	for _, portfolio := range portfolios {
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
		}).Info("Portfolio deleted as part of user cleanup (CASCADE: all related data)")
	}

	logrus.WithFields(logrus.Fields{
		"userID":         userID,
		"portfolioCount": portfolioCount,
	}).Info("User data cleanup completed successfully")

	c.JSON(http.StatusOK, gin.H{
		"message":           "User data cleaned up successfully",
		"portfoliosDeleted": portfolioCount,
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
	portfolios, err := h.portfolioRepo.GetByOwnerIDBasic(userID, 1000, 0)
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
		categories, err := h.categoryRepo.GetByPortfolioID(string(rune(portfolio.ID)))
		if err == nil {
			totalCategories += len(categories)

			// Count projects in each category
			for _, category := range categories {
				projects, err := h.projectRepo.GetByCategoryID(string(rune(category.ID)))
				if err == nil {
					totalProjects += len(projects)
				}
			}
		}

		sections, err := h.sectionRepo.GetByPortfolioID(string(rune(portfolio.ID)))
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
