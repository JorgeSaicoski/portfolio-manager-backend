package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/models"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/infrastructure/audit"
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

type BulkReorderRequest struct {
	Items []struct {
		ID       uint `json:"id" binding:"required"`
		Position uint `json:"position" binding:"required,min=1"`
	} `json:"items" binding:"required,min=1"`
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
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "GET_CATEGORIES_BY_USER_DB_ERROR",
			"where":     "backend/internal/application/handler/category.go",
			"function":  "GetByUser",
			"userID":    userID,
			"error":     err.Error(),
		}).Error("Failed to retrieve categories")
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
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":  "UPDATE_CATEGORY_INVALID_ID",
			"where":      "backend/internal/application/handler/category.go",
			"function":   "Update",
			"userID":     userID,
			"categoryID": categoryID,
			"error":      err.Error(),
		}).Warn("Invalid category ID")
		response.BadRequest(c, "Invalid category ID")
		return
	}

	// Parse request body
	var updateData models.Category
	if err := c.ShouldBindJSON(&updateData); err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":  "UPDATE_CATEGORY_BAD_REQUEST",
			"where":      "backend/internal/application/handler/category.go",
			"function":   "Update",
			"userID":     userID,
			"categoryID": id,
			"error":      err.Error(),
		}).Warn("Invalid request data")
		response.BadRequest(c, "Invalid request data")
		return
	}

	// Set the ID and owner
	updateData.ID = uint(id)
	updateData.OwnerID = userID

	// Validate category data
	if err := validator.ValidateCategory(&updateData); err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":   "UPDATE_CATEGORY_VALIDATION_ERROR",
			"where":       "backend/internal/application/handler/category.go",
			"function":    "Update",
			"userID":      userID,
			"categoryID":  id,
			"portfolioID": updateData.PortfolioID,
			"error":       err.Error(),
		}).Warn("Category validation failed")
		response.BadRequest(c, err.Error())
		return
	}

	// Check if category exists and belongs to user
	existing, err := h.repo.GetByIDBasic(uint(id))
	if err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":  "UPDATE_CATEGORY_NOT_FOUND",
			"where":      "backend/internal/application/handler/category.go",
			"function":   "Update",
			"userID":     userID,
			"categoryID": id,
			"error":      err.Error(),
		}).Warn("Category not found")
		response.NotFound(c, "Category not found")
		return
	}

	if existing.OwnerID != userID {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":  "UPDATE_CATEGORY_FORBIDDEN",
			"where":      "backend/internal/application/handler/category.go",
			"function":   "Update",
			"userID":     userID,
			"categoryID": id,
			"ownerID":    existing.OwnerID,
		}).Warn("Access denied")
		response.ForbiddenWithDetails(c, "Access denied", map[string]interface{}{
			"resource_type": "category",
			"resource_id":   existing.ID,
			"owner_id":      existing.OwnerID,
			"action":        "update",
		})
		return
	}

	// Update category
	if err := h.repo.Update(&updateData); err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":   "UPDATE_CATEGORY_DB_ERROR",
			"where":       "backend/internal/application/handler/category.go",
			"function":    "Update",
			"userID":      userID,
			"categoryID":  id,
			"portfolioID": updateData.PortfolioID,
			"error":       err.Error(),
		}).Error("Failed to update category")
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
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "CREATE_CATEGORY_BAD_REQUEST",
			"where":     "backend/internal/application/handler/category.go",
			"function":  "Create",
			"userID":    userID,
			"error":     err.Error(),
		}).Warn("Failed to parse category creation request")
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
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":   "CREATE_CATEGORY_VALIDATION_ERROR",
			"where":       "backend/internal/application/handler/category.go",
			"function":    "Create",
			"userID":      userID,
			"portfolioID": newCategory.PortfolioID,
			"error":       err.Error(),
		}).Warn("Category validation failed")
		response.BadRequest(c, err.Error())
		return
	}

	// Validate that the portfolio exists and belongs to the user
	portfolio, err := h.portfolioRepo.GetByIDBasic(newCategory.PortfolioID)
	if err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":   "CREATE_CATEGORY_PORTFOLIO_NOT_FOUND",
			"where":       "backend/internal/application/handler/category.go",
			"function":    "Create",
			"userID":      userID,
			"portfolioID": newCategory.PortfolioID,
			"error":       err.Error(),
		}).Warn("Portfolio not found")
		response.NotFound(c, "Portfolio not found")
		return
	}

	if portfolio.OwnerID != userID {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":   "CREATE_CATEGORY_FORBIDDEN",
			"where":       "backend/internal/application/handler/category.go",
			"function":    "Create",
			"userID":      userID,
			"portfolioID": newCategory.PortfolioID,
			"ownerID":     portfolio.OwnerID,
		}).Warn("Access denied to portfolio")
		response.ForbiddenWithDetails(c, "Access denied", map[string]interface{}{
			"resource_type": "portfolio",
			"resource_id":   portfolio.ID,
			"owner_id":      portfolio.OwnerID,
			"action":        "create_category",
		})
		return
	}

	logrus.WithFields(logrus.Fields{
		"userID":       userID,
		"title":        newCategory.Title,
		"portfolio_id": newCategory.PortfolioID,
	}).Info("Creating category - position will be set by database trigger")

	// Create category (position is automatically set by database trigger)
	if err := h.repo.Create(&newCategory); err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":   "CREATE_CATEGORY_DB_ERROR",
			"where":       "backend/internal/application/handler/category.go",
			"function":    "Create",
			"userID":      userID,
			"portfolioID": newCategory.PortfolioID,
			"error":       err.Error(),
		}).Error("Failed to create category")
		response.InternalError(c, "Failed to create category")
		return
	}

	audit.GetCreateLogger().WithFields(logrus.Fields{
		"operation":   "CREATE_CATEGORY",
		"userID":      userID,
		"categoryID":  newCategory.ID,
		"title":       newCategory.Title,
		"portfolioID": newCategory.PortfolioID,
	}).Info("Category created successfully")

	response.Created(c, "category", &newCategory, "Category created successfully")
}

