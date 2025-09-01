package router

import (
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

func (r *Router) RegisterPortfolioRoutes(engine *gin.Engine) {
	portfolios := engine.Group("/portfolios")
	portfolios.Use(middleware.AuthMiddleware())
	{
		portfolios.GET("/own", r.portfolioHandler.GetByUser)
		portfolios.POST("/own", r.portfolioHandler.Create)
		portfolios.PUT("/own/id/:id", r.portfolioHandler.Update)
		portfolios.DELETE("/own/id/:id", r.portfolioHandler.Delete)
	}
	// will create the public routes here.
	// get portfolios/id/:id
	// so the user will receive the portfolio that he is visiting
}
