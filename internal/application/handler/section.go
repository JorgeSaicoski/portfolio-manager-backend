package handler

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/models"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/infrastructure/audit"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/infrastructure/metrics"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/infrastructure/repo"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/shared/response"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/shared/validator"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type SectionHandler struct {
	repo          repo.SectionRepository
	portfolioRepo repo.PortfolioRepository
	metrics       *metrics.Collector
}

func NewSectionHandler(repo repo.SectionRepository, portfolioRepo repo.PortfolioRepository, metrics *metrics.Collector) *SectionHandler {
	return &SectionHandler{
		repo:          repo,
		portfolioRepo: portfolioRepo,
		metrics:       metrics,
	}
}

func (h *SectionHandler) GetByUser(c *gin.Context) {
	userID := c.GetString("userID") // From auth middleware

	// Parse pagination parameters - using default values if not provided
	page := 1
	limit := 10
	if pageParam := c.Query("page"); pageParam != "" {
		if p, err := strconv.Atoi(pageParam); err == nil && p > 0 {
			page = p
		}
	}
	if limitParam := c.Query("limit"); limitParam != "" {
		if l, err := strconv.Atoi(limitParam); err == nil && l > 0 {
			limit = l
		}
	}

	offset := (page - 1) * limit

	sections, err := h.repo.GetByOwnerID(userID, limit, offset)
	if err != nil {
		response.InternalError(c, "Failed to retrieve sections")
		return
	}

	response.SuccessWithPagination(c, 200, "sections", sections, page, limit)
}

func (h *SectionHandler) GetByPortfolio(c *gin.Context) {
	// Extract portfolio ID from URL parameter
	// This handler is used by two routes:
	// 1. /api/portfolios/public/:id/sections
	// 2. /api/sections/portfolio/:id
	// Both use :id as the parameter name
	portfolioID := c.Param("id")

	// Validate portfolio ID parameter
	if portfolioID == "" {
		logrus.WithFields(logrus.Fields{
			"handler":   "GetByPortfolio",
			"path":      c.Request.URL.Path,
			"allParams": c.Params,
		}).Error("Portfolio ID parameter is missing or empty")
		response.BadRequest(c, "Portfolio ID is required")
		return
	}

	logrus.WithFields(logrus.Fields{
		"portfolioID": portfolioID,
		"handler":     "GetByPortfolio",
		"path":        c.Request.URL.Path,
	}).Info("GetByPortfolio handler called")

	sections, err := h.repo.GetByPortfolioID(portfolioID)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"portfolioID": portfolioID,
			"error":       err.Error(),
			"errorType":   fmt.Sprintf("%T", err),
		}).Error("Failed to get sections from repository")
		response.InternalErrorWithDetails(c, "Failed to retrieve sections", err)
		return
	}

	logrus.WithFields(logrus.Fields{
		"portfolioID":   portfolioID,
		"sectionsCount": len(sections),
	}).Info("Successfully retrieved sections")

	response.OK(c, "sections", sections, "Success")
}

func (h *SectionHandler) GetByID(c *gin.Context) {
	sectionID := c.Param("id")

	// Parse section ID
	id, err := strconv.Atoi(sectionID)
	if err != nil {
		response.BadRequest(c, "Invalid section ID")
		return
	}

	section, err := h.repo.GetByID(uint(id))
	if err != nil {
		response.NotFound(c, "Section not found")
		return
	}

	response.OK(c, "section", section, "Success")
}

func (h *SectionHandler) GetByType(c *gin.Context) {
	sectionType := c.Query("type")
	if sectionType == "" {
		response.BadRequest(c, "Section type is required")
		return
	}

	sections, err := h.repo.GetByType(sectionType)
	if err != nil {
		response.InternalError(c, "Failed to retrieve sections")
		return
	}

	response.OK(c, "sections", sections, "Success")
}

func (h *SectionHandler) Create(c *gin.Context) {
	userID := c.GetString("userID") // From auth middleware

	// Parse request body
	var newSection models.Section
	if err := c.ShouldBindJSON(&newSection); err != nil {
		response.BadRequest(c, "Invalid request data")
		return
	}

	// Set the owner
	newSection.OwnerID = userID

	// Validate section data first (includes portfolioID check)
	if err := validator.ValidateSection(&newSection); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Validate portfolio exists and belongs to user
	portfolio, err := h.portfolioRepo.GetByIDBasic(newSection.PortfolioID)
	if err != nil {
		response.NotFound(c, "Portfolio not found")
		return
	}

	if portfolio.OwnerID != userID {
		response.ForbiddenWithDetails(c, "Access denied: portfolio belongs to another user", map[string]interface{}{
			"resource_type": "portfolio",
			"resource_id":   portfolio.ID,
			"owner_id":      portfolio.OwnerID,
			"action":        "create_section",
		})
		return
	}

	// Check for duplicate title
	isDuplicate, err := h.repo.CheckDuplicate(newSection.Title, newSection.PortfolioID, 0)
	if err != nil {
		response.InternalError(c, "Failed to check for duplicate section")
		return
	}
	if isDuplicate {
		response.BadRequest(c, "Section with this title already exists in this portfolio")
		return
	}

	// Create a section
	if err := h.repo.Create(&newSection); err != nil {
		// Check if error is due to foreign key constraint (invalid portfolio_id)
		errMsg := err.Error()
		if strings.Contains(errMsg, "fk_portfolios_sections") || strings.Contains(errMsg, "23503") {
			response.NotFound(c, "Portfolio not found")
			return
		}
		response.InternalError(c, "Failed to create section")
		return
	}

	response.Created(c, "section", &newSection, "Section created successfully")
}

