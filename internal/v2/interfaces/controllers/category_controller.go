package controllers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	appdto "github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/dto"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/usecases/category"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/interfaces/dto/request"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/interfaces/dto/response"
)

// CategoryController handles HTTP requests for category operations
type CategoryController struct {
	createUseCase         *category.CreateCategoryUseCase
	getUseCase            *category.GetCategoryUseCase
	getPublicUseCase      *category.GetCategoryPublicUseCase
	listUseCase           *category.ListCategoriesUseCase
	updateUseCase         *category.UpdateCategoryUseCase
	updatePositionUseCase *category.UpdateCategoryPositionUseCase
	bulkReorderUseCase    *category.BulkReorderCategoriesUseCase
	deleteUseCase         *category.DeleteCategoryUseCase
}

// NewCategoryController creates a new category controller instance
func NewCategoryController(
	createUC *category.CreateCategoryUseCase,
	getUC *category.GetCategoryUseCase,
	getPublicUC *category.GetCategoryPublicUseCase,
	listUC *category.ListCategoriesUseCase,
	updateUC *category.UpdateCategoryUseCase,
	updatePositionUC *category.UpdateCategoryPositionUseCase,
	bulkReorderUC *category.BulkReorderCategoriesUseCase,
	deleteUC *category.DeleteCategoryUseCase,
) *CategoryController {
	return &CategoryController{
		createUseCase:         createUC,
		getUseCase:            getUC,
		getPublicUseCase:      getPublicUC,
		listUseCase:           listUC,
		updateUseCase:         updateUC,
		updatePositionUseCase: updatePositionUC,
		bulkReorderUseCase:    bulkReorderUC,
		deleteUseCase:         deleteUC,
	}
}

// Create handles POST /api/categories/own
func (ctrl *CategoryController) Create(c *gin.Context) {
	// Extract userID from context (set by auth middleware)
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Error: "unauthorized: missing user ID"})
		return
	}

	// Bind and validate HTTP request DTO
	var req request.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: err.Error()})
		return
	}

	// Map HTTP request DTO to application DTO
	input := appdto.CreateCategoryInput{
		Title:       req.Title,
		Description: req.Description,
		Position:    req.Position,
		OwnerID:     userID,
		PortfolioID: req.PortfolioID,
	}

	// Execute use case
	categoryDTO, err := ctrl.createUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		status := mapErrorToHTTPStatus(err)
		c.JSON(status, response.ErrorResponse{Error: err.Error()})
		return
	}

	// Map application DTO to HTTP response DTO
	resp := response.CategoryResponse{
		ID:          categoryDTO.ID,
		Title:       categoryDTO.Title,
		Description: categoryDTO.Description,
		Position:    categoryDTO.Position,
		OwnerID:     categoryDTO.OwnerID,
		PortfolioID: categoryDTO.PortfolioID,
		CreatedAt:   categoryDTO.CreatedAt,
		UpdatedAt:   categoryDTO.UpdatedAt,
	}

	// Return HTTP response with API_OVERVIEW.md format
	c.JSON(http.StatusCreated, response.DataResponse{
		Data:    resp,
		Message: "Success",
	})
}

// List handles GET /api/categories/own
func (ctrl *CategoryController) List(c *gin.Context) {
	// Extract userID from context (set by auth middleware)
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Error: "unauthorized: missing user ID"})
		return
	}

	// Bind and validate query parameters
	var req request.ListCategoriesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: err.Error()})
		return
	}

	// Set default pagination values if not provided
	if req.Page == 0 {
		req.Page = 1
	}
	if req.Limit == 0 {
		req.Limit = 10
	}

	// Map to application DTO
	input := appdto.ListCategoriesInput{
		PortfolioID: 0, // List all categories for user
		Pagination: appdto.PaginationDTO{
			Page:  req.Page,
			Limit: req.Limit,
		},
	}

	// Execute use case
	output, err := ctrl.listUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		status := mapErrorToHTTPStatus(err)
		c.JSON(status, response.ErrorResponse{Error: err.Error()})
		return
	}

	// Map to HTTP response DTOs
	categories := make([]response.CategoryResponse, len(output.Categories))
	for i, cat := range output.Categories {
		categories[i] = response.CategoryResponse{
			ID:          cat.ID,
			Title:       cat.Title,
			Description: cat.Description,
			Position:    cat.Position,
			OwnerID:     cat.OwnerID,
			PortfolioID: cat.PortfolioID,
			CreatedAt:   cat.CreatedAt,
			UpdatedAt:   cat.UpdatedAt,
		}
	}

	// Return HTTP response with API_OVERVIEW.md format
	c.JSON(http.StatusOK, response.PaginatedDataResponse{
		Data:    categories,
		Page:    output.Pagination.Page,
		Limit:   output.Pagination.Limit,
		Total:   output.Pagination.Total,
		Message: "Success",
	})
}

