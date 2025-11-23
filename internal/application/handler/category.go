package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/models"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/infrastructure/metrics"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/infrastructure/repo"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/shared/response"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/shared/validator"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type CategoryHandler struct {
	repo          repo.CategoryRepository
	portfolioRepo repo.PortfolioRepository
	metrics       *metrics.Collector
}

func NewCategoryHandler(repo repo.CategoryRepository, portfolioRepo repo.PortfolioRepository, metrics *metrics.Collector) *CategoryHandler {
	return &CategoryHandler{
		repo:          repo,
		portfolioRepo: portfolioRepo,
		metrics:       metrics,
	}
}

func (h *CategoryHandler) GetByUser(c *gin.Context) {
	userID := c.GetString("userID") // From auth middleware

	// Parse pagination parameters
	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	offset := (page - 1) * limit

	categories, err := h.repo.GetByOwnerIDBasic(userID, limit, offset)
	if err != nil {
		response.InternalError(c, "Failed to retrieve categories")
		return
	}

	response.SuccessWithPagination(c, http.StatusOK, "categories", categories, page, limit)
}
func (h *CategoryHandler) Update(c *gin.Context) {
	userID := c.GetString("userID") // From auth middleware
	categoryID := c.Param("id")

	// Parse category ID
	id, err := strconv.Atoi(categoryID)
	if err != nil {
		response.BadRequest(c, "Invalid category ID")
		return
	}

	// Parse request body
	var updateData models.Category
	if err := c.ShouldBindJSON(&updateData); err != nil {
		response.BadRequest(c, "Invalid request data")
		return
	}

	// Set the ID and owner
	updateData.ID = uint(id)
	updateData.OwnerID = userID

	// Validate category data
	if err := validator.ValidateCategory(&updateData); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Check if category exists and belongs to user
	existing, err := h.repo.GetByIDBasic(uint(id))
	if err != nil {
		response.NotFound(c, "Category not found")
		return
	}

	if existing.OwnerID != userID {
		response.Forbidden(c, "Access denied")
		return
	}

	// Update category
	if err := h.repo.Update(&updateData); err != nil {
		response.InternalError(c, "Failed to update category")
		return
	}

	response.OK(c, "category", &updateData, "Category updated successfully")
}

func (h *CategoryHandler) Create(c *gin.Context) {
	userID := c.GetString("userID") // From auth middleware

	// Log initial request
	logrus.WithFields(logrus.Fields{
		"userID": userID,
		"path":   c.Request.URL.Path,
	}).Info("Category creation request received")

	// Parse request body
	var newCategory models.Category
	if err := c.ShouldBindJSON(&newCategory); err != nil {
		logrus.WithFields(logrus.Fields{
			"userID": userID,
			"error":  err.Error(),
		}).Error("Failed to parse category creation request")
		response.BadRequest(c, "Invalid request data")
		return
	}

	// Log parsed request
	reqJSON, _ := json.Marshal(newCategory)
	logrus.WithFields(logrus.Fields{
		"userID":  userID,
		"request": string(reqJSON),
	}).Info("Parsed category creation request")

	// Set the owner
	newCategory.OwnerID = userID

	// Validate category data
	if err := validator.ValidateCategory(&newCategory); err != nil {
		logrus.WithFields(logrus.Fields{
			"userID": userID,
			"error":  err.Error(),
		}).Error("Category validation failed")
		response.BadRequest(c, err.Error())
		return
	}

	// Validate that the portfolio exists and belongs to the user
	portfolio, err := h.portfolioRepo.GetByIDBasic(newCategory.PortfolioID)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"userID":       userID,
			"portfolio_id": newCategory.PortfolioID,
		}).Error("Portfolio not found")
		response.NotFound(c, "Portfolio not found")
		return
	}

	if portfolio.OwnerID != userID {
		logrus.WithFields(logrus.Fields{
			"userID":       userID,
			"portfolio_id": newCategory.PortfolioID,
			"owner_id":     portfolio.OwnerID,
		}).Error("Access denied to portfolio")
		response.Forbidden(c, "Access denied")
		return
	}

	logrus.WithFields(logrus.Fields{
		"userID":       userID,
		"title":        newCategory.Title,
		"portfolio_id": newCategory.PortfolioID,
	}).Info("Creating category - position will be set by database trigger")

	// Create category (position is automatically set by database trigger)
	if err := h.repo.Create(&newCategory); err != nil {
		logrus.WithFields(logrus.Fields{
			"userID":       userID,
			"error":        err.Error(),
			"portfolio_id": newCategory.PortfolioID,
		}).Error("Failed to create category")
		response.InternalError(c, "Failed to create category")
		return
	}

	logrus.WithFields(logrus.Fields{
		"userID":     userID,
		"categoryID": newCategory.ID,
		"title":      newCategory.Title,
	}).Info("Category created successfully")

	response.Created(c, "category", &newCategory, "Category created successfully")
}

