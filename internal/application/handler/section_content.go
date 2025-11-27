package handler

import (
	"strconv"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/models"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/infrastructure/audit"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/infrastructure/metrics"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/infrastructure/repo"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/shared/dto/request"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/shared/dto/response"
	resp "github.com/JorgeSaicoski/portfolio-manager/backend/internal/shared/response"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/shared/validator"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type SectionContentHandler struct {
	repo          repo.SectionContentRepository
	sectionRepo   repo.SectionRepository   // For authorization checks
	portfolioRepo repo.PortfolioRepository // For full ownership validation
	metrics       *metrics.Collector
}

func NewSectionContentHandler(repo repo.SectionContentRepository, sectionRepo repo.SectionRepository, portfolioRepo repo.PortfolioRepository, metrics *metrics.Collector) *SectionContentHandler {
	return &SectionContentHandler{
		repo:          repo,
		sectionRepo:   sectionRepo,
		portfolioRepo: portfolioRepo,
		metrics:       metrics,
	}
}

// Create creates a new section content block
func (h *SectionContentHandler) Create(c *gin.Context) {
	userID := c.GetString("userID") // From auth middleware

	// Parse request body
	var req request.CreateSectionContentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "CREATE_SECTION_CONTENT_BAD_REQUEST",
			"where":     "backend/internal/application/handler/section_content.go",
			"function":  "Create",
			"userID":    userID,
			"error":     err.Error(),
		}).Warn("Invalid request data")
		resp.BadRequest(c, "Invalid request data")
		return
	}

	// Check if section exists and belongs to user's portfolio
	section, err := h.sectionRepo.GetByID(req.SectionID)
	if err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "CREATE_SECTION_CONTENT_SECTION_NOT_FOUND",
			"where":     "backend/internal/application/handler/section_content.go",
			"function":  "Create",
			"userID":    userID,
			"sectionID": req.SectionID,
			"error":     err.Error(),
		}).Warn("Section not found")
		resp.NotFound(c, "Section not found")
		return
	}

	portfolio, err := h.portfolioRepo.GetByIDBasic(section.PortfolioID)
	if err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":   "CREATE_SECTION_CONTENT_PORTFOLIO_NOT_FOUND",
			"where":       "backend/internal/application/handler/section_content.go",
			"function":    "Create",
			"userID":      userID,
			"sectionID":   req.SectionID,
			"portfolioID": section.PortfolioID,
			"error":       err.Error(),
		}).Warn("Portfolio not found")
		resp.NotFound(c, "Portfolio not found")
		return
	}

	if portfolio.OwnerID != userID {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":   "CREATE_SECTION_CONTENT_FORBIDDEN",
			"where":       "backend/internal/application/handler/section_content.go",
			"function":    "Create",
			"userID":      userID,
			"sectionID":   req.SectionID,
			"portfolioID": section.PortfolioID,
			"ownerID":     portfolio.OwnerID,
		}).Warn("Access denied: section belongs to another user's portfolio")
		resp.Forbidden(c, "Access denied: section belongs to another user's portfolio")
		return
	}

	// Create content model
	content := &models.SectionContent{
		SectionID: req.SectionID,
		Type:      req.Type,
		Content:   req.Content,
		Order:     0, // Default order
		Metadata:  req.Metadata,
		OwnerID:   userID,
	}

	// Set custom order if provided
	if req.Order != nil {
		content.Order = *req.Order
	}

	// Validate content
	if err := validator.ValidateSectionContent(content); err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "CREATE_SECTION_CONTENT_VALIDATION_ERROR",
			"where":     "backend/internal/application/handler/section_content.go",
			"function":  "Create",
			"userID":    userID,
			"sectionID": req.SectionID,
			"error":     err.Error(),
		}).Warn("Section content validation failed")
		resp.BadRequest(c, err.Error())
		return
	}

	// Create content
	if err := h.repo.Create(content); err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "CREATE_SECTION_CONTENT_DB_ERROR",
			"where":     "backend/internal/application/handler/section_content.go",
			"function":  "Create",
			"userID":    userID,
			"sectionID": req.SectionID,
			"error":     err.Error(),
		}).Error("Failed to create content")
		resp.InternalError(c, "Failed to create content")
		return
	}

	resp.Created(c, "content", response.ToSectionContentResponse(content), "Content created successfully")
}

