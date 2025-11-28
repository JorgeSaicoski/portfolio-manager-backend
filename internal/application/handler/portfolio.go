package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/models"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/infrastructure/audit"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/infrastructure/metrics"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/infrastructure/repo"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/shared/dto"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/shared/dto/request"
	dtoresponse "github.com/JorgeSaicoski/portfolio-manager/backend/internal/shared/dto/response"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/shared/validator"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
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
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "GET_PORTFOLIOS_BY_USER_DB_ERROR",
			"where":     "backend/internal/application/handler/portfolio.go",
			"function":  "GetByUser",
			"userID":    userID,
			"error":     err.Error(),
		}).Error("Failed to retrieve portfolios")
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to retrieve portfolios",
		})
		return
	}

	c.JSON(http.StatusOK, dto.PaginatedResponse{
		Data:    dtoresponse.ToPortfolioListResponse(portfolios),
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
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":   "UPDATE_PORTFOLIO_INVALID_ID",
			"where":       "backend/internal/application/handler/portfolio.go",
			"function":    "Update",
			"userID":      userID,
			"portfolioID": portfolioID,
			"error":       err.Error(),
		}).Warn("Invalid portfolio ID")
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid portfolio ID",
		})
		return
	}

	// Parse request body
	var req request.UpdatePortfolioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":   "UPDATE_PORTFOLIO_BAD_REQUEST",
			"where":       "backend/internal/application/handler/portfolio.go",
			"function":    "Update",
			"userID":      userID,
			"portfolioID": id,
			"error":       err.Error(),
		}).Warn("Invalid request data")
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

	// Validate portfolio data
	if err := validator.ValidatePortfolio(&updateData); err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":   "UPDATE_PORTFOLIO_VALIDATION_ERROR",
			"where":       "backend/internal/application/handler/portfolio.go",
			"function":    "Update",
			"userID":      userID,
			"portfolioID": id,
			"error":       err.Error(),
		}).Warn("Portfolio validation failed")
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	// Check if portfolio exists and belongs to user
	existing, err := h.repo.GetByIDBasic(uint(id))
	if err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":   "UPDATE_PORTFOLIO_NOT_FOUND",
			"where":       "backend/internal/application/handler/portfolio.go",
			"function":    "Update",
			"userID":      userID,
			"portfolioID": id,
			"error":       err.Error(),
		}).Warn("Portfolio not found")
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error: "Portfolio not found",
		})
		return
	}
	if existing.OwnerID != userID {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":   "UPDATE_PORTFOLIO_FORBIDDEN",
			"where":       "backend/internal/application/handler/portfolio.go",
			"function":    "Update",
			"userID":      userID,
			"portfolioID": id,
			"ownerID":     existing.OwnerID,
		}).Warn("Access denied")
		c.JSON(http.StatusForbidden, dto.ErrorResponse{
			Error: "Access denied",
		})
		return
	}

	// Check for duplicate title
	isDuplicate, err := h.repo.CheckDuplicate(updateData.Title, updateData.OwnerID, updateData.ID)
	if err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":   "UPDATE_PORTFOLIO_DUPLICATE_CHECK_ERROR",
			"where":       "backend/internal/application/handler/portfolio.go",
			"function":    "Update",
			"userID":      userID,
			"portfolioID": id,
			"title":       updateData.Title,
			"error":       err.Error(),
		}).Error("Failed to check for duplicate portfolio")
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to check for duplicate portfolio",
		})
		return
	}
	if isDuplicate {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":   "UPDATE_PORTFOLIO_DUPLICATE_TITLE",
			"where":       "backend/internal/application/handler/portfolio.go",
			"function":    "Update",
			"userID":      userID,
			"portfolioID": id,
			"title":       updateData.Title,
		}).Warn("Portfolio with this title already exists")
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Portfolio with this title already exists",
		})
		return
	}

	// Update portfolio
	if err := h.repo.Update(&updateData); err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":   "UPDATE_PORTFOLIO_DB_ERROR",
			"where":       "backend/internal/application/handler/portfolio.go",
			"function":    "Update",
			"userID":      userID,
			"portfolioID": id,
			"error":       err.Error(),
		}).Error("Failed to update portfolio")
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to update portfolio",
		})
		return
	}

	// Audit log for update operation
	audit.GetUpdateLogger().WithFields(logrus.Fields{
		"operation":   "UPDATE_PORTFOLIO",
		"portfolioID": updateData.ID,
		"title":       updateData.Title,
		"userID":      userID,
	}).Info("Portfolio updated successfully")

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Message: "Portfolio updated successfully",
		Data:    dtoresponse.ToPortfolioResponse(&updateData),
	})
}

