package handler

import (
	"net/http"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/metrics"
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
