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
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/infrastructure/postgres/entities"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/infrastructure/postgres/repositories"
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

	// Initialize repositories
	userRepo := repositories.NewUserRepository(db)
	portfolioRepo := repositories.NewPortfolioRepository(db)
	categoryRepo := repositories.NewCategoryRepository(db)
	sectionRepo := repositories.NewSectionRepository(db)
	projectRepo := repositories.NewProjectRepository(db)
	sectionContentRepo := repositories.NewSectionContentRepository(db)

	// Initialize audit logger (stub for now)
	auditLogger := &stubAuditLogger{}

	// Initialize use cases
	portfolioUseCases := initPortfolioUseCases(portfolioRepo, categoryRepo, sectionRepo, auditLogger)
	categoryUseCases := initCategoryUseCases(categoryRepo, portfolioRepo, sectionRepo, auditLogger)
	sectionUseCases := initSectionUseCases(sectionRepo, portfolioRepo, auditLogger)
	projectUseCases := initProjectUseCases(projectRepo, categoryRepo, portfolioRepo, auditLogger)
	sectionContentUseCases := initSectionContentUseCases(sectionContentRepo, sectionRepo, portfolioRepo, auditLogger)
	userUseCases := initUserUseCases(userRepo, auditLogger)

	// Initialize controllers
	portfolioController := controllers.NewPortfolioController(
		portfolioUseCases.create,
		portfolioUseCases.get,
		portfolioUseCases.getPublic,
		portfolioUseCases.list,
		portfolioUseCases.update,
		portfolioUseCases.delete,
		categoryRepo,
		sectionRepo,
	)

	categoryController := controllers.NewCategoryController(
		categoryUseCases.create,
		categoryUseCases.get,
		categoryUseCases.getPublic,
		categoryUseCases.list,
		categoryUseCases.update,
		categoryUseCases.delete,
		categoryUseCases.bulkReorder,
		categoryUseCases.listByPortfolio,
		projectRepo,
	)

	sectionController := controllers.NewSectionController(
		sectionUseCases.create,
		sectionUseCases.get,
		sectionUseCases.getPublic,
		sectionUseCases.list,
		sectionUseCases.update,
		sectionUseCases.delete,
		sectionUseCases.bulkReorder,
		sectionUseCases.listByPortfolio,
		sectionContentRepo,
	)

	projectController := controllers.NewProjectController(
		projectUseCases.create,
		projectUseCases.get,
		projectUseCases.getPublic,
		projectUseCases.list,
		projectUseCases.update,
		projectUseCases.delete,
		projectUseCases.listByCategory,
		projectUseCases.searchBySkills,
		projectUseCases.searchByClient,
	)

	sectionContentController := controllers.NewSectionContentController(
		sectionContentUseCases.create,
		sectionContentUseCases.update,
		sectionContentUseCases.updateOrder,
		sectionContentUseCases.delete,
		sectionContentUseCases.getPublic,
		sectionContentUseCases.listBySection,
	)

	userController := controllers.NewUserController(
		userUseCases.getCurrent,
		userUseCases.updateCurrent,
	)

	healthController := controllers.NewHealthController(db)

	// Initialize auth middleware
	authMiddleware := middleware.NewAuthMiddleware()

	// Setup Gin router
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

	// Start server
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

type portfolioUseCases struct {
	create    *portfolio.CreatePortfolioUseCase
	get       *portfolio.GetPortfolioUseCase
	getPublic *portfolio.GetPortfolioPublicUseCase
	list      *portfolio.ListPortfoliosUseCase
	update    *portfolio.UpdatePortfolioUseCase
	delete    *portfolio.DeletePortfolioUseCase
}

