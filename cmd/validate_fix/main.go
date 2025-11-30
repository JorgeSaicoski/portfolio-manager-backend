package main

import (
	"fmt"
	"log"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/handler"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/infrastructure/db"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/infrastructure/metrics"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/infrastructure/repo"
)

// Simple validation script to check if UserHandler compiles correctly
func main() {
	// Initialize database
	database, err := db.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize metrics
	metricsCollector := metrics.NewCollector()

	// Initialize repositories
	portfolioRepo := repo.NewPortfolioRepository(database.DB)
	categoryRepo := repo.NewCategoryRepository(database.DB)
	sectionRepo := repo.NewSectionRepository(database.DB)
	projectRepo := repo.NewProjectRepository(database.DB)
	imageRepo := repo.NewImageRepository(database.DB)
	sectionContentRepo := repo.NewSectionContentRepository(database.DB)

	// Initialize handler - this will fail to compile if signature is wrong
	userHandler := handler.NewUserHandler(
		portfolioRepo,
		categoryRepo,
		sectionRepo,
		projectRepo,
		imageRepo,
		sectionContentRepo,
	)

	if userHandler == nil {
		log.Fatal("Failed to create user handler")
	}

	fmt.Println("✅ UserHandler compiled successfully with all dependencies")
	fmt.Println("✅ All repository interfaces are correctly implemented")
	fmt.Println("✅ Code changes are syntactically correct")
}
