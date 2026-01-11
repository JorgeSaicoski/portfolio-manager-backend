package controllers

import (
	"net/http"
	"strconv"

	contracts2 "github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/contracts"
	appdto "github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/dto"
	portfolio2 "github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/usecases/portfolio"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/interfaces/dto/request"
	response2 "github.com/JorgeSaicoski/portfolio-manager/backend/internal/interfaces/dto/response"
	pkgerrors "github.com/JorgeSaicoski/portfolio-manager/backend/pkg/errors"
	"github.com/gin-gonic/gin"
)

// PortfolioController handles HTTP requests for portfolio operations
type PortfolioController struct {
	createUseCase    *portfolio2.CreatePortfolioUseCase
	getUseCase       *portfolio2.GetPortfolioUseCase
	getPublicUseCase *portfolio2.GetPortfolioPublicUseCase
	listUseCase      *portfolio2.ListPortfoliosUseCase
	updateUseCase    *portfolio2.UpdatePortfolioUseCase
	deleteUseCase    *portfolio2.DeletePortfolioUseCase
	categoryRepo     contracts2.CategoryRepository
	sectionRepo      contracts2.SectionRepository
}

// NewPortfolioController creates a new portfolio controller instance
func NewPortfolioController(
	createUC *portfolio2.CreatePortfolioUseCase,
	getUC *portfolio2.GetPortfolioUseCase,
	getPublicUC *portfolio2.GetPortfolioPublicUseCase,
	listUC *portfolio2.ListPortfoliosUseCase,
	updateUC *portfolio2.UpdatePortfolioUseCase,
	deleteUC *portfolio2.DeletePortfolioUseCase,
	categoryRepo contracts2.CategoryRepository,
	sectionRepo contracts2.SectionRepository,
) *PortfolioController {
	return &PortfolioController{
		createUseCase:    createUC,
		getUseCase:       getUC,
		getPublicUseCase: getPublicUC,
		listUseCase:      listUC,
		updateUseCase:    updateUC,
		deleteUseCase:    deleteUC,
		categoryRepo:     categoryRepo,
		sectionRepo:      sectionRepo,
	}
}

// Create handles POST /api/v2/portfolios
func (ctrl *PortfolioController) Create(c *gin.Context) {
	// 1. Extract userID from context (set by auth middleware)
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, response2.ErrorResponse{Error: "unauthorized: missing user ID"})
		return
	}

	// 2. Bind and validate HTTP request DTO
	var req request.CreatePortfolioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response2.ErrorResponse{Error: err.Error()})
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
		status := pkgerrors.ToHTTPStatus(err)
		c.JSON(status, response2.ErrorResponse{Error: err.Error()})
		return
	}

	// 5. Map application DTO to HTTP response DTO
	resp := response2.PortfolioResponse{
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
		c.JSON(http.StatusUnauthorized, response2.ErrorResponse{Error: "unauthorized: missing user ID"})
		return
	}

	// 2. Bind and validate query parameters
	var req request.ListPortfoliosRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, response2.ErrorResponse{Error: err.Error()})
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
		status := pkgerrors.ToHTTPStatus(err)
		c.JSON(status, response2.ErrorResponse{Error: err.Error()})
		return
	}

	// 5. Map to HTTP response DTOs
	portfolios := make([]response2.PortfolioResponse, len(output.Portfolios))
	for i, p := range output.Portfolios {
		portfolios[i] = response2.PortfolioResponse{
			ID:          p.ID,
			Title:       p.Title,
			Description: p.Description,
			OwnerID:     p.OwnerID,
			CreatedAt:   p.CreatedAt,
			UpdatedAt:   p.UpdatedAt,
		}
	}

	resp := response2.ListPortfoliosResponse{
		Portfolios: portfolios,
		Pagination: response2.PaginationResponse{
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
		c.JSON(http.StatusUnauthorized, response2.ErrorResponse{Error: "unauthorized: missing user ID"})
		return
	}

	// 2. Parse portfolio ID from URL parameter
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response2.ErrorResponse{Error: "invalid portfolio ID"})
		return
	}

	// 3. Execute use case
	portfolioDTO, err := ctrl.getUseCase.Execute(c.Request.Context(), uint(id))
	if err != nil {
		status := pkgerrors.ToHTTPStatus(err)
		c.JSON(status, response2.ErrorResponse{Error: err.Error()})
		return
	}

	// 4. Authorization check: verify ownership
	if portfolioDTO.OwnerID != userID {
		c.JSON(http.StatusForbidden, response2.ErrorResponse{Error: "forbidden: you don't own this portfolio"})
		return
	}

	// 5. Map to HTTP response DTO
	resp := response2.PortfolioResponse{
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
		c.JSON(http.StatusUnauthorized, response2.ErrorResponse{Error: "unauthorized: missing user ID"})
		return
	}

	// 2. Parse portfolio ID from URL parameter
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response2.ErrorResponse{Error: "invalid portfolio ID"})
		return
	}

	// 3. Bind and validate HTTP request DTO
	var req request.UpdatePortfolioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response2.ErrorResponse{Error: err.Error()})
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
		status := pkgerrors.ToHTTPStatus(err)
		c.JSON(status, response2.ErrorResponse{Error: err.Error()})
		return
	}

	// 6. Return success response
	c.JSON(http.StatusOK, response2.SuccessResponse{Message: "portfolio updated successfully"})
}

