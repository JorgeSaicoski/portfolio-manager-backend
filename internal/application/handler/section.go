package handler

import (
	"encoding/json"
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

type SectionBulkReorderRequest struct {
	Items []struct {
		ID       uint `json:"id" binding:"required"`
		Position uint `json:"position" binding:"required,min=1"`
	} `json:"items" binding:"required,min=1"`
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

	sections, total, err := h.repo.GetByOwnerID(userID, limit, offset)
	if err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "GET_SECTIONS_BY_USER_DB_ERROR",
			"where":     "backend/internal/application/handler/section.go",
			"function":  "GetByUser",
			"userID":    userID,
			"error":     err.Error(),
		}).Error("Failed to retrieve sections")
		response.InternalError(c, "Failed to retrieve sections")
		return
	}

	response.SuccessWithPagination(c, 200, "sections", sections, page, limit, total)
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
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "GET_SECTIONS_BY_PORTFOLIO_MISSING_ID",
			"where":     "backend/internal/application/handler/section.go",
			"function":  "GetByPortfolio",
			"path":      c.Request.URL.Path,
			"allParams": c.Params,
		}).Warn("Portfolio ID parameter is missing or empty")
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
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":   "GET_SECTIONS_BY_PORTFOLIO_DB_ERROR",
			"where":       "backend/internal/application/handler/section.go",
			"function":    "GetByPortfolio",
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
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "GET_SECTION_BY_ID_INVALID_ID",
			"where":     "backend/internal/application/handler/section.go",
			"function":  "GetByID",
			"sectionID": sectionID,
			"error":     err.Error(),
		}).Warn("Invalid section ID")
		response.BadRequest(c, "Invalid section ID")
		return
	}

	section, err := h.repo.GetByID(uint(id))
	if err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "GET_SECTION_BY_ID_NOT_FOUND",
			"where":     "backend/internal/application/handler/section.go",
			"function":  "GetByID",
			"sectionID": id,
			"error":     err.Error(),
		}).Warn("Section not found")
		response.NotFound(c, "Section not found")
		return
	}

	response.OK(c, "section", section, "Success")
}

