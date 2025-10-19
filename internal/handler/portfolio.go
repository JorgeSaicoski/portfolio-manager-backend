package handler

import (
	"net/http"
	"strconv"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/dto"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/dto/request"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/dto/response"
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

	// Parse pagination parameters
	var pagination dto.PaginationQuery
	if err := c.ShouldBindQuery(&pagination); err != nil {
		pagination = dto.PaginationQuery{Page: 1, Limit: 10}
	}

	page, limit := pagination.GetPageAndLimit()
	offset := pagination.GetOffset()

	portfolios, err := h.repo.GetByOwnerIDBasic(userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to retrieve portfolios",
		})
		return
	}

	c.JSON(http.StatusOK, dto.PaginatedResponse{
		Data:    response.ToPortfolioListResponse(portfolios),
		Page:    page,
		Limit:   limit,
		Message: "Success",
	})
}
func (h *PortfolioHandler) Update(c *gin.Context) {
	userID := c.GetString("userID") // From auth middleware
	portfolioID := c.Param("id")

	// Parse portfolio ID
	id, err := strconv.Atoi(portfolioID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid portfolio ID",
		})
		return
	}

	// Parse request body
	var req request.UpdatePortfolioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid request data",
		})
		return
	}

	// Convert DTO to model
	updateData := models.Portfolio{
		Title:       req.Title,
		Description: req.Description,
		OwnerID:     userID,
	}
	updateData.ID = uint(id)

	// Update portfolio
	if err := h.repo.Update(&updateData); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to update portfolio",
		})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Message: "Portfolio updated successfully",
		Data:    response.ToPortfolioResponse(&updateData),
	})
}

func (h *PortfolioHandler) Create(c *gin.Context) {
	userID := c.GetString("userID") // From auth middleware

	// Parse request body
	var req request.CreatePortfolioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid request data",
		})
		return
	}

	// Convert DTO to model
	newPortfolio := models.Portfolio{
		Title:       req.Title,
		Description: req.Description,
		OwnerID:     userID,
	}

	// Create portfolio
	if err := h.repo.Create(&newPortfolio); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to create portfolio",
		})
		return
	}

	c.JSON(http.StatusCreated, dto.SuccessResponse{
		Message: "Portfolio created successfully",
		Data:    response.ToPortfolioResponse(&newPortfolio),
	})
}

func (h *PortfolioHandler) Delete(c *gin.Context) {
	userID := c.GetString("userID")
	portfolioID := c.Param("id")

	id, err := strconv.Atoi(portfolioID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid portfolio ID",
		})
		return
	}

	// Use basic method - only fetch id and owner_id for authorization
	portfolio, err := h.repo.GetByIDBasic(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error: "Portfolio not found",
		})
		return
	}

	if portfolio.OwnerID != userID {
		c.JSON(http.StatusForbidden, dto.ErrorResponse{
			Error: "Access denied",
		})
		return
	}

	if err := h.repo.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to delete portfolio",
		})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Message: "Portfolio deleted successfully",
	})
}

func (h *PortfolioHandler) GetByIDPublic(c *gin.Context) {
	portfolioID := c.Param("id")

	// Parse portfolio ID
	id, err := strconv.Atoi(portfolioID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid portfolio ID",
		})
		return
	}

	// Get complete portfolio with relationships
	portfolio, err := h.repo.GetByIDWithRelations(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error: "Portfolio not found",
		})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Message: "Success",
		Data:    response.ToPortfolioDetailResponse(portfolio),
	})
}