// Delete handles DELETE /api/v2/portfolios/:id
func (ctrl *PortfolioController) Delete(c *gin.Context) {
	// 1. Extract userID from context (set by auth middleware)
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, response2.ErrorResponse{Error: "unauthorized: missing user ID"})
		return
	}

	// 2. Parse portfolio ID from URL parameter
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response2.ErrorResponse{Error: "invalid portfolio ID"})
		return
	}

	// 3. Execute use case (use case handles ownership check)
	err = ctrl.deleteUseCase.Execute(c.Request.Context(), uint(id), userID)
	if err != nil {
		status := pkgerrors.ToHTTPStatus(err)
		c.JSON(status, response2.ErrorResponse{Error: err.Error()})
		return
	}

	// 4. Return success response
	c.JSON(http.StatusOK, response2.SuccessResponse{Message: "portfolio deleted successfully"})
}

// GetPublicByID handles GET /api/portfolios/id/:id and GET /api/portfolios/public/:id
func (ctrl *PortfolioController) GetPublicByID(c *gin.Context) {
	// Parse portfolio ID from URL parameter
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response2.ErrorResponse{Error: "invalid portfolio ID"})
		return
	}

	// Execute use case (no auth required for public access)
	portfolioDTO, err := ctrl.getPublicUseCase.Execute(c.Request.Context(), uint(id))
	if err != nil {
		status := pkgerrors.ToHTTPStatus(err)
		c.JSON(status, response2.ErrorResponse{Error: err.Error()})
		return
	}

	// Map to HTTP response DTO (don't include OwnerID in public response)
	resp := response2.PortfolioResponse{
		ID:          portfolioDTO.ID,
		Title:       portfolioDTO.Title,
		Description: portfolioDTO.Description,
		CreatedAt:   portfolioDTO.CreatedAt,
		UpdatedAt:   portfolioDTO.UpdatedAt,
	}

	// Return HTTP response with API_OVERVIEW.md format
	c.JSON(http.StatusOK, response2.DataResponse{
		Data:    resp,
		Message: "Success",
	})
}

// GetPublicCategories handles GET /api/portfolios/public/:id/categories
func (ctrl *PortfolioController) GetPublicCategories(c *gin.Context) {
	// Parse portfolio ID from URL parameter
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response2.ErrorResponse{Error: "invalid portfolio ID"})
		return
	}

	// Verify portfolio exists
	_, err = ctrl.getPublicUseCase.Execute(c.Request.Context(), uint(id))
	if err != nil {
		status := pkgerrors.ToHTTPStatus(err)
		c.JSON(status, response2.ErrorResponse{Error: err.Error()})
		return
	}

	// Get all categories for the portfolio
	categories, err := ctrl.categoryRepo.GetByPortfolioID(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, response2.ErrorResponse{Error: "failed to retrieve categories"})
		return
	}

	// Map to HTTP response DTOs
	categoryResponses := make([]response2.CategoryResponse, len(categories))
	for i, cat := range categories {
		categoryResponses[i] = response2.CategoryResponse{
			ID:          cat.ID,
			Title:       cat.Title,
			Description: cat.Description,
			Position:    cat.Position,
			PortfolioID: cat.PortfolioID,
			CreatedAt:   cat.CreatedAt,
			UpdatedAt:   cat.UpdatedAt,
		}
	}

	// Return HTTP response with API_OVERVIEW.md format
	c.JSON(http.StatusOK, response2.DataResponse{
		Data:    categoryResponses,
		Message: "Success",
	})
}

// GetPublicSections handles GET /api/portfolios/public/:id/sections
func (ctrl *PortfolioController) GetPublicSections(c *gin.Context) {
	// Parse portfolio ID from URL parameter
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response2.ErrorResponse{Error: "invalid portfolio ID"})
		return
	}

	// Verify portfolio exists
	_, err = ctrl.getPublicUseCase.Execute(c.Request.Context(), uint(id))
	if err != nil {
		status := pkgerrors.ToHTTPStatus(err)
		c.JSON(status, response2.ErrorResponse{Error: err.Error()})
		return
	}

	// Get all sections for the portfolio
	sections, err := ctrl.sectionRepo.GetByPortfolioID(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, response2.ErrorResponse{Error: "failed to retrieve sections"})
		return
	}

	// Map to HTTP response DTOs
	sectionResponses := make([]response2.SectionResponse, len(sections))
	for i, sec := range sections {
		sectionResponses[i] = response2.SectionResponse{
			ID:          sec.ID,
			Title:       sec.Title,
			Description: sec.Description,
			Position:    sec.Position,
			Type:        sec.Type,
			PortfolioID: sec.PortfolioID,
			CreatedAt:   sec.CreatedAt,
			UpdatedAt:   sec.UpdatedAt,
		}
	}

	// Return HTTP response with API_OVERVIEW.md format
	c.JSON(http.StatusOK, response2.DataResponse{
		Data:    sectionResponses,
		Message: "Success",
	})
}
