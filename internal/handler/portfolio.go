package handler

import (
	"net/http"
	"strconv"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/metrics"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/models"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/repo"
	"github.com/gin-gonic/gin"
)

type PortfolioHandler struct {
	repo    repo.PortfolioRepository
	metrics *metrics.Collector
}

func NewPortfolioHandler(repo repo.PortfolioRepository, metrics *metrics.Collector) *PortfolioHandler {
	return &PortfolioHandler{
		repo:    repo,
		metrics: metrics,
	}
}

func (h *PortfolioHandler) GetByUser(c *gin.Context) {
	userID := c.GetString("userID") // From auth middleware

	portfolios, err := h.repo.GetByOwnerID(userID, 10, 0) // limit: 10, offset: 0
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve portfolios",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"portfolios": portfolios,
		"message":    "Success",
	})
}

func (h *PortfolioHandler) Update(c *gin.Context) {
	userID := c.GetString("userID") // From auth middleware
	portfolioID := c.Param("id")

	// Parse portfolio ID
	id, err := strconv.Atoi(portfolioID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid portfolio ID",
		})
		return
	}

	// Parse request body
	var updateData models.Portfolio
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request data",
		})
		return
	}

	// Set the ID and owner
	updateData.ID = uint(id)
	updateData.OwnerID = userID

	// Update portfolio
	if err := h.repo.Update(&updateData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update portfolio",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"portfolio": &updateData,
		"message":   "Portfolio updated successfully",
	})
}

func (h *PortfolioHandler) Create(c *gin.Context) {
	userID := c.GetString("userID") // From auth middleware

	// Parse request body
	var newPortfolio models.Portfolio
	if err := c.ShouldBindJSON(&newPortfolio); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request data",
		})
		return
	}

	// Set the owner
	newPortfolio.OwnerID = userID

	// Create portfolio
	if err := h.repo.Create(&newPortfolio); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create portfolio",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"portfolio": &newPortfolio,
		"message":   "Portfolio created successfully",
	})
}

func (h *PortfolioHandler) Delete(c *gin.Context) {
	userID := c.GetString("userID") // From auth middleware
	portfolioID := c.Param("id")

	// Parse portfolio ID
	id, err := strconv.Atoi(portfolioID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid portfolio ID",
		})
		return
	}

	// Check if portfolio exists and belongs to user
	portfolio, err := h.repo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Portfolio not found",
		})
		return
	}

	// Check ownership
	if portfolio.OwnerID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Access denied",
		})
		return
	}

	// Delete portfolio
	if err := h.repo.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete portfolio",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Portfolio deleted successfully",
	})
}
