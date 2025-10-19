package router

import (
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

func (r *Router) RegisterProjectRoutes(apiGroup *gin.RouterGroup) {
	projects := apiGroup.Group("/projects")
	projects.Use(middleware.AuthMiddleware())
	{
		projects.POST("/", r.projectHandler.Create)
		projects.PUT("/id/:id", r.projectHandler.Update)
		projects.DELETE("/id/:id", r.projectHandler.Delete)
	}
	// Public routes - no auth required
	apiGroup.GET("/projects/id/:id", r.projectHandler.GetByID)
	apiGroup.GET("/projects/category/:categoryId", r.projectHandler.GetByCategory)
	apiGroup.GET("/projects/search/skills", r.projectHandler.GetBySkills)
	apiGroup.GET("/projects/search/client", r.projectHandler.GetByClient)
}