// GetByID handles GET /api/categories/own/:id
func (ctrl *CategoryController) GetByID(c *gin.Context) {
	// Extract userID from context (set by auth middleware)
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Error: "unauthorized: missing user ID"})
		return
	}

	// Parse category ID from URL parameter
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: "invalid category ID"})
		return
	}

	// Execute use case
	categoryDTO, err := ctrl.getUseCase.Execute(c.Request.Context(), uint(id), userID)
	if err != nil {
		status := mapErrorToHTTPStatus(err)
		c.JSON(status, response.ErrorResponse{Error: err.Error()})
		return
	}

	// Map to HTTP response DTO
	resp := response.CategoryResponse{
		ID:          categoryDTO.ID,
		Title:       categoryDTO.Title,
		Description: categoryDTO.Description,
		Position:    categoryDTO.Position,
		OwnerID:     categoryDTO.OwnerID,
		PortfolioID: categoryDTO.PortfolioID,
		CreatedAt:   categoryDTO.CreatedAt,
		UpdatedAt:   categoryDTO.UpdatedAt,
	}

	// Return HTTP response with API_OVERVIEW.md format
	c.JSON(http.StatusOK, response.DataResponse{
		Data:    resp,
		Message: "Success",
	})
}

// Update handles PUT /api/categories/own/:id
func (ctrl *CategoryController) Update(c *gin.Context) {
	// Extract userID from context (set by auth middleware)
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Error: "unauthorized: missing user ID"})
		return
	}

	// Parse category ID from URL parameter
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: "invalid category ID"})
		return
	}

	// Bind and validate HTTP request DTO
	var req request.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: err.Error()})
		return
	}

	// Map to application DTO
	input := appdto.UpdateCategoryInput{
		ID:          uint(id),
		Title:       req.Title,
		Description: req.Description,
		Position:    req.Position,
		OwnerID:     userID,
	}

	// Execute use case
	err = ctrl.updateUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		status := mapErrorToHTTPStatus(err)
		c.JSON(status, response.ErrorResponse{Error: err.Error()})
		return
	}

	// Return success response with API_OVERVIEW.md format
	c.JSON(http.StatusOK, response.DataResponse{
		Data:    nil,
		Message: "Category updated successfully",
	})
}

// UpdatePosition handles PUT /api/categories/own/:id/position
func (ctrl *CategoryController) UpdatePosition(c *gin.Context) {
	// Extract userID from context (set by auth middleware)
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Error: "unauthorized: missing user ID"})
		return
	}

	// Parse category ID from URL parameter
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: "invalid category ID"})
		return
	}

	// Bind and validate HTTP request DTO
	var req request.UpdateCategoryPositionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: err.Error()})
		return
	}

	// Execute use case
	err = ctrl.updatePositionUseCase.Execute(c.Request.Context(), uint(id), req.Position, userID)
	if err != nil {
		status := mapErrorToHTTPStatus(err)
		c.JSON(status, response.ErrorResponse{Error: err.Error()})
		return
	}

	// Return success response with API_OVERVIEW.md format
	c.JSON(http.StatusOK, response.DataResponse{
		Data:    nil,
		Message: "Category position updated successfully",
	})
}

