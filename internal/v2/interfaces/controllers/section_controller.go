package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	appdto "github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/dto"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/usecases/section"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/interfaces/dto/request"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/interfaces/dto/response"
	pkgerrors "github.com/JorgeSaicoski/portfolio-manager/backend/pkg/errors"
)

// SectionController handles HTTP requests for section operations
type SectionController struct {
	createUseCase         *section.CreateSectionUseCase
	getUseCase            *section.GetSectionUseCase
	getPublicUseCase      *section.GetSectionPublicUseCase
	listUseCase           *section.ListSectionsUseCase
	updateUseCase         *section.UpdateSectionUseCase
	updatePositionUseCase *section.UpdateSectionPositionUseCase
	bulkReorderUseCase    *section.BulkReorderSectionsUseCase
	deleteUseCase         *section.DeleteSectionUseCase
}

// NewSectionController creates a new section controller instance
func NewSectionController(
	createUC *section.CreateSectionUseCase,
	getUC *section.GetSectionUseCase,
	getPublicUC *section.GetSectionPublicUseCase,
	listUC *section.ListSectionsUseCase,
	updateUC *section.UpdateSectionUseCase,
	updatePositionUC *section.UpdateSectionPositionUseCase,
	bulkReorderUC *section.BulkReorderSectionsUseCase,
	deleteUC *section.DeleteSectionUseCase,
) *SectionController {
	return &SectionController{
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

// Create handles POST /api/sections/own
func (ctrl *SectionController) Create(c *gin.Context) {
	// Extract userID from context (set by auth middleware)
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Error: "unauthorized: missing user ID"})
		return
	}

	// Bind and validate HTTP request DTO
	var req request.CreateSectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: err.Error()})
		return
	}

	// Map HTTP request DTO to application DTO
	input := appdto.CreateSectionInput{
		Title:       req.Title,
		Description: req.Description,
		Position:    req.Position,
		Type:        req.Type,
		OwnerID:     userID,
		PortfolioID: req.PortfolioID,
	}

	// Execute use case
	sectionDTO, err := ctrl.createUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		status := pkgerrors.ToHTTPStatus(err)
		c.JSON(status, response.ErrorResponse{Error: err.Error()})
		return
	}

	// Map application DTO to HTTP response DTO
	resp := response.SectionResponse{
		ID:          sectionDTO.ID,
		Title:       sectionDTO.Title,
		Description: sectionDTO.Description,
		Position:    sectionDTO.Position,
		Type:        sectionDTO.Type,
		OwnerID:     sectionDTO.OwnerID,
		PortfolioID: sectionDTO.PortfolioID,
		CreatedAt:   sectionDTO.CreatedAt,
		UpdatedAt:   sectionDTO.UpdatedAt,
	}

	// Return HTTP response with API_OVERVIEW.md format
	c.JSON(http.StatusCreated, response.DataResponse{
		Data:    resp,
		Message: "Success",
	})
}

// List handles GET /api/sections/own
func (ctrl *SectionController) List(c *gin.Context) {
	// Extract userID from context (set by auth middleware)
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Error: "unauthorized: missing user ID"})
		return
	}

	// Bind and validate query parameters
	var req request.ListSectionsRequest
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
	input := appdto.ListSectionsInput{
		PortfolioID: 0, // List all sections for user
		Pagination: appdto.PaginationDTO{
			Page:  req.Page,
			Limit: req.Limit,
		},
	}

	// Execute use case
	output, err := ctrl.listUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		status := pkgerrors.ToHTTPStatus(err)
		c.JSON(status, response.ErrorResponse{Error: err.Error()})
		return
	}

	// Map to HTTP response DTOs
	sections := make([]response.SectionResponse, len(output.Sections))
	for i, sec := range output.Sections {
		sections[i] = response.SectionResponse{
			ID:          sec.ID,
			Title:       sec.Title,
			Description: sec.Description,
			Position:    sec.Position,
			Type:        sec.Type,
			OwnerID:     sec.OwnerID,
			PortfolioID: sec.PortfolioID,
			CreatedAt:   sec.CreatedAt,
			UpdatedAt:   sec.UpdatedAt,
		}
	}

	// Return HTTP response with API_OVERVIEW.md format
	c.JSON(http.StatusOK, response.PaginatedDataResponse{
		Data:    sections,
		Page:    output.Pagination.Page,
		Limit:   output.Pagination.Limit,
		Total:   output.Pagination.Total,
		Message: "Success",
	})
}

// GetByID handles GET /api/sections/own/:id
func (ctrl *SectionController) GetByID(c *gin.Context) {
	// Extract userID from context (set by auth middleware)
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Error: "unauthorized: missing user ID"})
		return
	}

	// Parse section ID from URL parameter
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: "invalid section ID"})
		return
	}

	// Execute use case
	sectionDTO, err := ctrl.getUseCase.Execute(c.Request.Context(), uint(id), userID)
	if err != nil {
		status := pkgerrors.ToHTTPStatus(err)
		c.JSON(status, response.ErrorResponse{Error: err.Error()})
		return
	}

	// Map to HTTP response DTO
	resp := response.SectionResponse{
		ID:          sectionDTO.ID,
		Title:       sectionDTO.Title,
		Description: sectionDTO.Description,
		Position:    sectionDTO.Position,
		Type:        sectionDTO.Type,
		OwnerID:     sectionDTO.OwnerID,
		PortfolioID: sectionDTO.PortfolioID,
		CreatedAt:   sectionDTO.CreatedAt,
		UpdatedAt:   sectionDTO.UpdatedAt,
	}

	// Return HTTP response with API_OVERVIEW.md format
	c.JSON(http.StatusOK, response.DataResponse{
		Data:    resp,
		Message: "Success",
	})
}