// GetBySectionID retrieves all content blocks for a section
func (h *SectionContentHandler) GetBySectionID(c *gin.Context) {
	sectionID := c.Param("sectionId")

	// Parse section ID
	id, err := strconv.ParseUint(sectionID, 10, 32)
	if err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "GET_SECTION_CONTENTS_INVALID_ID",
			"where":     "backend/internal/application/handler/section_content.go",
			"function":  "GetBySectionID",
			"sectionID": sectionID,
			"error":     err.Error(),
		}).Warn("Invalid section ID")
		resp.BadRequest(c, "Invalid section ID")
		return
	}

	// Get contents
	contents, err := h.repo.GetBySectionID(uint(id))
	if err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "GET_SECTION_CONTENTS_DB_ERROR",
			"where":     "backend/internal/application/handler/section_content.go",
			"function":  "GetBySectionID",
			"sectionID": id,
			"error":     err.Error(),
		}).Error("Failed to retrieve contents")
		resp.InternalError(c, "Failed to retrieve contents")
		return
	}

	resp.OK(c, "contents", response.ToSectionContentListResponse(contents), "Success")
}

// GetByID retrieves a single content block
func (h *SectionContentHandler) GetByID(c *gin.Context) {
	contentID := c.Param("id")

	// Parse content ID
	id, err := strconv.Atoi(contentID)
	if err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "GET_SECTION_CONTENT_INVALID_ID",
			"where":     "backend/internal/application/handler/section_content.go",
			"function":  "GetByID",
			"contentID": contentID,
			"error":     err.Error(),
		}).Warn("Invalid content ID")
		resp.BadRequest(c, "Invalid content ID")
		return
	}

	// Get content
	content, err := h.repo.GetByID(uint(id))
	if err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "GET_SECTION_CONTENT_NOT_FOUND",
			"where":     "backend/internal/application/handler/section_content.go",
			"function":  "GetByID",
			"contentID": id,
			"error":     err.Error(),
		}).Warn("Content not found")
		resp.NotFound(c, "Content not found")
		return
	}

	resp.OK(c, "content", response.ToSectionContentResponse(content), "Success")
}

