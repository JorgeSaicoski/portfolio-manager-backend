package router

import (
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/shared/middleware"
	"github.com/gin-gonic/gin"
)

func (r *Router) RegisterCategoryRoutes(apiGroup *gin.RouterGroup) {
	categories := apiGroup.Group("/categories")

	// Protected routes - require authentication
	protected := categories.Group("/own")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.GET("", r.categoryHandler.GetByUser)
		protected.POST("", r.categoryHandler.Create)
		protected.GET("/:id", r.categoryHandler.GetByIDPublic) // Authenticated users can also view
		protected.PUT("/:id", r.categoryHandler.Update)
		protected.PUT("/:id/position", r.categoryHandler.UpdatePosition)
		protected.DELETE("/:id", r.categoryHandler.Delete)
	}

	// Public routes - no auth required
	categories.GET("/id/:id", r.categoryHandler.GetByIDPublic)
	categories.GET("/public/:id", r.categoryHandler.GetByIDPublic)
	categories.GET("/public/:id/projects", r.projectHandler.GetByCategory)
}
