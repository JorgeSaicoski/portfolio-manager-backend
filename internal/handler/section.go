package handler

import (
	"net/http"
	"strconv"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/metrics"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/models"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/repo"
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

func (h *SectionHandler) GetByPortfolio(c *gin.Context) {
	portfolioID := c.Param("portfolioId")

	sections, err := h.repo.GetByPortfolioIDBasic(portfolioID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve sections",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sections": sections,
		"message":  "Success",
	})
}

func (h *SectionHandler) GetByID(c *gin.Context) {
	sectionID := c.Param("id")

	// Parse section ID
	id, err := strconv.Atoi(sectionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid section ID",
		})
		return
	}

	section, err := h.repo.GetByIDWithRelations(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Section not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"section": section,
		"message": "Success",
	})
}

func (h *SectionHandler) GetByType(c *gin.Context) {
	sectionType := c.Query("type")
	if sectionType == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Section type is required",
		})
		return
	}

	sections, err := h.repo.GetByType(sectionType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve sections",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sections": sections,
		"message":  "Success",
	})
}

func (h *SectionHandler) Create(c *gin.Context) {
	userID := c.GetString("userID") // From auth middleware

	// Parse request body
	var newSection models.Section
	if err := c.ShouldBindJSON(&newSection); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request data",
		})
		return
	}

	// Set the owner
	newSection.OwnerID = userID

	// Create a section
	if err := h.repo.Create(&newSection); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create section",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"section": &newSection,
		"message": "Section created successfully",
	})
}

func (h *SectionHandler) Update(c *gin.Context) {
	userID := c.GetString("userID") // From auth middleware
	sectionID := c.Param("id")

	// Parse section ID
	id, err := strconv.Atoi(sectionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid section ID",
		})
		return
	}

	// Parse request body
	var updateData models.Section
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request data",
		})
		return
	}

	// Set the ID and owner
	updateData.ID = uint(id)
	updateData.OwnerID = userID

	// Update section
	if err := h.repo.Update(&updateData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update section",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"section": &updateData,
		"message": "Section updated successfully",
	})
}

func (h *SectionHandler) Delete(c *gin.Context) {
	userID := c.GetString("userID")
	sectionID := c.Param("id")

	id, err := strconv.Atoi(sectionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid section ID",
		})
		return
	}

	// Get a section to check ownership
	section, err := h.repo.GetByIDWithRelations(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Section not found",
		})
		return
	}

	if section.OwnerID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Access denied",
		})
		return
	}

	if err := h.repo.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete section",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Section deleted successfully",
	})
}
