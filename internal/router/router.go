package router

import (
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/handler"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Router struct {
	db               *gorm.DB
	portfolioHandler *handler.PortfolioHandler
}

func NewRouter(db *gorm.DB) *Router {
	return &Router{
		db:               db,
		portfolioHandler: handler.NewPortfolioHandler(db),
	}
}

func (r *Router) SetupRoutes() *gin.Engine {
	engine := gin.Default()

	r.RegisterPortfolioRoutes(engine)

	return engine
}
