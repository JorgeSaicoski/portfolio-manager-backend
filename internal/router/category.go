package router

import (
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

func (r *Router) RegisterCategoryRoutes(apiGroup *gin.RouterGroup) {
	categories := apiGroup.Group("/categories")
	categories.Use(middleware.AuthMiddleware())
	{
		categories.GET("/own", r.categoryHandler.GetByUser)
		categories.POST("/own", r.categoryHandler.Create)
		categories.PUT("/own/id/:id", r.categoryHandler.Update)
		categories.DELETE("/own/id/:id", r.categoryHandler.Delete)
	}
	// Public routes - no auth required
	apiGroup.GET("/categories/id/:id", r.categoryHandler.GetByIDPublic)
}