// Update updates a section content block
func (h *SectionContentHandler) Update(c *gin.Context) {
	userID := c.GetString("userID") // From auth middleware
	contentID := c.Param("id")

	// Parse content ID
	id, err := strconv.Atoi(contentID)
	if err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "UPDATE_SECTION_CONTENT_INVALID_ID",
			"where":     "backend/internal/application/handler/section_content.go",
			"function":  "Update",
			"userID":    userID,
			"contentID": contentID,
			"error":     err.Error(),
		}).Warn("Invalid content ID")
		resp.BadRequest(c, "Invalid content ID")
		return
	}

	// Get existing content
	existing, err := h.repo.GetByID(uint(id))
	if err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "UPDATE_SECTION_CONTENT_NOT_FOUND",
			"where":     "backend/internal/application/handler/section_content.go",
			"function":  "Update",
			"userID":    userID,
			"contentID": id,
			"error":     err.Error(),
		}).Warn("Content not found")
		resp.NotFound(c, "Content not found")
		return
	}

	// Check if section belongs to user
	section, err := h.sectionRepo.GetByID(existing.SectionID)
	if err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "UPDATE_SECTION_CONTENT_SECTION_NOT_FOUND",
			"where":     "backend/internal/application/handler/section_content.go",
			"function":  "Update",
			"userID":    userID,
			"contentID": id,
			"sectionID": existing.SectionID,
			"error":     err.Error(),
		}).Warn("Section not found")
		resp.NotFound(c, "Section not found")
		return
	}

	if section.OwnerID != userID {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "UPDATE_SECTION_CONTENT_FORBIDDEN",
			"where":     "backend/internal/application/handler/section_content.go",
			"function":  "Update",
			"userID":    userID,
			"contentID": id,
			"sectionID": existing.SectionID,
			"ownerID":   section.OwnerID,
		}).Warn("Access denied")
		resp.Forbidden(c, "Access denied")
		return
	}

	// Parse request body
	var req request.UpdateSectionContentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "UPDATE_SECTION_CONTENT_BAD_REQUEST",
			"where":     "backend/internal/application/handler/section_content.go",
			"function":  "Update",
			"userID":    userID,
			"contentID": id,
			"error":     err.Error(),
		}).Warn("Invalid request data")
		resp.BadRequest(c, "Invalid request data")
		return
	}

	// Update fields
	if req.Type != "" {
		existing.Type = req.Type
	}
	if req.Content != "" {
		existing.Content = req.Content
	}
	if req.Order != nil {
		existing.Order = *req.Order
	}
	if req.Metadata != nil {
		existing.Metadata = req.Metadata
	}

	// Validate content
	if err := validator.ValidateSectionContent(existing); err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "UPDATE_SECTION_CONTENT_VALIDATION_ERROR",
			"where":     "backend/internal/application/handler/section_content.go",
			"function":  "Update",
			"userID":    userID,
			"contentID": id,
			"sectionID": existing.SectionID,
			"error":     err.Error(),
		}).Warn("Section content validation failed")
		resp.BadRequest(c, err.Error())
		return
	}

	// Update content
	if err := h.repo.Update(existing); err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "UPDATE_SECTION_CONTENT_DB_ERROR",
			"where":     "backend/internal/application/handler/section_content.go",
			"function":  "Update",
			"userID":    userID,
			"contentID": id,
			"sectionID": existing.SectionID,
			"error":     err.Error(),
		}).Error("Failed to update content")
		resp.InternalError(c, "Failed to update content")
		return
	}

	resp.OK(c, "content", response.ToSectionContentResponse(existing), "Content updated successfully")
}

// UpdateOrder updates only the order field of a content block
func (h *SectionContentHandler) UpdateOrder(c *gin.Context) {
	userID := c.GetString("userID") // From auth middleware
	contentID := c.Param("id")

	// Parse content ID
	id, err := strconv.Atoi(contentID)
	if err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "UPDATE_SECTION_CONTENT_ORDER_INVALID_ID",
			"where":     "backend/internal/application/handler/section_content.go",
			"function":  "UpdateOrder",
			"userID":    userID,
			"contentID": contentID,
			"error":     err.Error(),
		}).Warn("Invalid content ID")
		resp.BadRequest(c, "Invalid content ID")
		return
	}

	// Get existing content
	existing, err := h.repo.GetByID(uint(id))
	if err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "UPDATE_SECTION_CONTENT_ORDER_NOT_FOUND",
			"where":     "backend/internal/application/handler/section_content.go",
			"function":  "UpdateOrder",
			"userID":    userID,
			"contentID": id,
			"error":     err.Error(),
		}).Warn("Content not found")
		resp.NotFound(c, "Content not found")
		return
	}

	// Check if section belongs to user
	section, err := h.sectionRepo.GetByID(existing.SectionID)
	if err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "UPDATE_SECTION_CONTENT_ORDER_SECTION_NOT_FOUND",
			"where":     "backend/internal/application/handler/section_content.go",
			"function":  "UpdateOrder",
			"userID":    userID,
			"contentID": id,
			"sectionID": existing.SectionID,
			"error":     err.Error(),
		}).Warn("Section not found")
		resp.NotFound(c, "Section not found")
		return
	}

	if section.OwnerID != userID {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "UPDATE_SECTION_CONTENT_ORDER_FORBIDDEN",
			"where":     "backend/internal/application/handler/section_content.go",
			"function":  "UpdateOrder",
			"userID":    userID,
			"contentID": id,
			"sectionID": existing.SectionID,
			"ownerID":   section.OwnerID,
		}).Warn("Access denied")
		resp.Forbidden(c, "Access denied")
		return
	}

	// Parse request body
	var req request.UpdateSectionContentOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "UPDATE_SECTION_CONTENT_ORDER_BAD_REQUEST",
			"where":     "backend/internal/application/handler/section_content.go",
			"function":  "UpdateOrder",
			"userID":    userID,
			"contentID": id,
			"error":     err.Error(),
		}).Warn("Invalid request data")
		resp.BadRequest(c, "Invalid request data")
		return
	}

	// Update order
	if err := h.repo.UpdateOrder(uint(id), req.Order); err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "UPDATE_SECTION_CONTENT_ORDER_DB_ERROR",
			"where":     "backend/internal/application/handler/section_content.go",
			"function":  "UpdateOrder",
			"userID":    userID,
			"contentID": id,
			"order":     req.Order,
			"error":     err.Error(),
		}).Error("Failed to update content order")
		resp.InternalError(c, "Failed to update content order")
		return
	}

	resp.OK(c, "message", "Content order updated successfully", "Success")
}

