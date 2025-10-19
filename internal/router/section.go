package router

import (
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

func (r *Router) RegisterSectionRoutes(apiGroup *gin.RouterGroup) {
	sections := apiGroup.Group("/sections")
	sections.Use(middleware.AuthMiddleware())
	{
		sections.POST("/", r.sectionHandler.Create)
		sections.PUT("/id/:id", r.sectionHandler.Update)
		sections.DELETE("/id/:id", r.sectionHandler.Delete)
	}
	// Public routes - no auth required
	apiGroup.GET("/sections/id/:id", r.sectionHandler.GetByID)
	apiGroup.GET("/sections/portfolio/:portfolioId", r.sectionHandler.GetByPortfolio)
	apiGroup.GET("/sections/type", r.sectionHandler.GetByType)
}
