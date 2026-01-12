package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/usecases/category"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/usecases/portfolio"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/usecases/project"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/usecases/section"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/usecases/section_content"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/usecases/user"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/infrastructure/logging"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/infrastructure/postgres/entities"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/infrastructure/postgres/repositories"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/infrastructure/prometheus"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/interfaces/controllers"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/interfaces/middleware"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize database
	db, err := initDatabase()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Run migrations
	if err := runMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// 1. Create Repositories (inject DB)
	userRepo := repositories.NewUserRepository(db)
	portfolioRepo := repositories.NewPortfolioRepository(db)
	categoryRepo := repositories.NewCategoryRepository(db)
	sectionRepo := repositories.NewSectionRepository(db)
	projectRepo := repositories.NewProjectRepository(db)
	sectionContentRepo := repositories.NewSectionContentRepository(db)

	// 2. Create Services (inject config/clients)
	auditLogger := logging.NewAuditLogger()
	metricsCollector := prometheus.NewMetricsCollector()

	// 3. Create Use Cases (inject repositories & services)
	// Portfolio use cases
	createPortfolioUC := portfolio.NewCreatePortfolioUseCase(portfolioRepo, auditLogger, metricsCollector)
	getPortfolioUC := portfolio.NewGetPortfolioUseCase(portfolioRepo)
	getPortfolioPublicUC := portfolio.NewGetPortfolioPublicUseCase(portfolioRepo)
	listPortfoliosUC := portfolio.NewListPortfoliosUseCase(portfolioRepo)
	updatePortfolioUC := portfolio.NewUpdatePortfolioUseCase(portfolioRepo, auditLogger, metricsCollector)
	deletePortfolioUC := portfolio.NewDeletePortfolioUseCase(portfolioRepo, auditLogger, metricsCollector)

	// Category use cases
	createCategoryUC := category.NewCreateCategoryUseCase(categoryRepo, portfolioRepo, auditLogger, metricsCollector)
	getCategoryUC := category.NewGetCategoryUseCase(categoryRepo, portfolioRepo, auditLogger)
	getCategoryPublicUC := category.NewGetCategoryPublicUseCase(categoryRepo)
	listCategoriesUC := category.NewListCategoriesUseCase(categoryRepo)
	updateCategoryUC := category.NewUpdateCategoryUseCase(categoryRepo, portfolioRepo, auditLogger, metricsCollector)
	updateCategoryPositionUC := category.NewUpdateCategoryPositionUseCase(categoryRepo, portfolioRepo, auditLogger)
	bulkReorderCategoriesUC := category.NewBulkReorderCategoriesUseCase(categoryRepo, portfolioRepo, auditLogger)
	deleteCategoryUC := category.NewDeleteCategoryUseCase(categoryRepo, portfolioRepo, auditLogger, metricsCollector)

	// Section use cases
	createSectionUC := section.NewCreateSectionUseCase(sectionRepo, portfolioRepo, auditLogger, metricsCollector)
	getSectionUC := section.NewGetSectionUseCase(sectionRepo, portfolioRepo, auditLogger)
	getSectionPublicUC := section.NewGetSectionPublicUseCase(sectionRepo)
	listSectionsUC := section.NewListSectionsUseCase(sectionRepo)
	updateSectionUC := section.NewUpdateSectionUseCase(sectionRepo, portfolioRepo, auditLogger, metricsCollector)
	updateSectionPositionUC := section.NewUpdateSectionPositionUseCase(sectionRepo, portfolioRepo, auditLogger)
	bulkReorderSectionsUC := section.NewBulkReorderSectionsUseCase(sectionRepo, portfolioRepo, auditLogger)
	deleteSectionUC := section.NewDeleteSectionUseCase(sectionRepo, portfolioRepo, auditLogger, metricsCollector)

	// Project use cases
	createProjectUC := project.NewCreateProjectUseCase(projectRepo, categoryRepo, auditLogger, metricsCollector)
	getProjectUC := project.NewGetProjectUseCase(projectRepo, categoryRepo, auditLogger)
	getProjectPublicUC := project.NewGetProjectPublicUseCase(projectRepo)
	listProjectsUC := project.NewListProjectsUseCase(projectRepo)
	updateProjectUC := project.NewUpdateProjectUseCase(projectRepo, categoryRepo, auditLogger)
	deleteProjectUC := project.NewDeleteProjectUseCase(projectRepo, categoryRepo, auditLogger)

	// Section content use cases
	createSectionContentUC := section_content.NewCreateSectionContentUseCase(sectionContentRepo, sectionRepo, portfolioRepo, auditLogger)
	updateSectionContentUC := section_content.NewUpdateSectionContentUseCase(sectionContentRepo, sectionRepo, portfolioRepo, auditLogger)
	updateSectionContentOrderUC := section_content.NewUpdateSectionContentOrderUseCase(sectionContentRepo, sectionRepo, portfolioRepo, auditLogger)
	deleteSectionContentUC := section_content.NewDeleteSectionContentUseCase(sectionContentRepo, sectionRepo, portfolioRepo, auditLogger)
	getSectionContentPublicUC := section_content.NewGetSectionContentPublicUseCase(sectionContentRepo)
	listSectionContentsBySectionUC := section_content.NewListSectionContentsBySectionUseCase(sectionContentRepo)

	// User use cases
	getCurrentUserUC := user.NewGetCurrentUserUseCase(userRepo)
	updateCurrentUserUC := user.NewUpdateCurrentUserUseCase(userRepo, auditLogger)

	// 4. Create Controllers (inject use cases)
	portfolioController := controllers.NewPortfolioController(
		createPortfolioUC, getPortfolioUC, getPortfolioPublicUC,
		listPortfoliosUC, updatePortfolioUC, deletePortfolioUC,
		categoryRepo, sectionRepo,
	)

	categoryController := controllers.NewCategoryController(
		createCategoryUC, getCategoryUC, getCategoryPublicUC,
		listCategoriesUC, updateCategoryUC, updateCategoryPositionUC,
		bulkReorderCategoriesUC, deleteCategoryUC,
	)

	sectionController := controllers.NewSectionController(
		createSectionUC, getSectionUC, getSectionPublicUC,
		listSectionsUC, updateSectionUC, updateSectionPositionUC,
		bulkReorderSectionsUC, deleteSectionUC,
	)

	projectController := controllers.NewProjectController(
		createProjectUC, getProjectUC, getProjectPublicUC,
		listProjectsUC, updateProjectUC, deleteProjectUC,
		projectRepo,
	)

	sectionContentController := controllers.NewSectionContentController(
		createSectionContentUC, updateSectionContentUC, updateSectionContentOrderUC,
		deleteSectionContentUC, getSectionContentPublicUC, listSectionContentsBySectionUC,
	)

	userController := controllers.NewUserController(getCurrentUserUC, updateCurrentUserUC)
	healthController := controllers.NewHealthController(db)

	// 5. Create Middleware (inject services)
	// TODO: Create real auth provider instead of nil
	authMiddleware := middleware.NewAuthMiddleware(nil)

	// Setup and start server
	router := setupRouter(
		authMiddleware,
		portfolioController,
		categoryController,
		sectionController,
		projectController,
		sectionContentController,
		userController,
		healthController,
	)
	startServer(router, db)
}

