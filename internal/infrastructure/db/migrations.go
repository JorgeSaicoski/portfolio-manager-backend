package db

import (
	"fmt"
	"log"

	"gorm.io/gorm"
)

// ApplyPerformanceIndexes adds database indexes for frequently queried fields
func ApplyPerformanceIndexes(db *gorm.DB) error {
	log.Println("Applying performance indexes...")

	// Index on sections.portfolio_id for faster foreign key lookups
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_sections_portfolio_id
		ON sections(portfolio_id)
		WHERE deleted_at IS NULL
	`).Error; err != nil {
		return fmt.Errorf("failed to create index on sections.portfolio_id: %w", err)
	}

	// Index on projects.category_id for faster foreign key lookups
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_projects_category_id
		ON projects(category_id)
		WHERE deleted_at IS NULL
	`).Error; err != nil {
		return fmt.Errorf("failed to create index on projects.category_id: %w", err)
	}

	// Index on projects.owner_id for faster owner-based queries
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_projects_owner_id
		ON projects(owner_id)
		WHERE deleted_at IS NULL
	`).Error; err != nil {
		return fmt.Errorf("failed to create index on projects.owner_id: %w", err)
	}

	// Index on section_contents.section_id for faster foreign key lookups
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_section_contents_section_id
		ON section_contents(section_id)
		WHERE deleted_at IS NULL
	`).Error; err != nil {
		return fmt.Errorf("failed to create index on section_contents.section_id: %w", err)
	}

	// Index on categories.portfolio_id for faster foreign key lookups
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_categories_portfolio_id
		ON categories(portfolio_id)
		WHERE deleted_at IS NULL
	`).Error; err != nil {
		return fmt.Errorf("failed to create index on categories.portfolio_id: %w", err)
	}

	// Composite index for sections ordered by position within a portfolio
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_sections_portfolio_position
		ON sections(portfolio_id, position)
		WHERE deleted_at IS NULL
	`).Error; err != nil {
		return fmt.Errorf("failed to create composite index on sections(portfolio_id, position): %w", err)
	}

	// Composite index for projects ordered by position within a category
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_projects_category_position
		ON projects(category_id, position)
		WHERE deleted_at IS NULL
	`).Error; err != nil {
		return fmt.Errorf("failed to create composite index on projects(category_id, position): %w", err)
	}

	// Composite index for section_contents ordered within a section
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_section_contents_section_order
		ON section_contents(section_id, "order")
		WHERE deleted_at IS NULL
	`).Error; err != nil {
		return fmt.Errorf("failed to create composite index on section_contents(section_id, order): %w", err)
	}

	// Index on portfolios.owner_id for faster owner-based queries
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_portfolios_owner_id
		ON portfolios(owner_id)
		WHERE deleted_at IS NULL
	`).Error; err != nil {
		return fmt.Errorf("failed to create index on portfolios.owner_id: %w", err)
	}

	// Index on sections.owner_id for faster owner-based queries
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_sections_owner_id
		ON sections(owner_id)
		WHERE deleted_at IS NULL
	`).Error; err != nil {
		return fmt.Errorf("failed to create index on sections.owner_id: %w", err)
	}

	log.Println("Performance indexes applied successfully")
	return nil
}

// PopulateInitialCategoryCount updates existing portfolios with their current category count
// This is a one-time migration to populate the new category_count field
func PopulateInitialCategoryCount(db *gorm.DB) error {
	log.Println("Populating initial category_count for existing portfolios...")

	// Update all portfolios with their current category count
	// Using a subquery to count categories per portfolio
	result := db.Exec(`
		UPDATE portfolios
		SET category_count = (
			SELECT COUNT(*)
			FROM categories
			WHERE categories.portfolio_id = portfolios.id
			AND categories.deleted_at IS NULL
		)
		WHERE portfolios.deleted_at IS NULL
		AND portfolios.category_count = 0
	`)

	if result.Error != nil {
		return fmt.Errorf("failed to populate category_count: %w", result.Error)
	}

	log.Printf("Updated category_count for %d portfolios\n", result.RowsAffected)
	return nil
}