func (h *CategoryHandler) Delete(c *gin.Context) {
	userID := c.GetString("userID")
	categoryID := c.Param("id")

	id, err := strconv.Atoi(categoryID)
	if err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":  "DELETE_CATEGORY_INVALID_ID",
			"where":      "backend/internal/application/handler/category.go",
			"function":   "Delete",
			"userID":     userID,
			"categoryID": categoryID,
			"error":      err.Error(),
		}).Warn("Invalid category ID")
		response.BadRequest(c, "Invalid category ID")
		return
	}

	// Fetch category to check ownership and get portfolio_id
	category, err := h.repo.GetByID(uint(id))
	if err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":  "DELETE_CATEGORY_NOT_FOUND",
			"where":      "backend/internal/application/handler/category.go",
			"function":   "Delete",
			"userID":     userID,
			"categoryID": id,
			"error":      err.Error(),
		}).Warn("Category not found")
		response.NotFound(c, "Category not found")
		return
	}

	if category.OwnerID != userID {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":  "DELETE_CATEGORY_FORBIDDEN",
			"where":      "backend/internal/application/handler/category.go",
			"function":   "Delete",
			"userID":     userID,
			"categoryID": id,
			"ownerID":    category.OwnerID,
		}).Warn("Access denied")
		response.ForbiddenWithDetails(c, "Access denied", map[string]interface{}{
			"resource_type": "category",
			"resource_id":   category.ID,
			"owner_id":      category.OwnerID,
			"action":        "delete",
		})
		return
	}

	// Delete category (CASCADE: all related projects will be deleted)
	if err := h.repo.Delete(uint(id)); err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":   "DELETE_CATEGORY_DB_ERROR",
			"where":       "backend/internal/application/handler/category.go",
			"function":    "Delete",
			"categoryID":  id,
			"userID":      userID,
			"portfolioID": category.PortfolioID,
			"error":       err.Error(),
		}).Error("Failed to delete category")

		response.InternalError(c, "Failed to delete category")
		return
	}

	audit.GetDeleteLogger().WithFields(logrus.Fields{
		"operation":   "DELETE_CATEGORY",
		"categoryID":  id,
		"userID":      userID,
		"portfolioID": category.PortfolioID,
		"title":       category.Title,
		"cascade":     "projects",
	}).Info("Category deleted successfully (CASCADE: all related projects)")

	response.OK(c, "message", "Category deleted successfully", "Success")
}

