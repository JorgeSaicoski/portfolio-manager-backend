package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/dto"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/usecases/section_content"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/interfaces/dto/request"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/interfaces/dto/response"
	"github.com/JorgeSaicoski/portfolio-manager/backend/pkg/pkgerrors"
)

// SectionContentController handles HTTP requests for section contents
type SectionContentController struct {
	createUseCase        *section_content.CreateSectionContentUseCase
	updateUseCase        *section_content.UpdateSectionContentUseCase
	updateOrderUseCase   *section_content.UpdateSectionContentOrderUseCase
	deleteUseCase        *section_content.DeleteSectionContentUseCase
	getPublicUseCase     *section_content.GetSectionContentPublicUseCase
	listBySectionUseCase *section_content.ListSectionContentsBySectionUseCase
}

// NewSectionContentController creates a new section content controller instance
func NewSectionContentController(
	createUC *section_content.CreateSectionContentUseCase,
	updateUC *section_content.UpdateSectionContentUseCase,
	updateOrderUC *section_content.UpdateSectionContentOrderUseCase,
	deleteUC *section_content.DeleteSectionContentUseCase,
	getPublicUC *section_content.GetSectionContentPublicUseCase,
	listBySectionUC *section_content.ListSectionContentsBySectionUseCase,
) *SectionContentController {
	return &SectionContentController{
		createUseCase:        createUC,
		updateUseCase:        updateUC,
		updateOrderUseCase:   updateOrderUC,
		deleteUseCase:        deleteUC,
		getPublicUseCase:     getPublicUC,
		listBySectionUseCase: listBySectionUC,
	}
}

// Create handles POST /api/section-contents/own
func (ctrl *SectionContentController) Create(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Error: "unauthorized"})
		return
	}

	var req request.CreateSectionContentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: err.Error()})
		return
	}

	input := dto.CreateSectionContentInput{
		SectionID: req.SectionID,
		Type:      req.Type,
		Content:   req.Content,
		Order:     req.Order,
		ImageID:   req.ImageID,
		OwnerID:   userID,
	}

	content, err := ctrl.createUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		status := pkgerrors.ToHTTPStatus(err)
		c.JSON(status, response.ErrorResponse{Error: err.Error()})
		return
	}

	resp := response.SectionContentResponse{
		ID:        content.ID,
		SectionID: content.SectionID,
		Type:      content.Type,
		Content:   content.Content,
		Order:     content.Order,
		ImageID:   content.ImageID,
		OwnerID:   content.OwnerID,
		CreatedAt: content.CreatedAt,
		UpdatedAt: content.UpdatedAt,
	}

	c.JSON(http.StatusCreated, response.DataResponse{
		Data:    resp,
		Message: "Success",
	})
}

// Update handles PUT /api/section-contents/own/:id
func (ctrl *SectionContentController) Update(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Error: "unauthorized"})
		return
	}

	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: "invalid section content ID"})
		return
	}

	var req request.UpdateSectionContentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: err.Error()})
		return
	}

	input := dto.UpdateSectionContentInput{
		ID:      uint(id),
		Type:    req.Type,
		Content: req.Content,
		Order:   req.Order,
		ImageID: req.ImageID,
		OwnerID: userID,
	}

	if err := ctrl.updateUseCase.Execute(c.Request.Context(), input); err != nil {
		status := pkgerrors.ToHTTPStatus(err)
		c.JSON(status, response.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.DataResponse{
		Data:    nil,
		Message: "Success",
	})
}

// UpdateOrder handles PATCH /api/section-contents/own/:id/order
func (ctrl *SectionContentController) UpdateOrder(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Error: "unauthorized"})
		return
	}

	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: "invalid section content ID"})
		return
	}

	var req request.UpdateSectionContentOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: err.Error()})
		return
	}

	if err := ctrl.updateOrderUseCase.Execute(c.Request.Context(), uint(id), req.Order, userID); err != nil {
		status := pkgerrors.ToHTTPStatus(err)
		c.JSON(status, response.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.DataResponse{
		Data:    nil,
		Message: "Success",
	})
}

// Delete handles DELETE /api/section-contents/own/:id
func (ctrl *SectionContentController) Delete(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Error: "unauthorized"})
		return
	}

	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: "invalid section content ID"})
		return
	}

	if err := ctrl.deleteUseCase.Execute(c.Request.Context(), uint(id), userID); err != nil {
		status := pkgerrors.ToHTTPStatus(err)
		c.JSON(status, response.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.DataResponse{
		Data:    nil,
		Message: "Success",
	})
}

// GetByID handles GET /api/section-contents/:id
func (ctrl *SectionContentController) GetByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: "invalid section content ID"})
		return
	}

	content, err := ctrl.getPublicUseCase.Execute(c.Request.Context(), uint(id))
	if err != nil {
		status := pkgerrors.ToHTTPStatus(err)
		c.JSON(status, response.ErrorResponse{Error: err.Error()})
		return
	}

	resp := response.SectionContentResponse{
		ID:        content.ID,
		SectionID: content.SectionID,
		Type:      content.Type,
		Content:   content.Content,
		Order:     content.Order,
		ImageID:   content.ImageID,
		OwnerID:   content.OwnerID,
		CreatedAt: content.CreatedAt,
		UpdatedAt: content.UpdatedAt,
	}

	c.JSON(http.StatusOK, response.DataResponse{
		Data:    resp,
		Message: "Success",
	})
}

// ListBySection handles GET /api/section-contents/sections/:sectionId/contents
func (ctrl *SectionContentController) ListBySection(c *gin.Context) {
	sectionIDParam := c.Param("sectionId")
	sectionID, err := strconv.ParseUint(sectionIDParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: "invalid section ID"})
		return
	}

	contents, err := ctrl.listBySectionUseCase.Execute(c.Request.Context(), uint(sectionID))
	if err != nil {
		status := pkgerrors.ToHTTPStatus(err)
		c.JSON(status, response.ErrorResponse{Error: err.Error()})
		return
	}

	respContents := make([]response.SectionContentResponse, len(contents))
	for i, content := range contents {
		respContents[i] = response.SectionContentResponse{
			ID:        content.ID,
			SectionID: content.SectionID,
			Type:      content.Type,
			Content:   content.Content,
			Order:     content.Order,
			ImageID:   content.ImageID,
			OwnerID:   content.OwnerID,
			CreatedAt: content.CreatedAt,
			UpdatedAt: content.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, response.DataResponse{
		Data:    respContents,
		Message: "Success",
	})
}
