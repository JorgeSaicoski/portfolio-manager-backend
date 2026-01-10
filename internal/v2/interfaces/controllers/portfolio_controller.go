package controllers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	appdto "github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/dto"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/usecases/portfolio"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/interfaces/dto/request"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/interfaces/dto/response"
)

// PortfolioController handles HTTP requests for portfolio operations
type PortfolioController struct {
	createUseCase *portfolio.CreatePortfolioUseCase
	getUseCase    *portfolio.GetPortfolioUseCase
	listUseCase   *portfolio.ListPortfoliosUseCase
	updateUseCase *portfolio.UpdatePortfolioUseCase
	deleteUseCase *portfolio.DeletePortfolioUseCase
}

// NewPortfolioController creates a new portfolio controller instance
func NewPortfolioController(
	createUC *portfolio.CreatePortfolioUseCase,
	getUC *portfolio.GetPortfolioUseCase,
	listUC *portfolio.ListPortfoliosUseCase,
	updateUC *portfolio.UpdatePortfolioUseCase,
	deleteUC *portfolio.DeletePortfolioUseCase,
) *PortfolioController {
	return &PortfolioController{
		createUseCase: createUC,
		getUseCase:    getUC,
		listUseCase:   listUC,
		updateUseCase: updateUC,
		deleteUseCase: deleteUC,
	}
}

// Create handles POST /api/v2/portfolios
func (ctrl *PortfolioController) Create(c *gin.Context) {
	// 1. Extract userID from context (set by auth middleware)
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Error: "unauthorized: missing user ID"})
		return
	}

	// 2. Bind and validate HTTP request DTO
	var req request.CreatePortfolioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: err.Error()})
		return
	}

	// 3. Map HTTP request DTO to application DTO
	input := appdto.CreatePortfolioInput{
		Title:       req.Title,
		Description: req.Description,
		OwnerID:     userID,
	}

	// 4. Execute use case
	portfolioDTO, err := ctrl.createUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		status := mapErrorToHTTPStatus(err)
		c.JSON(status, response.ErrorResponse{Error: err.Error()})
		return
	}

	// 5. Map application DTO to HTTP response DTO
	resp := response.PortfolioResponse{
		ID:          portfolioDTO.ID,
		Title:       portfolioDTO.Title,
		Description: portfolioDTO.Description,
		OwnerID:     portfolioDTO.OwnerID,
		CreatedAt:   portfolioDTO.CreatedAt,
		UpdatedAt:   portfolioDTO.UpdatedAt,
	}

	// 6. Return HTTP response
	c.JSON(http.StatusCreated, resp)
}

// List handles GET /api/v2/portfolios
func (ctrl *PortfolioController) List(c *gin.Context) {
	// 1. Extract userID from context (set by auth middleware)
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Error: "unauthorized: missing user ID"})
		return
	}

	// 2. Bind and validate query parameters
	var req request.ListPortfoliosRequest
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

	// 3. Map to application DTO
	input := appdto.ListPortfoliosInput{
		OwnerID: userID,
		Pagination: appdto.PaginationDTO{
			Page:  req.Page,
			Limit: req.Limit,
		},
	}

	// 4. Execute use case
	output, err := ctrl.listUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		status := mapErrorToHTTPStatus(err)
		c.JSON(status, response.ErrorResponse{Error: err.Error()})
		return
	}

	// 5. Map to HTTP response DTOs
	portfolios := make([]response.PortfolioResponse, len(output.Portfolios))
	for i, p := range output.Portfolios {
		portfolios[i] = response.PortfolioResponse{
			ID:          p.ID,
			Title:       p.Title,
			Description: p.Description,
			OwnerID:     p.OwnerID,
			CreatedAt:   p.CreatedAt,
			UpdatedAt:   p.UpdatedAt,
		}
	}

	resp := response.ListPortfoliosResponse{
		Portfolios: portfolios,
		Pagination: response.PaginationResponse{
			Total: output.Pagination.Total,
			Page:  output.Pagination.Page,
			Limit: output.Pagination.Limit,
		},
	}

	// 6. Return HTTP response
	c.JSON(http.StatusOK, resp)
}

// GetByID handles GET /api/v2/portfolios/:id
func (ctrl *PortfolioController) GetByID(c *gin.Context) {
	// 1. Extract userID from context (set by auth middleware)
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Error: "unauthorized: missing user ID"})
		return
	}

	// 2. Parse portfolio ID from URL parameter
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: "invalid portfolio ID"})
		return
	}

	// 3. Execute use case
	portfolioDTO, err := ctrl.getUseCase.Execute(c.Request.Context(), uint(id))
	if err != nil {
		status := mapErrorToHTTPStatus(err)
		c.JSON(status, response.ErrorResponse{Error: err.Error()})
		return
	}

	// 4. Authorization check: verify ownership
	if portfolioDTO.OwnerID != userID {
		c.JSON(http.StatusForbidden, response.ErrorResponse{Error: "forbidden: you don't own this portfolio"})
		return
	}

	// 5. Map to HTTP response DTO
	resp := response.PortfolioResponse{
		ID:          portfolioDTO.ID,
		Title:       portfolioDTO.Title,
		Description: portfolioDTO.Description,
		OwnerID:     portfolioDTO.OwnerID,
		CreatedAt:   portfolioDTO.CreatedAt,
		UpdatedAt:   portfolioDTO.UpdatedAt,
	}

	// 6. Return HTTP response
	c.JSON(http.StatusOK, resp)
}

// Update handles PUT /api/v2/portfolios/:id
func (ctrl *PortfolioController) Update(c *gin.Context) {
	// 1. Extract userID from context (set by auth middleware)
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Error: "unauthorized: missing user ID"})
		return
	}

	// 2. Parse portfolio ID from URL parameter
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: "invalid portfolio ID"})
		return
	}

	// 3. Bind and validate HTTP request DTO
	var req request.UpdatePortfolioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: err.Error()})
		return
	}

	// 4. Map to application DTO
	input := appdto.UpdatePortfolioInput{
		ID:          uint(id),
		Title:       req.Title,
		Description: req.Description,
		OwnerID:     userID, // For authorization check in use case
	}

	// 5. Execute use case (use case handles ownership check)
	err = ctrl.updateUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		status := mapErrorToHTTPStatus(err)
		c.JSON(status, response.ErrorResponse{Error: err.Error()})
		return
	}

	// 6. Return success response
	c.JSON(http.StatusOK, response.SuccessResponse{Message: "portfolio updated successfully"})
}

// Delete handles DELETE /api/v2/portfolios/:id
func (ctrl *PortfolioController) Delete(c *gin.Context) {
	// 1. Extract userID from context (set by auth middleware)
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Error: "unauthorized: missing user ID"})
		return
	}

	// 2. Parse portfolio ID from URL parameter
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: "invalid portfolio ID"})
		return
	}

	// 3. Execute use case (use case handles ownership check)
	err = ctrl.deleteUseCase.Execute(c.Request.Context(), uint(id), userID)
	if err != nil {
		status := mapErrorToHTTPStatus(err)
		c.JSON(status, response.ErrorResponse{Error: err.Error()})
		return
	}

	// 4. Return success response
	c.JSON(http.StatusOK, response.SuccessResponse{Message: "portfolio deleted successfully"})
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
