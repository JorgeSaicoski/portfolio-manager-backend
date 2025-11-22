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

	// Composite index for categories ordered by position within a portfolio
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_categories_portfolio_position
		ON categories(portfolio_id, position)
		WHERE deleted_at IS NULL
	`).Error; err != nil {
		return fmt.Errorf("failed to create composite index on categories(portfolio_id, position): %w", err)
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

// CreateCategoryPositionTrigger creates a database trigger to automatically set
// the position field for new categories based on the current maximum position
// within the same portfolio. This ensures proper ordering without race conditions.
func CreateCategoryPositionTrigger(db *gorm.DB) error {
	log.Println("Creating category position trigger...")

	// Create the trigger function
	if err := db.Exec(`
		CREATE OR REPLACE FUNCTION set_category_position()
		RETURNS TRIGGER AS $$
		BEGIN
			-- Only set position if it's NULL or 0
			IF NEW.position IS NULL OR NEW.position = 0 THEN
				SELECT COALESCE(MAX(position) + 1, 1) INTO NEW.position
				FROM categories
				WHERE portfolio_id = NEW.portfolio_id
				AND deleted_at IS NULL;
			END IF;
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql;
	`).Error; err != nil {
		return fmt.Errorf("failed to create set_category_position function: %w", err)
	}

	// Drop trigger if it exists and create it
	if err := db.Exec(`
		DROP TRIGGER IF EXISTS before_insert_category ON categories;
	`).Error; err != nil {
		return fmt.Errorf("failed to drop existing trigger: %w", err)
	}

	if err := db.Exec(`
		CREATE TRIGGER before_insert_category
		BEFORE INSERT ON categories
		FOR EACH ROW
		EXECUTE FUNCTION set_category_position();
	`).Error; err != nil {
		return fmt.Errorf("failed to create before_insert_category trigger: %w", err)
	}

	log.Println("Category position trigger created successfully")
	return nil
}

// DropCategoryCountColumn removes the category_count field from portfolios table
// This field is redundant and can cause sync issues; position is now managed by trigger
func DropCategoryCountColumn(db *gorm.DB) error {
	log.Println("Dropping category_count column from portfolios table...")

	// Check if column exists before dropping
	var columnExists bool
	err := db.Raw(`
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.columns
			WHERE table_name = 'portfolios'
			AND column_name = 'category_count'
		)
	`).Scan(&columnExists).Error

	if err != nil {
		return fmt.Errorf("failed to check if category_count column exists: %w", err)
	}

	if !columnExists {
		log.Println("category_count column does not exist, skipping drop")
		return nil
	}

	// Drop the column
	if err := db.Exec(`
		ALTER TABLE portfolios DROP COLUMN IF EXISTS category_count;
	`).Error; err != nil {
		return fmt.Errorf("failed to drop category_count column: %w", err)
	}

	log.Println("category_count column dropped successfully")
	return nil
}
