package handler

import (
	"strconv"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/metrics"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/models"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/repo"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/response"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/validator"
	"github.com/gin-gonic/gin"
)

type SectionHandler struct {
	repo    repo.SectionRepository
	metrics *metrics.Collector
}

func NewSectionHandler(repo repo.SectionRepository, metrics *metrics.Collector) *SectionHandler {
	return &SectionHandler{
		repo:    repo,
		metrics: metrics,
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

	sections, err := h.repo.GetByOwnerIDBasic(userID, limit, offset)
	if err != nil {
		response.InternalError(c, "Failed to retrieve sections")
		return
	}

	response.SuccessWithPagination(c, 200, "sections", sections, page, limit)
}

func (h *SectionHandler) GetByPortfolio(c *gin.Context) {
	portfolioID := c.Param("portfolioId")

	sections, err := h.repo.GetByPortfolioIDBasic(portfolioID)
	if err != nil {
		response.InternalError(c, "Failed to retrieve sections")
		return
	}

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

	section, err := h.repo.GetByIDWithRelations(uint(id))
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

	// Validate section data
	if err := validator.ValidateSection(&newSection); err != nil {
		response.BadRequest(c, err.Error())
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
	section, err := h.repo.GetByIDWithRelations(uint(id))
	if err != nil {
		response.NotFound(c, "Section not found")
		return
	}

	if section.OwnerID != userID {
		response.Forbidden(c, "Access denied")
		return
	}

	if err := h.repo.Delete(uint(id)); err != nil {
		response.InternalError(c, "Failed to delete section")
		return
	}

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

	// Check if section exists and belongs to user
	existing, err := h.repo.GetByIDWithRelations(uint(id))
	if err != nil {
		response.NotFound(c, "Section not found")
		return
	}

	if existing.OwnerID != userID {
		response.Forbidden(c, "Access denied")
		return
	}

	// Update position
	if err := h.repo.UpdatePosition(uint(id), req.Position); err != nil {
		response.InternalError(c, "Failed to update section position")
		return
	}

	response.OK(c, "message", "Section position updated successfully", "Success")
}
