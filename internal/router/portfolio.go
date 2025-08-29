package router

import (
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

func (r *Router) RegisterPortfolioRoutes(engine *gin.Engine) {
	portfolios := engine.Group("/portfolios")
	portfolios.Use(middleware.AuthMiddleware())
	{
		portfolios.GET("/user", r.portfolioHandler.GetByUser)
	}
}