func initPortfolioUseCases(portfolioRepo, categoryRepo, sectionRepo interface{}, auditLogger interface{}) portfolioUseCases {
	return portfolioUseCases{
		create: portfolio.NewCreatePortfolioUseCase(portfolioRepo.(interface {
			Create(context.Context, interface{}) (interface{}, error)
		}), auditLogger),
		get: portfolio.NewGetPortfolioUseCase(portfolioRepo.(interface {
			GetByID(context.Context, uint) (interface{}, error)
		})),
		getPublic: portfolio.NewGetPortfolioPublicUseCase(portfolioRepo.(interface {
			GetByID(context.Context, uint) (interface{}, error)
		})),
		list: portfolio.NewListPortfoliosUseCase(portfolioRepo.(interface {
			GetByOwnerID(context.Context, string, int, int) ([]interface{}, error)
		})),
		update: portfolio.NewUpdatePortfolioUseCase(portfolioRepo.(interface {
			Update(context.Context, interface{}) error
		}), auditLogger),
		delete: portfolio.NewDeletePortfolioUseCase(portfolioRepo.(interface {
			Delete(context.Context, uint) error
		}), categoryRepo, sectionRepo, auditLogger),
	}
}

type categoryUseCases struct {
	create          *category.CreateCategoryUseCase
	get             *category.GetCategoryUseCase
	getPublic       *category.GetCategoryPublicUseCase
	list            *category.ListCategoriesUseCase
	update          *category.UpdateCategoryUseCase
	delete          *category.DeleteCategoryUseCase
	bulkReorder     *category.BulkReorderCategoriesUseCase
	listByPortfolio *category.ListCategoriesByPortfolioUseCase
}

func initCategoryUseCases(categoryRepo, portfolioRepo, sectionRepo interface{}, auditLogger interface{}) categoryUseCases {
	return categoryUseCases{
		create:          category.NewCreateCategoryUseCase(categoryRepo, portfolioRepo, auditLogger),
		get:             category.NewGetCategoryUseCase(categoryRepo, portfolioRepo),
		getPublic:       category.NewGetCategoryPublicUseCase(categoryRepo),
		list:            category.NewListCategoriesUseCase(categoryRepo),
		update:          category.NewUpdateCategoryUseCase(categoryRepo, portfolioRepo, auditLogger),
		delete:          category.NewDeleteCategoryUseCase(categoryRepo, portfolioRepo, auditLogger),
		bulkReorder:     category.NewBulkReorderCategoriesUseCase(categoryRepo, portfolioRepo, auditLogger),
		listByPortfolio: category.NewListCategoriesByPortfolioUseCase(categoryRepo),
	}
}

type sectionUseCases struct {
	create          *section.CreateSectionUseCase
	get             *section.GetSectionUseCase
	getPublic       *section.GetSectionPublicUseCase
	list            *section.ListSectionsUseCase
	update          *section.UpdateSectionUseCase
	delete          *section.DeleteSectionUseCase
	bulkReorder     *section.BulkReorderSectionsUseCase
	listByPortfolio *section.ListSectionsByPortfolioUseCase
}

func initSectionUseCases(sectionRepo, portfolioRepo interface{}, auditLogger interface{}) sectionUseCases {
	return sectionUseCases{
		create:          section.NewCreateSectionUseCase(sectionRepo, portfolioRepo, auditLogger),
		get:             section.NewGetSectionUseCase(sectionRepo, portfolioRepo),
		getPublic:       section.NewGetSectionPublicUseCase(sectionRepo),
		list:            section.NewListSectionsUseCase(sectionRepo),
		update:          section.NewUpdateSectionUseCase(sectionRepo, portfolioRepo, auditLogger),
		delete:          section.NewDeleteSectionUseCase(sectionRepo, portfolioRepo, auditLogger),
		bulkReorder:     section.NewBulkReorderSectionsUseCase(sectionRepo, portfolioRepo, auditLogger),
		listByPortfolio: section.NewListSectionsByPortfolioUseCase(sectionRepo),
	}
}

type projectUseCases struct {
	create         *project.CreateProjectUseCase
	get            *project.GetProjectUseCase
	getPublic      *project.GetProjectPublicUseCase
	list           *project.ListProjectsUseCase
	update         *project.UpdateProjectUseCase
	delete         *project.DeleteProjectUseCase
	listByCategory *project.ListProjectsByCategoryUseCase
	searchBySkills *project.SearchProjectsBySkillsUseCase
	searchByClient *project.SearchProjectsByClientUseCase
}

