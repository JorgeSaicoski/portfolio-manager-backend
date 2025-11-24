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
		resp.BadRequest(c, "Invalid request data")
		return
	}

	// Check if section exists and belongs to user's portfolio
	section, err := h.sectionRepo.GetByID(req.SectionID)
	if err != nil {
		resp.NotFound(c, "Section not found")
		return
	}

	portfolio, err := h.portfolioRepo.GetByIDBasic(section.PortfolioID)
	if err != nil {
		resp.NotFound(c, "Portfolio not found")
		return
	}

	if portfolio.OwnerID != userID {
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
		resp.BadRequest(c, err.Error())
		return
	}

	// Create content
	if err := h.repo.Create(content); err != nil {
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
		resp.BadRequest(c, "Invalid section ID")
		return
	}

	// Get contents
	contents, err := h.repo.GetBySectionID(uint(id))
	if err != nil {
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
		resp.BadRequest(c, "Invalid content ID")
		return
	}

	// Get content
	content, err := h.repo.GetByID(uint(id))
	if err != nil {
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
		resp.BadRequest(c, "Invalid content ID")
		return
	}

	// Get existing content
	existing, err := h.repo.GetByID(uint(id))
	if err != nil {
		resp.NotFound(c, "Content not found")
		return
	}

	// Check if section belongs to user
	section, err := h.sectionRepo.GetByID(existing.SectionID)
	if err != nil {
		resp.NotFound(c, "Section not found")
		return
	}

	if section.OwnerID != userID {
		resp.Forbidden(c, "Access denied")
		return
	}

	// Parse request body
	var req request.UpdateSectionContentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
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
		resp.BadRequest(c, err.Error())
		return
	}

	// Update content
	if err := h.repo.Update(existing); err != nil {
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
		resp.BadRequest(c, "Invalid content ID")
		return
	}

	// Get existing content
	existing, err := h.repo.GetByID(uint(id))
	if err != nil {
		resp.NotFound(c, "Content not found")
		return
	}

	// Check if section belongs to user
	section, err := h.sectionRepo.GetByID(existing.SectionID)
	if err != nil {
		resp.NotFound(c, "Section not found")
		return
	}

	if section.OwnerID != userID {
		resp.Forbidden(c, "Access denied")
		return
	}

	// Parse request body
	var req request.UpdateSectionContentOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "Invalid request data")
		return
	}

	// Update order
	if err := h.repo.UpdateOrder(uint(id), req.Order); err != nil {
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
		resp.BadRequest(c, "Invalid content ID")
		return
	}

	// Get existing content
	existing, err := h.repo.GetByID(uint(id))
	if err != nil {
		resp.NotFound(c, "Content not found")
		return
	}

	// Check if section belongs to user
	section, err := h.sectionRepo.GetByID(existing.SectionID)
	if err != nil {
		resp.NotFound(c, "Section not found")
		return
	}

	if section.OwnerID != userID {
		resp.Forbidden(c, "Access denied")
		return
	}

	// Delete content
	if err := h.repo.Delete(uint(id)); err != nil {
		logrus.WithFields(logrus.Fields{
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