// BulkReorder handles PUT /api/categories/own/reorder
func (ctrl *CategoryController) BulkReorder(c *gin.Context) {
	// Extract userID from context (set by auth middleware)
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Error: "unauthorized: missing user ID"})
		return
	}

	// Bind and validate HTTP request DTO
	var req request.BulkReorderCategoriesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: err.Error()})
		return
	}

	// Map to application DTO
	items := make([]appdto.BulkUpdatePositionItem, len(req.Items))
	for i, item := range req.Items {
		items[i] = appdto.BulkUpdatePositionItem{
			ID:       item.ID,
			Position: item.Position,
		}
	}

	input := appdto.BulkUpdateCategoryPositionsInput{
		Items:   items,
		OwnerID: userID,
	}

	// Execute use case
	err := ctrl.bulkReorderUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		status := mapErrorToHTTPStatus(err)
		c.JSON(status, response.ErrorResponse{Error: err.Error()})
		return
	}

	// Return success response with API_OVERVIEW.md format
	c.JSON(http.StatusOK, response.DataResponse{
		Data:    nil,
		Message: "Categories reordered successfully",
	})
}

// Delete handles DELETE /api/categories/own/:id
func (ctrl *CategoryController) Delete(c *gin.Context) {
	// Extract userID from context (set by auth middleware)
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Error: "unauthorized: missing user ID"})
		return
	}

	// Parse category ID from URL parameter
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: "invalid category ID"})
		return
	}

	// Execute use case
	err = ctrl.deleteUseCase.Execute(c.Request.Context(), uint(id), userID)
	if err != nil {
		status := mapErrorToHTTPStatus(err)
		c.JSON(status, response.ErrorResponse{Error: err.Error()})
		return
	}

	// Return success response with API_OVERVIEW.md format
	c.JSON(http.StatusOK, response.DataResponse{
		Data:    nil,
		Message: "Category deleted successfully",
	})
}

// GetPublicByID handles GET /api/categories/id/:id and GET /api/categories/public/:id
func (ctrl *CategoryController) GetPublicByID(c *gin.Context) {
	// Parse category ID from URL parameter
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: "invalid category ID"})
		return
	}

	// Execute use case (no auth required for public access)
	categoryDTO, err := ctrl.getPublicUseCase.Execute(c.Request.Context(), uint(id))
	if err != nil {
		status := mapErrorToHTTPStatus(err)
		c.JSON(status, response.ErrorResponse{Error: err.Error()})
		return
	}

	// Map to HTTP response DTO (don't include OwnerID in public response)
	resp := response.CategoryResponse{
		ID:          categoryDTO.ID,
		Title:       categoryDTO.Title,
		Description: categoryDTO.Description,
		Position:    categoryDTO.Position,
		PortfolioID: categoryDTO.PortfolioID,
		CreatedAt:   categoryDTO.CreatedAt,
		UpdatedAt:   categoryDTO.UpdatedAt,
	}

	// Return HTTP response with API_OVERVIEW.md format
	c.JSON(http.StatusOK, response.DataResponse{
		Data:    resp,
		Message: "Success",
	})
}

// GetPublicProjects handles GET /api/categories/public/:id/projects
// TODO: Implement when Project use cases are created
func (ctrl *CategoryController) GetPublicProjects(c *gin.Context) {
	// Parse category ID from URL parameter
	idStr := c.Param("id")
	_, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: "invalid category ID"})
		return
	}

	// TODO: Implement projects retrieval when Project domain is ready
	c.JSON(http.StatusNotImplemented, response.ErrorResponse{Error: "not implemented yet"})
}

// mapErrorToHTTPStatus maps domain errors to appropriate HTTP status codes
func mapErrorToHTTPStatus(err error) int {
	errMsg := strings.ToLower(err.Error())

	switch {
	case strings.Contains(errMsg, "not found"):
		return http.StatusNotFound
	case strings.Contains(errMsg, "unauthorized"):
		return http.StatusForbidden
	case strings.Contains(errMsg, "already exists"):
		return http.StatusConflict
	case strings.Contains(errMsg, "required"):
		return http.StatusBadRequest
	case strings.Contains(errMsg, "invalid"):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