func (h *SectionHandler) Update(c *gin.Context) {
	userID := c.GetString("userID") // From auth middleware
	sectionID := c.Param("id")

	// Parse section ID
	id, err := strconv.Atoi(sectionID)
	if err != nil {
		response.BadRequest(c, "Invalid section ID")
		return
	}

	// Parse request body
	var updateData models.Section
	if err := c.ShouldBindJSON(&updateData); err != nil {
		response.BadRequest(c, "Invalid request data")
		return
	}

	// Set the ID and owner
	updateData.ID = uint(id)
	updateData.OwnerID = userID

	// Validate section data
	if err := validator.ValidateSection(&updateData); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Check if section exists and belongs to user
	existing, err := h.repo.GetByID(uint(id))
	if err != nil {
		response.NotFound(c, "Section not found")
		return
	}
	if existing.OwnerID != userID {
		response.ForbiddenWithDetails(c, "Access denied", map[string]interface{}{
			"resource_type": "section",
			"resource_id":   existing.ID,
			"owner_id":      existing.OwnerID,
			"action":        "update",
		})
		return
	}

	// Check for duplicate title
	isDuplicate, err := h.repo.CheckDuplicate(updateData.Title, updateData.PortfolioID, updateData.ID)
	if err != nil {
		response.InternalError(c, "Failed to check for duplicate section")
		return
	}
	if isDuplicate {
		response.BadRequest(c, "Section with this title already exists in this portfolio")
		return
	}

	// Update section
	if err := h.repo.Update(&updateData); err != nil {
		// Check if error is due to foreign key constraint (invalid portfolio_id)
		errMsg := err.Error()
		if strings.Contains(errMsg, "fk_portfolios_sections") || strings.Contains(errMsg, "23503") {
			response.NotFound(c, "Portfolio not found")
			return
		}
		response.InternalError(c, "Failed to update section")
		return
	}

	response.OK(c, "section", &updateData, "Section updated successfully")
}

func (h *SectionHandler) Delete(c *gin.Context) {
	userID := c.GetString("userID")
	sectionID := c.Param("id")

	id, err := strconv.Atoi(sectionID)
	if err != nil {
		response.BadRequest(c, "Invalid section ID")
		return
	}

	// Get a section to check ownership
	section, err := h.repo.GetByID(uint(id))
	if err != nil {
		response.NotFound(c, "Section not found")
		return
	}

	if section.OwnerID != userID {
		response.ForbiddenWithDetails(c, "Access denied", map[string]interface{}{
			"resource_type": "section",
			"resource_id":   section.ID,
			"owner_id":      section.OwnerID,
			"action":        "delete",
		})
		return
	}

	// Delete section (CASCADE: all related section_contents will be deleted)
	if err := h.repo.Delete(uint(id)); err != nil {
		logrus.WithFields(logrus.Fields{
			"sectionID":   id,
			"userID":      userID,
			"portfolioID": section.PortfolioID,
			"error":       err.Error(),
		}).Error("Failed to delete section")

		response.InternalError(c, "Failed to delete section")
		return
	}

	audit.GetDeleteLogger().WithFields(logrus.Fields{
		"operation":   "DELETE_SECTION",
		"sectionID":   id,
		"userID":      userID,
		"portfolioID": section.PortfolioID,
		"title":       section.Title,
		"cascade":     "section_contents",
	}).Info("Section deleted successfully (CASCADE: all related section_contents)")

	response.OK(c, "message", "Section deleted successfully", "Success")
}

// UpdatePosition updates the position field of a section
func (h *SectionHandler) UpdatePosition(c *gin.Context) {
	userID := c.GetString("userID") // From auth middleware
	sectionID := c.Param("id")

	// Parse section ID
	id, err := strconv.Atoi(sectionID)
	if err != nil {
		response.BadRequest(c, "Invalid section ID")
		return
	}

	// Parse request body
	var req struct {
		Position uint `json:"position" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request data")
		return
	}

	// Check if the section exists and belongs to a user
	existing, err := h.repo.GetByID(uint(id))
	if err != nil {
		response.NotFound(c, "Section not found")
		return
	}

	if existing.OwnerID != userID {
		response.ForbiddenWithDetails(c, "Access denied", map[string]interface{}{
			"resource_type": "section",
			"resource_id":   existing.ID,
			"owner_id":      existing.OwnerID,
			"action":        "update_position",
		})
		return
	}

	// Update position
	if err := h.repo.UpdatePosition(uint(id), req.Position); err != nil {
		response.InternalError(c, "Failed to update section position")
		return
	}

	response.OK(c, "message", "Section position updated successfully", "Success")
}