func (h *PortfolioHandler) Create(c *gin.Context) {
	userID := c.GetString("userID") // From auth middleware

	logrus.WithFields(logrus.Fields{
		"userID": userID,
		"path":   c.Request.URL.Path,
	}).Info("Portfolio creation request received")

	// Parse request body
	var req request.CreatePortfolioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Log the raw request body for debugging
		bodyBytes, _ := c.GetRawData()
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "CREATE_PORTFOLIO_BAD_REQUEST",
			"where":     "backend/internal/application/handler/portfolio.go",
			"function":  "Create",
			"error":     err.Error(),
			"raw_body":  string(bodyBytes),
			"userID":    userID,
		}).Warn("Failed to parse portfolio creation request")

		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid request data: " + err.Error(),
		})
		return
	}

	// Log the parsed request
	reqJSON, _ := json.Marshal(req)
	logrus.WithFields(logrus.Fields{
		"userID":  userID,
		"request": string(reqJSON),
	}).Info("Parsed portfolio creation request")

	// Convert DTO to model
	newPortfolio := models.Portfolio{
		Title:       req.Title,
		Description: req.Description,
		OwnerID:     userID,
	}

	logrus.WithFields(logrus.Fields{
		"title":       newPortfolio.Title,
		"description": newPortfolio.Description,
		"ownerID":     newPortfolio.OwnerID,
	}).Info("Portfolio model created, validating...")

	// Validate portfolio data
	if err := validator.ValidatePortfolio(&newPortfolio); err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "CREATE_PORTFOLIO_VALIDATION_ERROR",
			"where":     "backend/internal/application/handler/portfolio.go",
			"function":  "Create",
			"error":     err.Error(),
			"userID":    userID,
			"title":     newPortfolio.Title,
		}).Warn("Portfolio validation failed")

		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	logrus.Info("Portfolio validation passed, checking for duplicates...")

	// Check for duplicate title
	isDuplicate, err := h.repo.CheckDuplicate(newPortfolio.Title, newPortfolio.OwnerID, 0)
	if err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "CREATE_PORTFOLIO_DUPLICATE_CHECK_ERROR",
			"where":     "backend/internal/application/handler/portfolio.go",
			"function":  "Create",
			"error":     err.Error(),
			"userID":    userID,
			"title":     newPortfolio.Title,
		}).Error("Failed to check for duplicate portfolio")

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to check for duplicate portfolio",
		})
		return
	}
	if isDuplicate {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "CREATE_PORTFOLIO_DUPLICATE_TITLE",
			"where":     "backend/internal/application/handler/portfolio.go",
			"function":  "Create",
			"userID":    userID,
			"title":     newPortfolio.Title,
		}).Warn("Duplicate portfolio title detected")

		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Portfolio with this title already exists",
		})
		return
	}

	logrus.Info("No duplicates found, creating portfolio...")

	// Create portfolio
	if err := h.repo.Create(&newPortfolio); err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "CREATE_PORTFOLIO_DB_ERROR",
			"where":     "backend/internal/application/handler/portfolio.go",
			"function":  "Create",
			"error":     err.Error(),
			"userID":    userID,
			"title":     newPortfolio.Title,
		}).Error("Failed to create portfolio in database")

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to create portfolio",
		})
		return
	}

	// Audit log for create operation
	audit.GetCreateLogger().WithFields(logrus.Fields{
		"operation":   "CREATE_PORTFOLIO",
		"portfolioID": newPortfolio.ID,
		"title":       newPortfolio.Title,
		"userID":      userID,
	}).Info("Portfolio created successfully")

	c.JSON(http.StatusCreated, dto.SuccessResponse{
		Message: "Portfolio created successfully",
		Data:    dtoresponse.ToPortfolioResponse(&newPortfolio),
	})
}

