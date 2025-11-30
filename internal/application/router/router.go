package router

import (
	handler2 "github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/handler"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/infrastructure/metrics"
	repo2 "github.com/JorgeSaicoski/portfolio-manager/backend/internal/infrastructure/repo"
	"gorm.io/gorm"
)

type Router struct {
	db                    *gorm.DB
	portfolioHandler      *handler2.PortfolioHandler
	categoryHandler       *handler2.CategoryHandler
	projectHandler        *handler2.ProjectHandler
	sectionHandler        *handler2.SectionHandler
	sectionContentHandler *handler2.SectionContentHandler
	imageHandler          *handler2.ImageHandler
	userHandler           *handler2.UserHandler
	imageRepo             repo2.ImageRepository
	metrics               *metrics.Collector
}

func NewRouter(db *gorm.DB, metrics *metrics.Collector) *Router {
	portfolioRepo := repo2.NewPortfolioRepository(db)
	portfolioHandler := handler2.NewPortfolioHandler(portfolioRepo, metrics)

	categoryRepo := repo2.NewCategoryRepository(db)
	categoryHandler := handler2.NewCategoryHandler(categoryRepo, portfolioRepo, metrics)

	projectRepo := repo2.NewProjectRepository(db)
	projectHandler := handler2.NewProjectHandler(projectRepo, categoryRepo, portfolioRepo, metrics)

	sectionRepo := repo2.NewSectionRepository(db)
	sectionHandler := handler2.NewSectionHandler(sectionRepo, portfolioRepo, metrics)

	imageRepo := repo2.NewImageRepository(db)
	imageHandler := handler2.NewImageHandler(imageRepo, metrics)

	sectionContentRepo := repo2.NewSectionContentRepository(db)
	sectionContentHandler := handler2.NewSectionContentHandler(sectionContentRepo, sectionRepo, portfolioRepo, imageRepo, metrics)

	userHandler := handler2.NewUserHandler(
		portfolioRepo,
		categoryRepo,
		sectionRepo,
		projectRepo,
		imageRepo,
		sectionContentRepo,
	)

	return &Router{
		db:                    db,
		portfolioHandler:      portfolioHandler,
		categoryHandler:       categoryHandler,
		projectHandler:        projectHandler,
		sectionHandler:        sectionHandler,
		sectionContentHandler: sectionContentHandler,
		imageHandler:          imageHandler,
		userHandler:           userHandler,
		imageRepo:             imageRepo,
		metrics:               metrics,
	}
}
