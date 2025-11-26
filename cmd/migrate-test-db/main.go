package main

import (
	"log"
	"path/filepath"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/infrastructure/db"
	"github.com/joho/godotenv"
)

func main() {

	// Load .env.test from project root (consistent with other test helpers)
	projectRoot := filepath.Join("..", "..", "..")
	envTestPath := filepath.Join(projectRoot, ".env.test")
	if err := godotenv.Load(envTestPath); err != nil {
		log.Printf("Note: .env.test not found at %s, using environment variables or defaults", envTestPath)
	}

	// Initialize database connection
	database := db.NewDatabase()
	if err := database.Initialize(); err != nil {
		log.Fatalf("✗ Failed to initialize database: %v", err)
	}
	// Ensure the database is closed and log any error from Close()
	defer func() {
		if cerr := database.Close(); cerr != nil {
			log.Printf("Warning: failed to close database: %v", cerr)
		}
	}()

	log.Println("✓ Database connection established")

	// Run all GORM migrations including triggers
	if err := database.Migrate(); err != nil {
		log.Fatalf("✗ Failed to migrate database: %v", err)
	}

	log.Println("✓ Test database migrations complete")

	// Verify trigger exists
	var triggerExists bool
	err := database.DB.Raw("SELECT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'before_insert_project')").Scan(&triggerExists).Error
	if err != nil {
		log.Fatalf("✗ Failed to check trigger existence: %v", err)
	}

	if !triggerExists {
		log.Fatal("✗ Trigger 'before_insert_project' does not exist!")
	}

	log.Println("✓ Trigger 'before_insert_project' verified")

	// Verify function exists
	var functionExists bool
	err = database.DB.Raw("SELECT EXISTS (SELECT 1 FROM pg_proc WHERE proname = 'set_project_position')").Scan(&functionExists).Error
	if err != nil {
		log.Fatalf("✗ Failed to check function existence: %v", err)
	}

	if !functionExists {
		log.Fatal("✗ Function 'set_project_position' does not exist!")
	}

	log.Println("✓ Function 'set_project_position' verified")
	log.Println("✓ All test database migrations and triggers verified successfully")
}
