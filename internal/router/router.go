package router

import (
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/handler"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/metrics"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/repo"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Router struct {
	db               *gorm.DB
	portfolioHandler *handler.PortfolioHandler
	metrics          *metrics.Collector
}

func NewRouter(db *gorm.DB, metrics *metrics.Collector) *Router {
	portfolioRepo := repo.NewPortfolioRepository(db)

	portfolioHandler := handler.NewPortfolioHandler(portfolioRepo, metrics)

	return &Router{
		db:               db,
		portfolioHandler: portfolioHandler,
		metrics:          metrics,
	}
}

func (r *Router) SetupRoutes() *gin.Engine {
	engine := gin.Default()

	r.RegisterPortfolioRoutes(engine)

	return engine
}
