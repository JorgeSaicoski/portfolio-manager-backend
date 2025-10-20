package test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/db"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/server"
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
	// In production, you'd want a proper health check here

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
	database.Initialize()

	// Run migrations
	database.Migrate()

	// Clean database before tests
	cleanDatabase(database.DB)

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
		ctx := context.Background()
		testServer.Shutdown(ctx)
	}
}

func teardownTestDatabase() {
	if testDB != nil && testDB.DB != nil {
		cleanDatabase(testDB.DB)
		sqlDB, err := testDB.DB.DB()
		if err == nil {
			sqlDB.Close()
		}
	}
}

// cleanDatabase truncates all tables
func cleanDatabase(db *gorm.DB) {
	// Disable foreign key checks
	db.Exec("SET session_replication_role = 'replica'")

	// Truncate all tables
	db.Exec("TRUNCATE TABLE projects CASCADE")
	db.Exec("TRUNCATE TABLE categories CASCADE")
	db.Exec("TRUNCATE TABLE sections CASCADE")
	db.Exec("TRUNCATE TABLE portfolios CASCADE")

	// Re-enable foreign key checks
	db.Exec("SET session_replication_role = 'origin'")
}

// Helper to run tests in a transaction (for isolation)
func runInTransaction(t *testing.T, fn func(tx *gorm.DB)) {
	tx := testDB.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			t.Fatalf("Test panicked: %v", r)
		}
	}()

	fn(tx)

	// Always rollback to keep tests isolated
	tx.Rollback()
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// NewTestRecorder creates a new httptest.ResponseRecorder with proper setup
func NewTestRecorder() *httptest.ResponseRecorder {
	return httptest.NewRecorder()
}

// PrintTestSeparator prints a visual separator in test output
func PrintTestSeparator(testName string) {
	fmt.Printf("\n=== Running: %s ===\n", testName)
}
