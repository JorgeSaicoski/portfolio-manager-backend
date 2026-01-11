package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/usecases/user"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/interfaces/dto/request"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/interfaces/dto/response"
	"github.com/JorgeSaicoski/portfolio-manager/backend/pkg/pkgerrors"
)

// UserController handles HTTP requests for user operations
type UserController struct {
	getCurrentUserUC *user.GetCurrentUserUseCase
	updateUserUC     *user.UpdateCurrentUserUseCase
}

// NewUserController creates a new user controller instance
func NewUserController(
	getCurrentUserUC *user.GetCurrentUserUseCase,
	updateUserUC *user.UpdateCurrentUserUseCase,
) *UserController {
	return &UserController{
		getCurrentUserUC: getCurrentUserUC,
		updateUserUC:     updateUserUC,
	}
}

// GetMe handles GET /api/users/me
func (ctrl *UserController) GetMe(c *gin.Context) {
	// Extract userID from context (set by auth middleware)
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Error: "unauthorized"})
		return
	}

	// Execute use case
	userDTO, err := ctrl.getCurrentUserUC.Execute(c.Request.Context(), userID)
	if err != nil {
		status := pkgerrors.ToHTTPStatus(err)
		c.JSON(status, response.ErrorResponse{Error: err.Error()})
		return
	}

	// Map to HTTP response DTO
	resp := response.UserResponse{
		ID:        userDTO.ID,
		Email:     userDTO.Email,
		Name:      userDTO.Name,
		CreatedAt: userDTO.CreatedAt,
		UpdatedAt: userDTO.UpdatedAt,
	}

	// Return HTTP response
	c.JSON(http.StatusOK, response.DataResponse{
		Data:    resp,
		Message: "Success",
	})
}

// UpdateMe handles PUT /api/users/me
func (ctrl *UserController) UpdateMe(c *gin.Context) {
	// Extract userID from context (set by auth middleware)
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Error: "unauthorized"})
		return
	}

	// Bind and validate HTTP request DTO
	var req request.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: err.Error()})
		return
	}

	// Map to use case input
	input := user.UpdateCurrentUserInput{
		UserID: userID,
		Name:   req.Name,
	}

	// Execute use case
	userDTO, err := ctrl.updateUserUC.Execute(c.Request.Context(), input)
	if err != nil {
		status := pkgerrors.ToHTTPStatus(err)
		c.JSON(status, response.ErrorResponse{Error: err.Error()})
		return
	}

	// Map to HTTP response DTO
	resp := response.UserResponse{
		ID:        userDTO.ID,
		Email:     userDTO.Email,
		Name:      userDTO.Name,
		CreatedAt: userDTO.CreatedAt,
		UpdatedAt: userDTO.UpdatedAt,
	}

	// Return HTTP response
	c.JSON(http.StatusOK, response.DataResponse{
		Data:    resp,
		Message: "Success",
	})
}
