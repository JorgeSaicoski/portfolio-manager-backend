package controllers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/dto"
	portfolio2 "github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/usecases/portfolio"
	interfaceDTO "github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/interfaces/dto"
	pkgErrors "github.com/JorgeSaicoski/portfolio-manager/backend/pkg/errors"
	"github.com/gin-gonic/gin"
)

// PortfolioController handles HTTP requests for portfolio operations
// This is a thin layer that delegates to use cases
type PortfolioController struct {
	createUC *portfolio2.CreatePortfolioUseCase
	getUC    *portfolio2.GetPortfolioUseCase
	listUC   *portfolio2.ListPortfoliosUseCase
	updateUC *portfolio2.UpdatePortfolioUseCase
	deleteUC *portfolio2.DeletePortfolioUseCase
}

// NewPortfolioController creates a new portfolio controller
func NewPortfolioController(
	createUC *portfolio2.CreatePortfolioUseCase,
	getUC *portfolio2.GetPortfolioUseCase,
	listUC *portfolio2.ListPortfoliosUseCase,
	updateUC *portfolio2.UpdatePortfolioUseCase,
	deleteUC *portfolio2.DeletePortfolioUseCase,
) *PortfolioController {
	return &PortfolioController{
		createUC: createUC,
		getUC:    getUC,
		listUC:   listUC,
		updateUC: updateUC,
		deleteUC: deleteUC,
	}
}

// Create handles POST /api/portfolios/own
func (ctrl *PortfolioController) Create(c *gin.Context) {
	// 1. Parse HTTP request
	var req interfaceDTO.CreatePortfolioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, interfaceDTO.ErrorResponse{
			Error: "Invalid request body",
			Details: map[string]interface{}{
				"validation_error": err.Error(),
			},
		})
		return
	}

	// 2. Get user ID from context (set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, interfaceDTO.ErrorResponse{
			Error: "Unauthorized",
		})
		return
	}

	// 3. Convert HTTP DTO to application input
	input := dto.CreatePortfolioInput{
		Title:       req.Title,
		Description: req.Description,
		OwnerID:     userID.(string),
	}

	// 4. Execute use case
	result, err := ctrl.createUC.Execute(c.Request.Context(), input)
	if err != nil {
		ctrl.handleError(c, err)
		return
	}

	// 5. Convert application DTO to HTTP response
	response := interfaceDTO.PortfolioResponse{
		ID:          result.ID,
		Title:       result.Title,
		Description: result.Description,
		OwnerID:     result.OwnerID,
		CreatedAt:   result.CreatedAt,
		UpdatedAt:   result.UpdatedAt,
	}

	c.JSON(http.StatusCreated, interfaceDTO.SuccessResponse{
		Message: "Portfolio created successfully",
		Data:    response,
	})
}

// GetByID handles GET /api/portfolios/public/:id
func (ctrl *PortfolioController) GetByID(c *gin.Context) {
	// 1. Parse ID from URL
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, interfaceDTO.ErrorResponse{
			Error: "Invalid portfolio ID",
		})
		return
	}

	// 2. Execute use case
	result, err := ctrl.getUC.Execute(c.Request.Context(), uint(id))
	if err != nil {
		ctrl.handleError(c, err)
		return
	}

	// 3. Convert application DTO to HTTP response
	response := interfaceDTO.PortfolioResponse{
		ID:          result.ID,
		Title:       result.Title,
		Description: result.Description,
		CreatedAt:   result.CreatedAt,
		UpdatedAt:   result.UpdatedAt,
	}

	c.JSON(http.StatusOK, interfaceDTO.SuccessResponse{
		Message: "Portfolio retrieved successfully",
		Data:    response,
	})
}

