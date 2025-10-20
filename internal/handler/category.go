package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/metrics"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/models"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/repo"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/response"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/validator"
	"github.com/gin-gonic/gin"
)

type CategoryHandler struct {
	repo    repo.CategoryRepository
	metrics *metrics.Collector
}

func NewCategoryHandler(repo repo.CategoryRepository, metrics *metrics.Collector) *CategoryHandler {
	return &CategoryHandler{
		repo:    repo,
		metrics: metrics,
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

	// Parse request body
	var newCategory models.Category
	if err := c.ShouldBindJSON(&newCategory); err != nil {
		response.BadRequest(c, "Invalid request data")
		return
	}

	// Set the owner
	newCategory.OwnerID = userID

	// Validate category data
	if err := validator.ValidateCategory(&newCategory); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Create category
	if err := h.repo.Create(&newCategory); err != nil {
		// Check if error is due to foreign key constraint (invalid portfolio_id)
		errMsg := err.Error()
		if strings.Contains(errMsg, "fk_portfolios_categories") || strings.Contains(errMsg, "23503") {
			response.NotFound(c, "Portfolio not found")
			return
		}
		response.InternalError(c, "Failed to create category")
		return
	}

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

	// Use basic method - only fetch id and owner_id for authorization
	category, err := h.repo.GetByIDBasic(uint(id))
	if err != nil {
		response.NotFound(c, "Category not found")
		return
	}

	if category.OwnerID != userID {
		response.Forbidden(c, "Access denied")
		return
	}

	if err := h.repo.Delete(uint(id)); err != nil {
		response.InternalError(c, "Failed to delete category")
		return
	}

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
