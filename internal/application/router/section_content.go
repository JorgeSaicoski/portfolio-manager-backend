package router

import (
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/shared/middleware"
	"github.com/gin-gonic/gin"
)

func (r *Router) RegisterSectionContentRoutes(apiGroup *gin.RouterGroup) {
	sectionContents := apiGroup.Group("/section-contents")

	// Protected routes - require authentication
	protected := sectionContents.Group("/own")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.POST("", r.sectionContentHandler.Create)
		protected.PUT("/:id", r.sectionContentHandler.Update)
		protected.PATCH("/:id/order", r.sectionContentHandler.UpdateOrder)
		protected.DELETE("/:id", r.sectionContentHandler.Delete)
	}

	// Public routes - no auth required
	sectionContents.GET("/:id", r.sectionContentHandler.GetByID)

	// Public route for getting all contents of a section
	apiGroup.GET("/sections/:sectionId/contents", r.sectionContentHandler.GetBySectionID)
}