func initProjectUseCases(projectRepo, categoryRepo, portfolioRepo interface{}, auditLogger interface{}) projectUseCases {
	return projectUseCases{
		create:         project.NewCreateProjectUseCase(projectRepo, categoryRepo, portfolioRepo, auditLogger),
		get:            project.NewGetProjectUseCase(projectRepo, categoryRepo, portfolioRepo),
		getPublic:      project.NewGetProjectPublicUseCase(projectRepo),
		list:           project.NewListProjectsUseCase(projectRepo),
		update:         project.NewUpdateProjectUseCase(projectRepo, categoryRepo, portfolioRepo, auditLogger),
		delete:         project.NewDeleteProjectUseCase(projectRepo, categoryRepo, portfolioRepo, auditLogger),
		listByCategory: project.NewListProjectsByCategoryUseCase(projectRepo),
		searchBySkills: project.NewSearchProjectsBySkillsUseCase(projectRepo),
		searchByClient: project.NewSearchProjectsByClientUseCase(projectRepo),
	}
}

type sectionContentUseCases struct {
	create        *section_content.CreateSectionContentUseCase
	update        *section_content.UpdateSectionContentUseCase
	updateOrder   *section_content.UpdateSectionContentOrderUseCase
	delete        *section_content.DeleteSectionContentUseCase
	getPublic     *section_content.GetSectionContentPublicUseCase
	listBySection *section_content.ListSectionContentsBySectionUseCase
}

func initSectionContentUseCases(sectionContentRepo, sectionRepo, portfolioRepo interface{}, auditLogger interface{}) sectionContentUseCases {
	return sectionContentUseCases{
		create:        section_content.NewCreateSectionContentUseCase(sectionContentRepo, sectionRepo, portfolioRepo, auditLogger),
		update:        section_content.NewUpdateSectionContentUseCase(sectionContentRepo, sectionRepo, portfolioRepo, auditLogger),
		updateOrder:   section_content.NewUpdateSectionContentOrderUseCase(sectionContentRepo, sectionRepo, portfolioRepo, auditLogger),
		delete:        section_content.NewDeleteSectionContentUseCase(sectionContentRepo, sectionRepo, portfolioRepo, auditLogger),
		getPublic:     section_content.NewGetSectionContentPublicUseCase(sectionContentRepo),
		listBySection: section_content.NewListSectionContentsBySectionUseCase(sectionContentRepo),
	}
}

type userUseCases struct {
	getCurrent    *user.GetCurrentUserUseCase
	updateCurrent *user.UpdateCurrentUserUseCase
}

