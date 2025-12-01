package router

import (
	"github.com/gin-gonic/gin"
	middleware2 "github.com/JorgeSaicoski/portfolio-manager/backend/internal/shared/middleware"
)

func (r *Router) RegisterUserRoutes(apiGroup *gin.RouterGroup) {
	// User management routes - require authentication
	users := apiGroup.Group("/users")
	users.Use(middleware2.AuthMiddleware())
	{
		// Get data summary for authenticated user
		users.GET("/me/summary", r.userHandler.GetUserDataSummary)

		// Delete all data for authenticated user (GDPR compliance)
		users.DELETE("/me/data", r.userHandler.CleanupUserData)
	}
}