func (h *CategoryHandler) GetByIDPublic(c *gin.Context) {
	categoryID := c.Param("id")

	// Parse category ID
	id, err := strconv.Atoi(categoryID)
	if err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":  "GET_CATEGORY_BY_ID_PUBLIC_INVALID_ID",
			"where":      "backend/internal/application/handler/category.go",
			"function":   "GetByIDPublic",
			"categoryID": categoryID,
			"error":      err.Error(),
		}).Warn("Invalid category ID")
		response.BadRequest(c, "Invalid category ID")
		return
	}

	// Get complete category with relationships
	category, err := h.repo.GetByIDWithRelations(uint(id))
	if err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":  "GET_CATEGORY_BY_ID_PUBLIC_NOT_FOUND",
			"where":      "backend/internal/application/handler/category.go",
			"function":   "GetByIDPublic",
			"categoryID": id,
			"error":      err.Error(),
		}).Warn("Category not found")
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
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":   "GET_CATEGORIES_BY_PORTFOLIO_DB_ERROR",
			"where":       "backend/internal/application/handler/category.go",
			"function":    "GetByPortfolio",
			"portfolioID": portfolioID,
			"error":       err.Error(),
		}).Error("Failed to retrieve categories")
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
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":  "UPDATE_CATEGORY_POSITION_INVALID_ID",
			"where":      "backend/internal/application/handler/category.go",
			"function":   "UpdatePosition",
			"userID":     userID,
			"categoryID": categoryID,
			"error":      err.Error(),
		}).Warn("Invalid category ID")
		response.BadRequest(c, "Invalid category ID")
		return
	}

	// Parse request body
	var req struct {
		Position uint `json:"position" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":  "UPDATE_CATEGORY_POSITION_BAD_REQUEST",
			"where":      "backend/internal/application/handler/category.go",
			"function":   "UpdatePosition",
			"userID":     userID,
			"categoryID": id,
			"error":      err.Error(),
		}).Warn("Invalid request data")
		response.BadRequest(c, "Invalid request data")
		return
	}

	// Check if category exists and belongs to user
	existing, err := h.repo.GetByIDBasic(uint(id))
	if err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":  "UPDATE_CATEGORY_POSITION_NOT_FOUND",
			"where":      "backend/internal/application/handler/category.go",
			"function":   "UpdatePosition",
			"userID":     userID,
			"categoryID": id,
			"error":      err.Error(),
		}).Warn("Category not found")
		response.NotFound(c, "Category not found")
		return
	}

	if existing.OwnerID != userID {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":  "UPDATE_CATEGORY_POSITION_FORBIDDEN",
			"where":      "backend/internal/application/handler/category.go",
			"function":   "UpdatePosition",
			"userID":     userID,
			"categoryID": id,
			"ownerID":    existing.OwnerID,
		}).Warn("Access denied")
		response.ForbiddenWithDetails(c, "Access denied", map[string]interface{}{
			"resource_type": "category",
			"resource_id":   existing.ID,
			"owner_id":      existing.OwnerID,
			"action":        "update_position",
		})
		return
	}

	// Store old position before update
	oldPosition := existing.Position

	// Update position
	if err := h.repo.UpdatePosition(uint(id), req.Position); err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":  "UPDATE_CATEGORY_POSITION_DB_ERROR",
			"where":      "backend/internal/application/handler/category.go",
			"function":   "UpdatePosition",
			"userID":     userID,
			"categoryID": id,
			"position":   req.Position,
			"error":      err.Error(),
		}).Error("Failed to update category position")
		response.InternalError(c, "Failed to update category position")
		return
	}

	// Audit log successful update
	audit.GetUpdateLogger().WithFields(logrus.Fields{
		"operation":   "UPDATE_CATEGORY_POSITION",
		"categoryID":  id,
		"oldPosition": oldPosition,
		"newPosition": req.Position,
		"userID":      userID,
	}).Info("Category position updated successfully")

	response.OK(c, "message", "Category position updated successfully", "Success")
}

// BulkReorder handles reordering multiple categories atomically
func (h *CategoryHandler) BulkReorder(c *gin.Context) {
	// Get user ID
	userID := c.GetString("userID") // From auth middleware

	// Parse request
	var req BulkReorderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	// Validate no duplicate positions
	positionMap := make(map[uint]bool)
	for _, item := range req.Items {
		if positionMap[item.Position] {
			response.BadRequest(c, fmt.Sprintf("Duplicate position: %d", item.Position))
			return
		}
		positionMap[item.Position] = true
	}

	// Verify all categories belong to user's portfolios
	categoryIDs := make([]uint, len(req.Items))
	for i, item := range req.Items {
		categoryIDs[i] = item.ID
	}

	categories, err := h.repo.GetByIDs(categoryIDs)
	if err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "BULK_REORDER_CATEGORIES",
			"userID":    userID,
			"error":     err.Error(),
		}).Error("Failed to fetch categories for bulk reorder")
		response.InternalError(c, "Failed to fetch categories")
		return
	}

	if len(categories) != len(req.Items) {
		response.NotFound(c, "Some categories not found")
		return
	}

	// Verify ownership
	for _, cat := range categories {
		portfolio, err := h.portfolioRepo.GetByID(cat.PortfolioID)
		if err != nil || portfolio.OwnerID != userID {
			response.ForbiddenWithDetails(c, "Access denied", map[string]interface{}{
				"resource_type": "category",
				"resource_id":   cat.ID,
			})
			return
		}
	}

	// Update positions in transaction
	if err := h.repo.BulkUpdatePositions(req.Items); err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "BULK_REORDER_CATEGORIES",
			"userID":    userID,
			"itemCount": len(req.Items),
			"error":     err.Error(),
		}).Error("Failed to bulk update category positions")
		response.InternalError(c, "Failed to update positions")
		return
	}

	// Audit log successful reorder
	itemDetails := make([]map[string]uint, len(req.Items))
	for i, item := range req.Items {
		itemDetails[i] = map[string]uint{
			"id":       item.ID,
			"position": item.Position,
		}
	}

	audit.GetUpdateLogger().WithFields(logrus.Fields{
		"operation": "BULK_REORDER_CATEGORIES",
		"userID":    userID,
		"itemCount": len(req.Items),
		"items":     itemDetails,
	}).Info("Categories reordered successfully")

	response.OK(c, "message", "Categories reordered successfully", "Success")
}