// Update handles PUT /api/sections/own/:id
func (ctrl *SectionController) Update(c *gin.Context) {
	// Extract userID from context (set by auth middleware)
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Error: "unauthorized: missing user ID"})
		return
	}

	// Parse section ID from URL parameter
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: "invalid section ID"})
		return
	}

	// Bind and validate HTTP request DTO
	var req request.UpdateSectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: err.Error()})
		return
	}

	// Map to application DTO
	input := appdto.UpdateSectionInput{
		ID:          uint(id),
		Title:       req.Title,
		Description: req.Description,
		Position:    req.Position,
		Type:        req.Type,
		OwnerID:     userID,
	}

	// Execute use case
	err = ctrl.updateUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		status := pkgerrors.ToHTTPStatus(err)
		c.JSON(status, response.ErrorResponse{Error: err.Error()})
		return
	}

	// Return success response with API_OVERVIEW.md format
	c.JSON(http.StatusOK, response.DataResponse{
		Data:    nil,
		Message: "Section updated successfully",
	})
}

// UpdatePosition handles PUT /api/sections/own/:id/position
func (ctrl *SectionController) UpdatePosition(c *gin.Context) {
	// Extract userID from context (set by auth middleware)
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Error: "unauthorized: missing user ID"})
		return
	}

	// Parse section ID from URL parameter
	idStr := c.Param("id")
	sectionID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: "invalid section ID"})
		return
	}

	// Bind and validate HTTP request DTO
	var req request.UpdateSectionPositionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: err.Error()})
		return
	}

	// Execute use case
	err = ctrl.updatePositionUseCase.Execute(c.Request.Context(), uint(sectionID), req.Position, userID)
	if err != nil {
		status := pkgerrors.ToHTTPStatus(err)
		c.JSON(status, response.ErrorResponse{Error: err.Error()})
		return
	}

	// Return success response with API_OVERVIEW.md format
	c.JSON(http.StatusOK, response.DataResponse{
		Data:    nil,
		Message: "Section position updated successfully",
	})
}

// BulkReorder handles PUT /api/sections/own/reorder
func (ctrl *SectionController) BulkReorder(c *gin.Context) {
	// Extract userID from context (set by auth middleware)
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Error: "unauthorized: missing user ID"})
		return
	}

	// Bind and validate HTTP request DTO
	var req request.BulkReorderSectionsRequest
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

	input := appdto.BulkUpdateSectionPositionsInput{
		Items:   items,
		OwnerID: userID,
	}

	// Execute use case
	if err := ctrl.bulkReorderUseCase.Execute(c.Request.Context(), input); err != nil {
		status := pkgerrors.ToHTTPStatus(err)
		c.JSON(status, response.ErrorResponse{Error: err.Error()})
		return
	}

	// Return success response with API_OVERVIEW.md format
	c.JSON(http.StatusOK, response.DataResponse{
		Data:    nil,
		Message: "Sections reordered successfully",
	})
}

// Delete handles DELETE /api/sections/own/:id
func (ctrl *SectionController) Delete(c *gin.Context) {
	// Extract userID from context (set by auth middleware)
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Error: "unauthorized: missing user ID"})
		return
	}

	// Parse section ID from URL parameter
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: "invalid section ID"})
		return
	}

	// Execute use case
	err = ctrl.deleteUseCase.Execute(c.Request.Context(), uint(id), userID)
	if err != nil {
		status := pkgerrors.ToHTTPStatus(err)
		c.JSON(status, response.ErrorResponse{Error: err.Error()})
		return
	}

	// Return success response with API_OVERVIEW.md format
	c.JSON(http.StatusOK, response.DataResponse{
		Data:    nil,
		Message: "Section deleted successfully",
	})
}

// GetPublicByID handles GET /api/sections/id/:id and GET /api/sections/public/:id
func (ctrl *SectionController) GetPublicByID(c *gin.Context) {
	// Parse section ID from URL parameter
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: "invalid section ID"})
		return
	}

	// Execute use case (no auth required for public access)
	sectionDTO, err := ctrl.getPublicUseCase.Execute(c.Request.Context(), uint(id))
	if err != nil {
		status := pkgerrors.ToHTTPStatus(err)
		c.JSON(status, response.ErrorResponse{Error: err.Error()})
		return
	}

	// Map to HTTP response DTO (don't include OwnerID in public response)
	resp := response.SectionResponse{
		ID:          sectionDTO.ID,
		Title:       sectionDTO.Title,
		Description: sectionDTO.Description,
		Position:    sectionDTO.Position,
		Type:        sectionDTO.Type,
		PortfolioID: sectionDTO.PortfolioID,
		CreatedAt:   sectionDTO.CreatedAt,
		UpdatedAt:   sectionDTO.UpdatedAt,
	}

	// Return HTTP response with API_OVERVIEW.md format
	c.JSON(http.StatusOK, response.DataResponse{
		Data:    resp,
		Message: "Success",
	})
}

// GetPublicSectionContents handles GET /api/sections/public/:id/contents
// TODO: Implement when SectionContent use cases are created
func (ctrl *SectionController) GetPublicSectionContents(c *gin.Context) {
	// Parse section ID from URL parameter
	idStr := c.Param("id")
	_, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: "invalid section ID"})
		return
	}

	// TODO: Implement section contents retrieval when SectionContent domain is ready
	c.JSON(http.StatusNotImplemented, response.ErrorResponse{Error: "not implemented yet"})
}