func initDatabase() (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_USER", "postgres"),
		getEnv("DB_PASSWORD", "postgres"),
		getEnv("DB_NAME", "portfolio"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_SSLMODE", "disable"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)
	sqlDB.SetConnMaxIdleTime(10 * time.Minute)

	log.Println("✅ Database connected successfully")
	return db, nil
}

func runMigrations(db *gorm.DB) error {
	log.Println("Running database migrations...")

	err := db.AutoMigrate(
		&entities.UserRecord{},
		&entities.PortfolioRecord{},
		&entities.CategoryRecord{},
		&entities.SectionRecord{},
		&entities.ProjectRecord{},
		&entities.SectionContentRecord{},
	)

	if err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	log.Println("✅ Migrations completed successfully")
	return nil
}

func setupRouter(
	authMiddleware *middleware.AuthMiddleware,
	portfolioCtrl *controllers.PortfolioController,
	categoryCtrl *controllers.CategoryController,
	sectionCtrl *controllers.SectionController,
	projectCtrl *controllers.ProjectController,
	sectionContentCtrl *controllers.SectionContentController,
	userCtrl *controllers.UserController,
	healthCtrl *controllers.HealthController,
) *gin.Engine {
	// Set Gin mode
	if getEnv("GIN_MODE", "debug") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// CORS middleware
	router.Use(corsMiddleware())

	// Health endpoints (no auth)
	router.GET("/health", healthCtrl.Health)
	router.GET("/health/db", healthCtrl.DatabaseHealth)

	// API routes
	api := router.Group("/api")
	{
		// Portfolio routes
		portfolios := api.Group("/portfolios")
		{
			portfolios.Use(authMiddleware.Authenticate()).POST("/own", portfolioCtrl.Create)
			portfolios.Use(authMiddleware.Authenticate()).GET("/own", portfolioCtrl.List)
			portfolios.Use(authMiddleware.Authenticate()).GET("/own/:id", portfolioCtrl.GetByID)
			portfolios.Use(authMiddleware.Authenticate()).PUT("/own/:id", portfolioCtrl.Update)
			portfolios.Use(authMiddleware.Authenticate()).DELETE("/own/:id", portfolioCtrl.Delete)

			portfolios.GET("/public/:id", portfolioCtrl.GetPublicByID)
			portfolios.GET("/id/:id", portfolioCtrl.GetPublicByID)
			portfolios.GET("/public/:id/categories", portfolioCtrl.GetPublicCategories)
			portfolios.GET("/public/:id/sections", portfolioCtrl.GetPublicSections)
		}

		// Category routes
		categories := api.Group("/categories")
		{
			categories.Use(authMiddleware.Authenticate()).POST("/own", categoryCtrl.Create)
			categories.Use(authMiddleware.Authenticate()).GET("/own", categoryCtrl.List)
			categories.Use(authMiddleware.Authenticate()).GET("/own/:id", categoryCtrl.GetByID)
			categories.Use(authMiddleware.Authenticate()).PUT("/own/:id", categoryCtrl.Update)
			categories.Use(authMiddleware.Authenticate()).DELETE("/own/:id", categoryCtrl.Delete)
			categories.Use(authMiddleware.Authenticate()).POST("/own/reorder", categoryCtrl.BulkReorder)

			categories.GET("/public/:id", categoryCtrl.GetPublicByID)
			categories.GET("/id/:id", categoryCtrl.GetPublicByID)
			categories.GET("/portfolio/:portfolioId", categoryCtrl.List)
			categories.GET("/portfolio/:portfolioId/projects", categoryCtrl.GetPublicProjects)
		}

		// Section routes
		sections := api.Group("/sections")
		{
			sections.Use(authMiddleware.Authenticate()).POST("/own", sectionCtrl.Create)
			sections.Use(authMiddleware.Authenticate()).GET("/own", sectionCtrl.List)
			sections.Use(authMiddleware.Authenticate()).GET("/own/:id", sectionCtrl.GetByID)
			sections.Use(authMiddleware.Authenticate()).PUT("/own/:id", sectionCtrl.Update)
			sections.Use(authMiddleware.Authenticate()).DELETE("/own/:id", sectionCtrl.Delete)
			sections.Use(authMiddleware.Authenticate()).POST("/own/reorder", sectionCtrl.BulkReorder)

			sections.GET("/public/:id", sectionCtrl.GetPublicByID)
			sections.GET("/id/:id", sectionCtrl.GetPublicByID)
			sections.GET("/portfolio/:portfolioId", sectionCtrl.List)
			sections.GET("/public/:id/contents", sectionCtrl.GetPublicSectionContents)
		}

		// Project routes
		projects := api.Group("/projects")
		{
			projects.Use(authMiddleware.Authenticate()).POST("/own", projectCtrl.Create)
			projects.Use(authMiddleware.Authenticate()).GET("/own", projectCtrl.List)
			projects.Use(authMiddleware.Authenticate()).GET("/own/:id", projectCtrl.GetByID)
			projects.Use(authMiddleware.Authenticate()).PUT("/own/:id", projectCtrl.Update)
			projects.Use(authMiddleware.Authenticate()).DELETE("/own/:id", projectCtrl.Delete)

			projects.GET("/public/:id", projectCtrl.GetPublicByID)
			projects.GET("/category/:categoryId", projectCtrl.GetByCategory)
			projects.GET("/search/skills", projectCtrl.SearchBySkills)
			projects.GET("/search/client", projectCtrl.SearchByClient)
		}

		// Section Content routes
		sectionContents := api.Group("/section-contents")
		{
			sectionContents.Use(authMiddleware.Authenticate()).POST("/own", sectionContentCtrl.Create)
			sectionContents.Use(authMiddleware.Authenticate()).PUT("/own/:id", sectionContentCtrl.Update)
			sectionContents.Use(authMiddleware.Authenticate()).PATCH("/own/:id/order", sectionContentCtrl.UpdateOrder)
			sectionContents.Use(authMiddleware.Authenticate()).DELETE("/own/:id", sectionContentCtrl.Delete)

			sectionContents.GET("/:id", sectionContentCtrl.GetByID)
			sectionContents.GET("/sections/:sectionId/contents", sectionContentCtrl.ListBySection)
		}

		// User routes
		users := api.Group("/users")
		{
			users.Use(authMiddleware.Authenticate()).GET("/me", userCtrl.GetMe)
			users.Use(authMiddleware.Authenticate()).PUT("/me", userCtrl.UpdateMe)
		}
	}

	log.Println("✅ Routes configured successfully")
	return router
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func startServer(router *gin.Engine, db *gorm.DB) {
	port := getEnv("PORT", "8000")
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		log.Printf("🚀 Server starting on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("🛑 Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	// Close database connection
	sqlDB, _ := db.DB()
	if sqlDB != nil {
		_ = sqlDB.Close()
	}

	log.Println("✅ Server exited gracefully")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
