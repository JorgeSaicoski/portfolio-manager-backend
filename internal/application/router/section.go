package router

import (
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/shared/middleware"
	"github.com/gin-gonic/gin"
)

func (r *Router) RegisterSectionRoutes(apiGroup *gin.RouterGroup) {
	sections := apiGroup.Group("/sections")

	// Protected routes - require authentication
	protected := sections.Group("/own")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.GET("", r.sectionHandler.GetByUser)
		protected.POST("", r.sectionHandler.Create)
		protected.GET("/:id", r.sectionHandler.GetByID)
		protected.PUT("/:id", r.sectionHandler.Update)
		protected.PUT("/:id/position", r.sectionHandler.UpdatePosition)
		protected.DELETE("/:id", r.sectionHandler.Delete)
	}

	// Public routes - no auth required
	sections.GET("/public/:id", r.sectionHandler.GetByID)
	sections.GET("/portfolio/:id", r.sectionHandler.GetByPortfolio)
	sections.GET("/type", r.sectionHandler.GetByType)
}
