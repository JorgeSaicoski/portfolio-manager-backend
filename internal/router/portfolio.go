package router

import (
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

func (r *Router) RegisterPortfolioRoutes(apiGroup *gin.RouterGroup) {
	portfolios := apiGroup.Group("/portfolios")
	portfolios.Use(middleware.AuthMiddleware())
	{
		portfolios.GET("/own", r.portfolioHandler.GetByUser)
		portfolios.POST("/own", r.portfolioHandler.Create)
		portfolios.PUT("/own/id/:id", r.portfolioHandler.Update)
		portfolios.DELETE("/own/id/:id", r.portfolioHandler.Delete)
	}
	// Public routes - no auth required
	apiGroup.GET("/portfolios/id/:id", r.portfolioHandler.GetByIDPublic)
}
