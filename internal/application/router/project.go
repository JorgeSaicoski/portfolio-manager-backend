package router

import (
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/shared/middleware"
	"github.com/gin-gonic/gin"
)

func (r *Router) RegisterProjectRoutes(apiGroup *gin.RouterGroup) {
	projects := apiGroup.Group("/projects")

	// Protected routes - require authentication
	protected := projects.Group("/own")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.GET("", r.projectHandler.GetByUser)
		protected.POST("", r.projectHandler.Create)
		protected.GET("/:id", r.projectHandler.GetByID)
		protected.PUT("/:id", r.projectHandler.Update)
		protected.DELETE("/:id", r.projectHandler.Delete)
	}

	// Public routes - no auth required
	projects.GET("/public/:id", r.projectHandler.GetByIDPublic)
	projects.GET("/category/:categoryId", r.projectHandler.GetByCategory)
	projects.GET("/search/skills", r.projectHandler.GetBySkills)
	projects.GET("/search/client", r.projectHandler.GetByClient)
}
