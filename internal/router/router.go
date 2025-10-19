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
	categoryHandler  *handler.CategoryHandler
	projectHandler   *handler.ProjectHandler
	sectionHandler   *handler.SectionHandler
	metrics          *metrics.Collector
}

func NewRouter(db *gorm.DB, metrics *metrics.Collector) *Router {
	portfolioRepo := repo.NewPortfolioRepository(db)
	portfolioHandler := handler.NewPortfolioHandler(portfolioRepo, metrics)

	categoryRepo := repo.NewCategoryRepository(db)
	categoryHandler := handler.NewCategoryHandler(categoryRepo, metrics)

	projectRepo := repo.NewProjectRepository(db)
	projectHandler := handler.NewProjectHandler(projectRepo, metrics)

	sectionRepo := repo.NewSectionRepository(db)
	sectionHandler := handler.NewSectionHandler(sectionRepo, metrics)

	return &Router{
		db:               db,
		portfolioHandler: portfolioHandler,
		categoryHandler:  categoryHandler,
		projectHandler:   projectHandler,
		sectionHandler:   sectionHandler,
		metrics:          metrics,
	}
}
