package router

import (
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/handler"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/metrics"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/repo"
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