func (h *CategoryHandler) Delete(c *gin.Context) {
	userID := c.GetString("userID")
	categoryID := c.Param("id")

	id, err := strconv.Atoi(categoryID)
	if err != nil {
		response.BadRequest(c, "Invalid category ID")
		return
	}

	// Fetch category to check ownership and get portfolio_id
	category, err := h.repo.GetByID(uint(id))
	if err != nil {
		response.NotFound(c, "Category not found")
		return
	}

	if category.OwnerID != userID {
		response.Forbidden(c, "Access denied")
		return
	}

	// Delete category (CASCADE: all related projects will be deleted)
	if err := h.repo.Delete(uint(id)); err != nil {
		logrus.WithFields(logrus.Fields{
			"categoryID":  id,
			"userID":      userID,
			"portfolioID": category.PortfolioID,
			"error":       err.Error(),
		}).Error("Failed to delete category")

		response.InternalError(c, "Failed to delete category")
		return
	}

	logrus.WithFields(logrus.Fields{
		"categoryID":  id,
		"userID":      userID,
		"portfolioID": category.PortfolioID,
		"title":       category.Title,
	}).Info("Category deleted successfully (CASCADE: all related projects)")

	response.OK(c, "message", "Category deleted successfully", "Success")
}

func (h *CategoryHandler) GetByIDPublic(c *gin.Context) {
	categoryID := c.Param("id")

	// Parse category ID
	id, err := strconv.Atoi(categoryID)
	if err != nil {
		response.BadRequest(c, "Invalid category ID")
		return
	}

	// Get complete category with relationships
	category, err := h.repo.GetByIDWithRelations(uint(id))
	if err != nil {
		response.NotFound(c, "Category not found")
		return
	}

	response.OK(c, "category", category, "Success")
}

func (h *CategoryHandler) GetByPortfolio(c *gin.Context) {
	portfolioID := c.Param("id")

	// Get categories for this portfolio
	categories, err := h.repo.GetByPortfolioID(portfolioID)
	if err != nil {
		response.InternalError(c, "Failed to retrieve categories")
		return
	}

	response.OK(c, "categories", categories, "Success")
}

// UpdatePosition updates the position field of a category
func (h *CategoryHandler) UpdatePosition(c *gin.Context) {
	userID := c.GetString("userID") // From auth middleware
	categoryID := c.Param("id")

	// Parse category ID
	id, err := strconv.Atoi(categoryID)
	if err != nil {
		response.BadRequest(c, "Invalid category ID")
		return
	}

	// Parse request body
	var req struct {
		Position uint `json:"position" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request data")
		return
	}

	// Check if category exists and belongs to user
	existing, err := h.repo.GetByIDBasic(uint(id))
	if err != nil {
		response.NotFound(c, "Category not found")
		return
	}

	if existing.OwnerID != userID {
		response.Forbidden(c, "Access denied")
		return
	}

	// Update position
	if err := h.repo.UpdatePosition(uint(id), req.Position); err != nil {
		response.InternalError(c, "Failed to update category position")
		return
	}

	response.OK(c, "message", "Category position updated successfully", "Success")
}