// ListByUser handles GET /api/portfolios/own
func (ctrl *PortfolioController) ListByUser(c *gin.Context) {
	// 1. Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, interfaceDTO.ErrorResponse{
			Error: "Unauthorized",
		})
		return
	}

	// 2. Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// 3. Build use case input
	input := dto.ListPortfoliosInput{
		OwnerID: userID.(string),
		Pagination: dto.PaginationDTO{
			Page:  page,
			Limit: limit,
		},
	}

	// 4. Execute use case
	result, err := ctrl.listUC.Execute(c.Request.Context(), input)
	if err != nil {
		ctrl.handleError(c, err)
		return
	}

	// 5. Convert application DTOs to HTTP response
	portfolios := make([]interfaceDTO.PortfolioResponse, len(result.Portfolios))
	for i, p := range result.Portfolios {
		portfolios[i] = interfaceDTO.PortfolioResponse{
			ID:          p.ID,
			Title:       p.Title,
			Description: p.Description,
			OwnerID:     p.OwnerID,
			CreatedAt:   p.CreatedAt,
			UpdatedAt:   p.UpdatedAt,
		}
	}

	response := interfaceDTO.ListPortfoliosResponse{
		Data:  portfolios,
		Total: result.Pagination.Total,
		Page:  result.Pagination.Page,
		Limit: result.Pagination.Limit,
	}

	c.JSON(http.StatusOK, response)
}

// Update handles PUT /api/portfolios/own/:id
func (ctrl *PortfolioController) Update(c *gin.Context) {
	// 1. Parse ID from URL
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, interfaceDTO.ErrorResponse{
			Error: "Invalid portfolio ID",
		})
		return
	}

	// 2. Parse HTTP request
	var req interfaceDTO.UpdatePortfolioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, interfaceDTO.ErrorResponse{
			Error: "Invalid request body",
			Details: map[string]interface{}{
				"validation_error": err.Error(),
			},
		})
		return
	}

	// 3. Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, interfaceDTO.ErrorResponse{
			Error: "Unauthorized",
		})
		return
	}

	// 4. Convert HTTP DTO to application input
	input := dto.UpdatePortfolioInput{
		ID:          uint(id),
		Title:       req.Title,
		Description: req.Description,
		OwnerID:     userID.(string),
	}

	// 5. Execute use case
	if err := ctrl.updateUC.Execute(c.Request.Context(), input); err != nil {
		ctrl.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, interfaceDTO.SuccessResponse{
		Message: "Portfolio updated successfully",
	})
}

// Delete handles DELETE /api/portfolios/own/:id
func (ctrl *PortfolioController) Delete(c *gin.Context) {
	// 1. Parse ID from URL
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, interfaceDTO.ErrorResponse{
			Error: "Invalid portfolio ID",
		})
		return
	}

	// 2. Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, interfaceDTO.ErrorResponse{
			Error: "Unauthorized",
		})
		return
	}

	// 3. Execute use case
	if err := ctrl.deleteUC.Execute(c.Request.Context(), uint(id), userID.(string)); err != nil {
		ctrl.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, interfaceDTO.SuccessResponse{
		Message: "Portfolio deleted successfully",
	})
}

// handleError converts application/domain errors to HTTP responses
func (ctrl *PortfolioController) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, pkgErrors.ErrNotFound):
		c.JSON(http.StatusNotFound, interfaceDTO.ErrorResponse{
			Error: "Resource not found",
		})
	case errors.Is(err, pkgErrors.ErrUnauthorized):
		c.JSON(http.StatusForbidden, interfaceDTO.ErrorResponse{
			Error: "You don't have permission to access this resource",
		})
	case errors.Is(err, pkgErrors.ErrDuplicate):
		c.JSON(http.StatusConflict, interfaceDTO.ErrorResponse{
			Error: err.Error(),
		})
	case errors.Is(err, pkgErrors.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, interfaceDTO.ErrorResponse{
			Error: err.Error(),
		})
	default:
		c.JSON(http.StatusInternalServerError, interfaceDTO.ErrorResponse{
			Error: "Internal server error",
		})
	}
}