func (h *SectionHandler) GetByType(c *gin.Context) {
	sectionType := c.Query("type")
	if sectionType == "" {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "GET_SECTIONS_BY_TYPE_MISSING_TYPE",
			"where":     "backend/internal/application/handler/section.go",
			"function":  "GetByType",
		}).Warn("Section type is required")
		response.BadRequest(c, "Section type is required")
		return
	}

	sections, err := h.repo.GetByType(sectionType)
	if err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":   "GET_SECTIONS_BY_TYPE_DB_ERROR",
			"where":       "backend/internal/application/handler/section.go",
			"function":    "GetByType",
			"sectionType": sectionType,
			"error":       err.Error(),
		}).Error("Failed to retrieve sections")
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
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "CREATE_SECTION_BAD_REQUEST",
			"where":     "backend/internal/application/handler/section.go",
			"function":  "Create",
			"userID":    userID,
			"error":     err.Error(),
		}).Warn("Invalid request data")
		response.BadRequest(c, "Invalid request data")
		return
	}

	reqJSON, _ := json.Marshal(newSection)
	logrus.WithFields(logrus.Fields{
		"userID":  userID,
		"request": string(reqJSON),
	}).Info("Parsed section creation request")

	// Set the owner
	newSection.OwnerID = userID

	// Validate section data first (includes portfolioID check)
	if err := validator.ValidateSection(&newSection); err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":   "CREATE_SECTION_VALIDATION_ERROR",
			"where":       "backend/internal/application/handler/section.go",
			"function":    "Create",
			"userID":      userID,
			"portfolioID": newSection.PortfolioID,
			"error":       err.Error(),
		}).Warn("Section validation failed")
		response.BadRequest(c, err.Error())
		return
	}

	// Validate portfolio exists and belongs to user
	portfolio, err := h.portfolioRepo.GetByIDBasic(newSection.PortfolioID)
	if err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":   "CREATE_SECTION_PORTFOLIO_NOT_FOUND",
			"where":       "backend/internal/application/handler/section.go",
			"function":    "Create",
			"userID":      userID,
			"portfolioID": newSection.PortfolioID,
			"error":       err.Error(),
		}).Warn("Portfolio not found")
		response.NotFound(c, "Portfolio not found")
		return
	}

	if portfolio.OwnerID != userID {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":   "CREATE_SECTION_FORBIDDEN",
			"where":       "backend/internal/application/handler/section.go",
			"function":    "Create",
			"userID":      userID,
			"portfolioID": newSection.PortfolioID,
			"ownerID":     portfolio.OwnerID,
		}).Warn("Access denied to portfolio")
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
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":   "CREATE_SECTION_DUPLICATE_CHECK_ERROR",
			"where":       "backend/internal/application/handler/section.go",
			"function":    "Create",
			"userID":      userID,
			"portfolioID": newSection.PortfolioID,
			"title":       newSection.Title,
			"error":       err.Error(),
		}).Error("Failed to check for duplicate section")
		response.InternalError(c, "Failed to check for duplicate section")
		return
	}
	if isDuplicate {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":   "CREATE_SECTION_DUPLICATE_TITLE",
			"where":       "backend/internal/application/handler/section.go",
			"function":    "Create",
			"userID":      userID,
			"portfolioID": newSection.PortfolioID,
			"title":       newSection.Title,
		}).Warn("Section with this title already exists in this portfolio")
		response.BadRequest(c, "Section with this title already exists in this portfolio")
		return
	}

	logrus.WithFields(logrus.Fields{
		"userID":       userID,
		"title":        newSection.Title,
		"portfolio_id": newSection.PortfolioID,
	}).Info("Creating section - position will be set by database trigger")

	// Create a section
	if err := h.repo.Create(&newSection); err != nil {
		// Check if error is due to foreign key constraint (invalid portfolio_id)
		errMsg := err.Error()
		if strings.Contains(errMsg, "fk_portfolios_sections") || strings.Contains(errMsg, "23503") {
			audit.GetErrorLogger().WithFields(logrus.Fields{
				"operation":   "CREATE_SECTION_FK_CONSTRAINT_ERROR",
				"where":       "backend/internal/application/handler/section.go",
				"function":    "Create",
				"userID":      userID,
				"portfolioID": newSection.PortfolioID,
				"error":       err.Error(),
			}).Warn("Portfolio not found during section creation")
			response.NotFound(c, "Portfolio not found")
			return
		}
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":   "CREATE_SECTION_DB_ERROR",
			"where":       "backend/internal/application/handler/section.go",
			"function":    "Create",
			"userID":      userID,
			"portfolioID": newSection.PortfolioID,
			"error":       err.Error(),
		}).Error("Failed to create section")
		response.InternalError(c, "Failed to create section")
		return
	}

	audit.GetCreateLogger().WithFields(logrus.Fields{
		"operation":   "CREATE_SECTION",
		"userID":      userID,
		"sectionID":   newSection.ID,
		"title":       newSection.Title,
		"portfolioID": newSection.PortfolioID,
		"position":    newSection.Position,
	}).Info("Section created successfully")

	response.Created(c, "section", &newSection, "Section created successfully")
}