// Delete deletes a section content block
func (h *SectionContentHandler) Delete(c *gin.Context) {
	userID := c.GetString("userID") // From auth middleware
	contentID := c.Param("id")

	// Parse content ID
	id, err := strconv.Atoi(contentID)
	if err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "DELETE_SECTION_CONTENT_INVALID_ID",
			"where":     "backend/internal/application/handler/section_content.go",
			"function":  "Delete",
			"userID":    userID,
			"contentID": contentID,
			"error":     err.Error(),
		}).Warn("Invalid content ID")
		resp.BadRequest(c, "Invalid content ID")
		return
	}

	// Get existing content
	existing, err := h.repo.GetByID(uint(id))
	if err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "DELETE_SECTION_CONTENT_NOT_FOUND",
			"where":     "backend/internal/application/handler/section_content.go",
			"function":  "Delete",
			"userID":    userID,
			"contentID": id,
			"error":     err.Error(),
		}).Warn("Content not found")
		resp.NotFound(c, "Content not found")
		return
	}

	// Check if section belongs to user
	section, err := h.sectionRepo.GetByID(existing.SectionID)
	if err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "DELETE_SECTION_CONTENT_SECTION_NOT_FOUND",
			"where":     "backend/internal/application/handler/section_content.go",
			"function":  "Delete",
			"userID":    userID,
			"contentID": id,
			"sectionID": existing.SectionID,
			"error":     err.Error(),
		}).Warn("Section not found")
		resp.NotFound(c, "Section not found")
		return
	}

	if section.OwnerID != userID {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "DELETE_SECTION_CONTENT_FORBIDDEN",
			"where":     "backend/internal/application/handler/section_content.go",
			"function":  "Delete",
			"userID":    userID,
			"contentID": id,
			"sectionID": existing.SectionID,
			"ownerID":   section.OwnerID,
		}).Warn("Access denied")
		resp.Forbidden(c, "Access denied")
		return
	}

	// Delete content
	if err := h.repo.Delete(uint(id)); err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "DELETE_SECTION_CONTENT_DB_ERROR",
			"where":     "backend/internal/application/handler/section_content.go",
			"function":  "Delete",
			"contentID": id,
			"userID":    userID,
			"sectionID": existing.SectionID,
			"error":     err.Error(),
		}).Error("Failed to delete section content")

		resp.InternalError(c, "Failed to delete content")
		return
	}

	audit.GetDeleteLogger().WithFields(logrus.Fields{
		"operation": "DELETE_SECTION_CONTENT",
		"contentID": id,
		"userID":    userID,
		"sectionID": existing.SectionID,
		"type":      existing.Type,
	}).Info("Section content deleted successfully")

	resp.OK(c, "message", "Content deleted successfully", "Success")
}
