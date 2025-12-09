package test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/infrastructure/db"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/infrastructure/server"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	testDB     *db.Database
	testServer *server.Server
	testLogger *logrus.Logger
	baseURL    string
)

// TestMain runs once before all tests
func TestMain(m *testing.M) {
	// Setup
	setupTestEnvironment()
	testDB = setupTestDatabase()
	testLogger = setupTestLogger()
	testServer = setupTestServer()

	// Start server in background
	go func() {
		if err := testServer.Start(); err != nil && err != http.ErrServerClosed {
			testLogger.Fatalf("Failed to start test server: %v", err)
		}
	}()

	// Wait for server to be ready
	if err := waitForServer(baseURL, 50); err != nil {
		testLogger.Fatalf("Test server not ready: %v", err)
	}
	fmt.Println("Test server is ready")

	// Run all tests
	code := m.Run()

	// Cleanup
	teardownTestServer()
	teardownTestDatabase()

	os.Exit(code)
}

func setupTestEnvironment() {
	// Load .env.test file from project root
	// Go up two directories from backend/cmd/test to project root
	projectRoot := filepath.Join("..", "..", "..")
	envTestPath := filepath.Join(projectRoot, ".env.test")

	// Load .env.test if it exists
	if err := godotenv.Load(envTestPath); err != nil {
		// .env.test is optional, continue with environment variables or defaults
		fmt.Printf("Note: .env.test not found at %s, using environment variables or defaults\n", envTestPath)
	}

	// Set test-specific environment variables (can override .env.test)
	os.Setenv("GIN_MODE", "test")
	os.Setenv("LOG_LEVEL", "error") // Reduce noise in test output

	// Ensure TESTING_MODE is enabled
	if os.Getenv("TESTING_MODE") == "" {
		os.Setenv("TESTING_MODE", "true")
	}

	// Set base URL from PORT
	port := os.Getenv("PORT")
	if port == "" {
		port = "8888"
		os.Setenv("PORT", port)
	}
	baseURL = fmt.Sprintf("http://localhost:%s", port)
}

func setupTestDatabase() *db.Database {
	database := db.NewDatabase()
	if err := database.Initialize(); err != nil {
		fmt.Printf("FATAL: Failed to initialize test database: %v\n", err)
		os.Exit(1)
	}

	// Run migrations
	if err := database.Migrate(); err != nil {
		fmt.Printf("FATAL: Failed to migrate test database: %v\n", err)
		os.Exit(1)
	}

	// Clean database before tests
	if err := cleanDatabaseWithError(database.DB); err != nil {
		fmt.Printf("FATAL: Failed to clean test database: %v\n", err)
		os.Exit(1)
	}

	return database
}

func setupTestLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Only show errors in tests
	logger.SetFormatter(&logrus.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})
	return logger
}

func setupTestServer() *server.Server {
	return server.NewServer("8888", testDB, testLogger)
}

func teardownTestServer() {
	if testServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := testServer.Shutdown(ctx); err != nil {
			testLogger.Printf("Error shutting down test server: %v", err)
		}
	}
}

func teardownTestDatabase() {
	if testDB != nil && testDB.DB != nil {
		if err := cleanDatabaseWithError(testDB.DB); err != nil {
			fmt.Printf("Warning: Failed to clean database during teardown: %v\n", err)
		}
		sqlDB, err := testDB.DB.DB()
		if err == nil {
			if err := sqlDB.Close(); err != nil {
				fmt.Printf("Warning: Failed to close database connection: %v\n", err)
			}
		}
	}
}

// cleanDatabase truncates all tables - logs errors but doesn't fail (for use in tests)
func cleanDatabase(db *gorm.DB) {
	if err := cleanDatabaseWithError(db); err != nil {
		fmt.Printf("Warning: Database cleanup error: %v\n", err)
	}
}

// cleanDatabaseWithError truncates all tables and returns any errors (for use in setup/teardown)
func cleanDatabaseWithError(db *gorm.DB) error {
	// Disable foreign key checks
	if err := db.Exec("SET session_replication_role = 'replica'").Error; err != nil {
		return fmt.Errorf("failed to disable foreign key checks: %w", err)
	}

	// Truncate all tables in proper order (children before parents)
	// Note: "images" table has been removed via RemoveImageFeature migration
	tables := []string{
		"section_contents",
		"projects",
		"categories",
		"sections",
		"portfolios",
	}

	for _, table := range tables {
		query := fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)
		if err := db.Exec(query).Error; err != nil {
			// If the table doesn't exist, log and continue (useful when images were removed)
			if err != nil && (strings.Contains(err.Error(), "does not exist") || strings.Contains(err.Error(), "42P01")) {
				fmt.Printf("Note: table %s does not exist, skipping\n", table)
				continue
			}

			// Attempt to re-enable foreign key checks before returning error
			retryErr := db.Exec("SET session_replication_role = 'origin'").Error
			if retryErr != nil {
				fmt.Printf("Warning: failed to re-enable foreign key checks: %v\n", retryErr)
			}
			return fmt.Errorf("failed to truncate table %s: %w", table, err)
		}
	}

	// Re-enable foreign key checks
	if err := db.Exec("SET session_replication_role = 'origin'").Error; err != nil {
		return fmt.Errorf("failed to re-enable foreign key checks: %w", err)
	}

	return nil
}

// waitForServer polls the health endpoint until the server is ready
func waitForServer(baseURL string, maxAttempts int) error {
	for i := 0; i < maxAttempts; i++ {
		resp, err := http.Get(baseURL + "/health")
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			return nil
		}
		if resp != nil {
			resp.Body.Close()
		}
		if i < maxAttempts-1 {
			time.Sleep(100 * time.Millisecond)
		}
	}
	return fmt.Errorf("server did not become ready after %d attempts", maxAttempts)
}