func (h *SectionHandler) Update(c *gin.Context) {
	userID := c.GetString("userID") // From auth middleware
	sectionID := c.Param("id")

	// Parse section ID
	id, err := strconv.Atoi(sectionID)
	if err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "UPDATE_SECTION_INVALID_ID",
			"where":     "backend/internal/application/handler/section.go",
			"function":  "Update",
			"userID":    userID,
			"sectionID": sectionID,
			"error":     err.Error(),
		}).Warn("Invalid section ID")
		response.BadRequest(c, "Invalid section ID")
		return
	}

	// Parse request body
	var updateData models.Section
	if err := c.ShouldBindJSON(&updateData); err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "UPDATE_SECTION_BAD_REQUEST",
			"where":     "backend/internal/application/handler/section.go",
			"function":  "Update",
			"userID":    userID,
			"sectionID": id,
			"error":     err.Error(),
		}).Warn("Invalid request data")
		response.BadRequest(c, "Invalid request data")
		return
	}

	// Set the ID and owner
	updateData.ID = uint(id)
	updateData.OwnerID = userID

	// Validate section data
	if err := validator.ValidateSection(&updateData); err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":   "UPDATE_SECTION_VALIDATION_ERROR",
			"where":       "backend/internal/application/handler/section.go",
			"function":    "Update",
			"userID":      userID,
			"sectionID":   id,
			"portfolioID": updateData.PortfolioID,
			"error":       err.Error(),
		}).Warn("Section validation failed")
		response.BadRequest(c, err.Error())
		return
	}

	// Check if section exists and belongs to user
	existing, err := h.repo.GetByID(uint(id))
	if err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "UPDATE_SECTION_NOT_FOUND",
			"where":     "backend/internal/application/handler/section.go",
			"function":  "Update",
			"userID":    userID,
			"sectionID": id,
			"error":     err.Error(),
		}).Warn("Section not found")
		response.NotFound(c, "Section not found")
		return
	}
	if existing.OwnerID != userID {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "UPDATE_SECTION_FORBIDDEN",
			"where":     "backend/internal/application/handler/section.go",
			"function":  "Update",
			"userID":    userID,
			"sectionID": id,
			"ownerID":   existing.OwnerID,
		}).Warn("Access denied")
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
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":   "UPDATE_SECTION_DUPLICATE_CHECK_ERROR",
			"where":       "backend/internal/application/handler/section.go",
			"function":    "Update",
			"userID":      userID,
			"sectionID":   id,
			"portfolioID": updateData.PortfolioID,
			"title":       updateData.Title,
			"error":       err.Error(),
		}).Error("Failed to check for duplicate section")
		response.InternalError(c, "Failed to check for duplicate section")
		return
	}
	if isDuplicate {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":   "UPDATE_SECTION_DUPLICATE_TITLE",
			"where":       "backend/internal/application/handler/section.go",
			"function":    "Update",
			"userID":      userID,
			"sectionID":   id,
			"portfolioID": updateData.PortfolioID,
			"title":       updateData.Title,
		}).Warn("Section with this title already exists in this portfolio")
		response.BadRequest(c, "Section with this title already exists in this portfolio")
		return
	}

	// Update section
	if err := h.repo.Update(&updateData); err != nil {
		// Check if error is due to foreign key constraint (invalid portfolio_id)
		errMsg := err.Error()
		if strings.Contains(errMsg, "fk_portfolios_sections") || strings.Contains(errMsg, "23503") {
			audit.GetErrorLogger().WithFields(logrus.Fields{
				"operation":   "UPDATE_SECTION_FK_CONSTRAINT_ERROR",
				"where":       "backend/internal/application/handler/section.go",
				"function":    "Update",
				"userID":      userID,
				"sectionID":   id,
				"portfolioID": updateData.PortfolioID,
				"error":       err.Error(),
			}).Warn("Portfolio not found")
			response.NotFound(c, "Portfolio not found")
			return
		}
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":   "UPDATE_SECTION_DB_ERROR",
			"where":       "backend/internal/application/handler/section.go",
			"function":    "Update",
			"userID":      userID,
			"sectionID":   id,
			"portfolioID": updateData.PortfolioID,
			"error":       err.Error(),
		}).Error("Failed to update section")
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
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "DELETE_SECTION_INVALID_ID",
			"where":     "backend/internal/application/handler/section.go",
			"function":  "Delete",
			"userID":    userID,
			"sectionID": sectionID,
			"error":     err.Error(),
		}).Warn("Invalid section ID")
		response.BadRequest(c, "Invalid section ID")
		return
	}

	// Get a section to check ownership
	section, err := h.repo.GetByID(uint(id))
	if err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "DELETE_SECTION_NOT_FOUND",
			"where":     "backend/internal/application/handler/section.go",
			"function":  "Delete",
			"userID":    userID,
			"sectionID": id,
			"error":     err.Error(),
		}).Warn("Section not found")
		response.NotFound(c, "Section not found")
		return
	}

	if section.OwnerID != userID {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "DELETE_SECTION_FORBIDDEN",
			"where":     "backend/internal/application/handler/section.go",
			"function":  "Delete",
			"userID":    userID,
			"sectionID": id,
			"ownerID":   section.OwnerID,
		}).Warn("Access denied")
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
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":   "DELETE_SECTION_DB_ERROR",
			"where":       "backend/internal/application/handler/section.go",
			"function":    "Delete",
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
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "UPDATE_SECTION_POSITION_INVALID_ID",
			"where":     "backend/internal/application/handler/section.go",
			"function":  "UpdatePosition",
			"userID":    userID,
			"sectionID": sectionID,
			"error":     err.Error(),
		}).Warn("Invalid section ID")
		response.BadRequest(c, "Invalid section ID")
		return
	}

	// Parse request body
	var req struct {
		Position uint `json:"position" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "UPDATE_SECTION_POSITION_BAD_REQUEST",
			"where":     "backend/internal/application/handler/section.go",
			"function":  "UpdatePosition",
			"userID":    userID,
			"sectionID": id,
			"error":     err.Error(),
		}).Warn("Invalid request data")
		response.BadRequest(c, "Invalid request data")
		return
	}

	// Check if the section exists and belongs to a user
	existing, err := h.repo.GetByID(uint(id))
	if err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "UPDATE_SECTION_POSITION_NOT_FOUND",
			"where":     "backend/internal/application/handler/section.go",
			"function":  "UpdatePosition",
			"userID":    userID,
			"sectionID": id,
			"error":     err.Error(),
		}).Warn("Section not found")
		response.NotFound(c, "Section not found")
		return
	}

	if existing.OwnerID != userID {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "UPDATE_SECTION_POSITION_FORBIDDEN",
			"where":     "backend/internal/application/handler/section.go",
			"function":  "UpdatePosition",
			"userID":    userID,
			"sectionID": id,
			"ownerID":   existing.OwnerID,
		}).Warn("Access denied")
		response.ForbiddenWithDetails(c, "Access denied", map[string]interface{}{
			"resource_type": "section",
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
			"operation": "UPDATE_SECTION_POSITION_DB_ERROR",
			"where":     "backend/internal/application/handler/section.go",
			"function":  "UpdatePosition",
			"userID":    userID,
			"sectionID": id,
			"position":  req.Position,
			"error":     err.Error(),
		}).Error("Failed to update section position")
		response.InternalError(c, "Failed to update section position")
		return
	}

	// Audit log successful update
	audit.GetUpdateLogger().WithFields(logrus.Fields{
		"operation":   "UPDATE_SECTION_POSITION",
		"sectionID":   id,
		"oldPosition": oldPosition,
		"newPosition": req.Position,
		"userID":      userID,
	}).Info("Section position updated successfully")

	response.OK(c, "message", "Section position updated successfully", "Success")
}

// BulkReorder handles reordering multiple sections atomically
func (h *SectionHandler) BulkReorder(c *gin.Context) {
	// Get user ID
	userID := c.GetString("userID") // From auth middleware

	// Parse request
	var req SectionBulkReorderRequest
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

	// Verify all sections belong to user's portfolios
	sectionIDs := make([]uint, len(req.Items))
	for i, item := range req.Items {
		sectionIDs[i] = item.ID
	}

	sections, err := h.repo.GetByIDs(sectionIDs)
	if err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "BULK_REORDER_SECTIONS",
			"userID":    userID,
			"error":     err.Error(),
		}).Error("Failed to fetch sections for bulk reorder")
		response.InternalError(c, "Failed to fetch sections")
		return
	}

	if len(sections) != len(req.Items) {
		response.NotFound(c, "Some sections not found")
		return
	}

	// Verify ownership
	for _, sec := range sections {
		portfolio, err := h.portfolioRepo.GetByID(sec.PortfolioID)
		if err != nil || portfolio.OwnerID != userID {
			response.ForbiddenWithDetails(c, "Access denied", map[string]interface{}{
				"resource_type": "section",
				"resource_id":   sec.ID,
			})
			return
		}
	}

	// Update positions in transaction
	if err := h.repo.BulkUpdatePositions(req.Items); err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "BULK_REORDER_SECTIONS",
			"userID":    userID,
			"itemCount": len(req.Items),
			"error":     err.Error(),
		}).Error("Failed to bulk update section positions")
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
		"operation": "BULK_REORDER_SECTIONS",
		"userID":    userID,
		"itemCount": len(req.Items),
		"items":     itemDetails,
	}).Info("Sections reordered successfully")

	response.OK(c, "message", "Sections reordered successfully", "Success")
}