func initUserUseCases(userRepo interface{}, auditLogger interface{}) userUseCases {
	return userUseCases{
		getCurrent:    user.NewGetCurrentUserUseCase(userRepo),
		updateCurrent: user.NewUpdateCurrentUserUseCase(userRepo, auditLogger),
	}
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
			// Protected routes
			portfolios.Use(authMiddleware.Authenticate()).POST("/own", portfolioCtrl.Create)
			portfolios.Use(authMiddleware.Authenticate()).GET("/own", portfolioCtrl.List)
			portfolios.Use(authMiddleware.Authenticate()).GET("/own/:id", portfolioCtrl.Get)
			portfolios.Use(authMiddleware.Authenticate()).PUT("/own/:id", portfolioCtrl.Update)
			portfolios.Use(authMiddleware.Authenticate()).DELETE("/own/:id", portfolioCtrl.Delete)

			// Public routes
			portfolios.GET("/public/:id", portfolioCtrl.GetPublic)
			portfolios.GET("/id/:id", portfolioCtrl.GetPublic)
			portfolios.GET("/public/:id/categories", portfolioCtrl.ListCategories)
			portfolios.GET("/public/:id/sections", portfolioCtrl.ListSections)
		}

		// Category routes
		categories := api.Group("/categories")
		{
			// Protected routes
			categories.Use(authMiddleware.Authenticate()).POST("/own", categoryCtrl.Create)
			categories.Use(authMiddleware.Authenticate()).GET("/own", categoryCtrl.List)
			categories.Use(authMiddleware.Authenticate()).GET("/own/:id", categoryCtrl.Get)
			categories.Use(authMiddleware.Authenticate()).PUT("/own/:id", categoryCtrl.Update)
			categories.Use(authMiddleware.Authenticate()).DELETE("/own/:id", categoryCtrl.Delete)
			categories.Use(authMiddleware.Authenticate()).POST("/own/reorder", categoryCtrl.BulkReorder)

			// Public routes
			categories.GET("/public/:id", categoryCtrl.GetPublic)
			categories.GET("/id/:id", categoryCtrl.GetPublic)
			categories.GET("/portfolio/:portfolioId", categoryCtrl.ListByPortfolio)
			categories.GET("/portfolio/:portfolioId/projects", categoryCtrl.ListProjects)
		}

		// Section routes
		sections := api.Group("/sections")
		{
			// Protected routes
			sections.Use(authMiddleware.Authenticate()).POST("/own", sectionCtrl.Create)
			sections.Use(authMiddleware.Authenticate()).GET("/own", sectionCtrl.List)
			sections.Use(authMiddleware.Authenticate()).GET("/own/:id", sectionCtrl.Get)
			sections.Use(authMiddleware.Authenticate()).PUT("/own/:id", sectionCtrl.Update)
			sections.Use(authMiddleware.Authenticate()).DELETE("/own/:id", sectionCtrl.Delete)
			sections.Use(authMiddleware.Authenticate()).POST("/own/reorder", sectionCtrl.BulkReorder)

			// Public routes
			sections.GET("/public/:id", sectionCtrl.GetPublic)
			sections.GET("/id/:id", sectionCtrl.GetPublic)
			sections.GET("/portfolio/:portfolioId", sectionCtrl.ListByPortfolio)
			sections.GET("/public/:id/contents", sectionCtrl.ListContents)
		}

		// Project routes
		projects := api.Group("/projects")
		{
			// Protected routes
			projects.Use(authMiddleware.Authenticate()).POST("/own", projectCtrl.Create)
			projects.Use(authMiddleware.Authenticate()).GET("/own", projectCtrl.List)
			projects.Use(authMiddleware.Authenticate()).GET("/own/:id", projectCtrl.Get)
			projects.Use(authMiddleware.Authenticate()).PUT("/own/:id", projectCtrl.Update)
			projects.Use(authMiddleware.Authenticate()).DELETE("/own/:id", projectCtrl.Delete)

			// Public routes
			projects.GET("/public/:id", projectCtrl.GetPublic)
			projects.GET("/category/:categoryId", projectCtrl.ListByCategory)
			projects.GET("/search/skills", projectCtrl.SearchBySkills)
			projects.GET("/search/client", projectCtrl.SearchByClient)
		}

		// Section Content routes
		sectionContents := api.Group("/section-contents")
		{
			// Protected routes
			sectionContents.Use(authMiddleware.Authenticate()).POST("/own", sectionContentCtrl.Create)
			sectionContents.Use(authMiddleware.Authenticate()).PUT("/own/:id", sectionContentCtrl.Update)
			sectionContents.Use(authMiddleware.Authenticate()).PATCH("/own/:id/order", sectionContentCtrl.UpdateOrder)
			sectionContents.Use(authMiddleware.Authenticate()).DELETE("/own/:id", sectionContentCtrl.Delete)

			// Public routes
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
		sqlDB.Close()
	}

	log.Println("✅ Server exited gracefully")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// stubAuditLogger is a simple stub implementation of AuditLogger
type stubAuditLogger struct{}

func (s *stubAuditLogger) LogCreate(ctx context.Context, entity string, id interface{}, metadata map[string]interface{}) {
	log.Printf("AUDIT [CREATE]: %s ID=%v metadata=%+v", entity, id, metadata)
}

func (s *stubAuditLogger) LogUpdate(ctx context.Context, entity string, id interface{}, metadata map[string]interface{}) {
	log.Printf("AUDIT [UPDATE]: %s ID=%v metadata=%+v", entity, id, metadata)
}

func (s *stubAuditLogger) LogDelete(ctx context.Context, entity string, id interface{}, metadata map[string]interface{}) {
	log.Printf("AUDIT [DELETE]: %s ID=%v metadata=%+v", entity, id, metadata)
}
