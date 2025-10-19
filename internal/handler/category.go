package handler

import (
	"net/http"
	"strconv"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/metrics"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/models"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/repo"
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
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve categories",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"categories": categories,
		"page":       page,
		"limit":      limit,
		"message":    "Success",
	})
}
func (h *CategoryHandler) Update(c *gin.Context) {
	userID := c.GetString("userID") // From auth middleware
	categoryID := c.Param("id")

	// Parse category ID
	id, err := strconv.Atoi(categoryID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid category ID",
		})
		return
	}

	// Parse request body
	var updateData models.Category
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request data",
		})
		return
	}

	// Set the ID and owner
	updateData.ID = uint(id)
	updateData.OwnerID = userID

	// Update category
	if err := h.repo.Update(&updateData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update category",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"category": &updateData,
		"message":  "Category updated successfully",
	})
}

func (h *CategoryHandler) Create(c *gin.Context) {
	userID := c.GetString("userID") // From auth middleware

	// Parse request body
	var newCategory models.Category
	if err := c.ShouldBindJSON(&newCategory); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request data",
		})
		return
	}

	// Set the owner
	newCategory.OwnerID = userID

	// Create category
	if err := h.repo.Create(&newCategory); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create category",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"category": &newCategory,
		"message":  "Category created successfully",
	})
}

func (h *CategoryHandler) Delete(c *gin.Context) {
	userID := c.GetString("userID")
	categoryID := c.Param("id")

	id, err := strconv.Atoi(categoryID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid category ID",
		})
		return
	}

	// Use basic method - only fetch id and owner_id for authorization
	category, err := h.repo.GetByIDBasic(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Category not found",
		})
		return
	}

	if category.OwnerID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Access denied",
		})
		return
	}

	if err := h.repo.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete category",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Category deleted successfully",
	})
}

func (h *CategoryHandler) GetByIDPublic(c *gin.Context) {
	categoryID := c.Param("id")

	// Parse category ID
	id, err := strconv.Atoi(categoryID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid category ID",
		})
		return
	}

	// Get complete category with relationships
	category, err := h.repo.GetByIDWithRelations(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Category not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"category": category,
		"message":  "Success",
	})
}