func (h *PortfolioHandler) Delete(c *gin.Context) {
	userID := c.GetString("userID")
	portfolioID := c.Param("id")

	id, err := strconv.Atoi(portfolioID)
	if err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":   "DELETE_PORTFOLIO_INVALID_ID",
			"where":       "backend/internal/application/handler/portfolio.go",
			"function":    "Delete",
			"userID":      userID,
			"portfolioID": portfolioID,
			"error":       err.Error(),
		}).Warn("Invalid portfolio ID")
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid portfolio ID",
		})
		return
	}

	// Use basic method - only fetch id and owner_id for authorization
	portfolio, err := h.repo.GetByIDBasic(uint(id))
	if err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":   "DELETE_PORTFOLIO_NOT_FOUND",
			"where":       "backend/internal/application/handler/portfolio.go",
			"function":    "Delete",
			"userID":      userID,
			"portfolioID": id,
			"error":       err.Error(),
		}).Warn("Portfolio not found")
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error: "Portfolio not found",
		})
		return
	}

	if portfolio.OwnerID != userID {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":   "DELETE_PORTFOLIO_FORBIDDEN",
			"where":       "backend/internal/application/handler/portfolio.go",
			"function":    "Delete",
			"userID":      userID,
			"portfolioID": id,
			"ownerID":     portfolio.OwnerID,
		}).Warn("Access denied")
		c.JSON(http.StatusForbidden, dto.ErrorResponse{
			Error: "Access denied",
		})
		return
	}

	if err := h.repo.Delete(uint(id)); err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":   "DELETE_PORTFOLIO_DB_ERROR",
			"where":       "backend/internal/application/handler/portfolio.go",
			"function":    "Delete",
			"portfolioID": id,
			"userID":      userID,
			"error":       err.Error(),
		}).Error("Failed to delete portfolio")

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to delete portfolio",
		})
		return
	}

	// Audit log for delete operation
	audit.GetDeleteLogger().WithFields(logrus.Fields{
		"operation":   "DELETE_PORTFOLIO",
		"portfolioID": id,
		"userID":      userID,
		"title":       portfolio.Title,
		"cascade":     "categories, sections, projects",
	}).Info("Portfolio deleted successfully (CASCADE: all related categories, sections, and projects)")

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Message: "Portfolio deleted successfully",
	})
}

func (h *PortfolioHandler) GetByIDPublic(c *gin.Context) {
	portfolioID := c.Param("id")

	// Parse portfolio ID
	id, err := strconv.Atoi(portfolioID)
	if err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":   "GET_PORTFOLIO_BY_ID_PUBLIC_INVALID_ID",
			"where":       "backend/internal/application/handler/portfolio.go",
			"function":    "GetByIDPublic",
			"portfolioID": portfolioID,
			"error":       err.Error(),
		}).Warn("Invalid portfolio ID")
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid portfolio ID",
		})
		return
	}

	// Get complete portfolio with relationships
	portfolio, err := h.repo.GetByIDWithRelations(uint(id))
	if err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":   "GET_PORTFOLIO_BY_ID_PUBLIC_NOT_FOUND",
			"where":       "backend/internal/application/handler/portfolio.go",
			"function":    "GetByIDPublic",
			"portfolioID": id,
			"error":       err.Error(),
		}).Warn("Portfolio not found")
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error: "Portfolio not found",
		})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Message: "Success",
		Data:    dtoresponse.ToPortfolioDetailResponse(portfolio),
	})
}
