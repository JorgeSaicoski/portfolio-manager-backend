package router

import (
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

func (r *Router) RegisterPortfolioRoutes(apiGroup *gin.RouterGroup) {
	portfolios := apiGroup.Group("/portfolios")

	// Protected routes - require authentication
	protected := portfolios.Group("/own")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.GET("", r.portfolioHandler.GetByUser)
		protected.POST("", r.portfolioHandler.Create)
		protected.GET("/:id", r.portfolioHandler.GetByIDPublic) // Authenticated users can also view
		protected.PUT("/:id", r.portfolioHandler.Update)
		protected.DELETE("/:id", r.portfolioHandler.Delete)
	}

	// Public routes - no auth required
	portfolios.GET("/id/:id", r.portfolioHandler.GetByIDPublic)
	portfolios.GET("/public/:id", r.portfolioHandler.GetByIDPublic)
	portfolios.GET("/public/:id/categories", r.categoryHandler.GetByPortfolio)
}
