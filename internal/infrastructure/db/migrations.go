package db

import (
	"fmt"
	"log"
	"strings"

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

// CreateProjectPositionTrigger creates a database trigger to automatically set
// the position field for new projects based on the current maximum position
// within the same category. This ensures proper ordering without race conditions.
func CreateProjectPositionTrigger(db *gorm.DB) error {
	log.Println("Creating project position trigger...")

	// Create the trigger function
	if err := db.Exec(`
		CREATE OR REPLACE FUNCTION set_project_position()
		RETURNS TRIGGER AS $$
		BEGIN
			-- Only set position if it's NULL or 0
			IF NEW.position IS NULL OR NEW.position = 0 THEN
				SELECT COALESCE(MAX(position) + 1, 1) INTO NEW.position
				FROM projects
				WHERE category_id = NEW.category_id
				AND deleted_at IS NULL;
			END IF;
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql;
	`).Error; err != nil {
		return fmt.Errorf("failed to create set_project_position function: %w", err)
	}

	// Drop trigger if it exists and create it
	if err := db.Exec(`
		DROP TRIGGER IF EXISTS before_insert_project ON projects;
	`).Error; err != nil {
		return fmt.Errorf("failed to drop existing trigger: %w", err)
	}

	if err := db.Exec(`
		CREATE TRIGGER before_insert_project
		BEFORE INSERT ON projects
		FOR EACH ROW
		EXECUTE FUNCTION set_project_position();
	`).Error; err != nil {
		return fmt.Errorf("failed to create before_insert_project trigger: %w", err)
	}

	log.Println("Project position trigger created successfully")
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

// AddCascadeDeleteConstraints adds ON DELETE CASCADE to all foreign key relationships
// This ensures orphaned data is automatically cleaned up when parent records are deleted.
// Fixes issue where deleting a portfolio leaves orphaned categories, sections, and projects.
func AddCascadeDeleteConstraints(db *gorm.DB) error {
	log.Println("Adding CASCADE DELETE constraints to foreign keys...")

	// Step 1: Drop existing foreign key constraints without CASCADE
	constraints := []struct {
		table      string
		constraint string
		column     string
		refTable   string
		refColumn  string
	}{
		// Categories -> Portfolios
		{
			table:      "categories",
			constraint: "fk_categories_portfolio",
			column:     "portfolio_id",
			refTable:   "portfolios",
			refColumn:  "id",
		},
		// Sections -> Portfolios
		{
			table:      "sections",
			constraint: "fk_sections_portfolio",
			column:     "portfolio_id",
			refTable:   "portfolios",
			refColumn:  "id",
		},
		// Projects -> Categories
		{
			table:      "projects",
			constraint: "fk_projects_category",
			column:     "category_id",
			refTable:   "categories",
			refColumn:  "id",
		},
		// SectionContent -> Sections (already has CASCADE, but update for consistency)
		{
			table:      "section_contents",
			constraint: "fk_section_contents_section",
			column:     "section_id",
			refTable:   "sections",
			refColumn:  "id",
		},
	}

	for _, fk := range constraints {
		// Check if the constraint exists
		var constraintExists bool
		err := db.Raw(`
			SELECT EXISTS (
				SELECT 1
				FROM information_schema.table_constraints
				WHERE constraint_name = ?
				AND table_name = ?
			)
		`, fk.constraint, fk.table).Scan(&constraintExists).Error

		if err != nil {
			log.Printf("Warning: failed to check if constraint %s exists: %v", fk.constraint, err)
			// Continue to try dropping anyway
		}

		// Drop the old constraint if it exists
		log.Printf("Dropping old constraint %s on table %s...", fk.constraint, fk.table)
		if err := db.Exec(fmt.Sprintf(`
			ALTER TABLE %s DROP CONSTRAINT IF EXISTS %s
		`, fk.table, fk.constraint)).Error; err != nil {
			log.Printf("Warning: failed to drop constraint %s: %v (may not exist)", fk.constraint, err)
			// Continue - constraint may not exist
		}

		// Also try to drop GORM's auto-generated constraint names
		gormConstraint := fmt.Sprintf("fk_%s_%s", fk.table, fk.column)
		log.Printf("Dropping GORM constraint %s on table %s...", gormConstraint, fk.table)
		if err := db.Exec(fmt.Sprintf(`
			ALTER TABLE %s DROP CONSTRAINT IF EXISTS %s
		`, fk.table, gormConstraint)).Error; err != nil {
			log.Printf("Warning: failed to drop GORM constraint %s: %v (may not exist)", gormConstraint, err)
			// Continue - constraint may not exist
		}

		// Add the new constraint with CASCADE DELETE
		log.Printf("Adding CASCADE constraint %s to table %s...", fk.constraint, fk.table)
		if err := db.Exec(fmt.Sprintf(`
			ALTER TABLE %s
			ADD CONSTRAINT %s
			FOREIGN KEY (%s)
			REFERENCES %s(%s)
			ON DELETE CASCADE
		`, fk.table, fk.constraint, fk.column, fk.refTable, fk.refColumn)).Error; err != nil {
			return fmt.Errorf("failed to add CASCADE constraint %s on %s: %w", fk.constraint, fk.table, err)
		}
	}

	log.Println("CASCADE DELETE constraints added successfully")
	log.Println("Migration complete: Deleting a portfolio will now cascade delete all related categories, sections, projects, and section_contents")
	return nil
}

// ApplyImageIndexes adds database indexes for the images table
func ApplyImageIndexes(db *gorm.DB) error {
	log.Println("Applying image table indexes...")

	// Index on images.entity_type for faster polymorphic queries
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_images_entity_type
		ON images(entity_type)
		WHERE deleted_at IS NULL
	`).Error; err != nil {
		return fmt.Errorf("failed to create index on images.entity_type: %w", err)
	}

	// Index on images.entity_id for faster polymorphic queries
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_images_entity_id
		ON images(entity_id)
		WHERE deleted_at IS NULL
	`).Error; err != nil {
		return fmt.Errorf("failed to create index on images.entity_id: %w", err)
	}

	// Index on images.owner_id for faster owner-based queries
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_images_owner_id
		ON images(owner_id)
		WHERE deleted_at IS NULL
	`).Error; err != nil {
		return fmt.Errorf("failed to create index on images.owner_id: %w", err)
	}

	// Composite index for polymorphic relationship queries (entity_type + entity_id)
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_images_entity
		ON images(entity_type, entity_id)
		WHERE deleted_at IS NULL
	`).Error; err != nil {
		return fmt.Errorf("failed to create composite index on images(entity_type, entity_id): %w", err)
	}

	// Index on is_main for faster main image queries
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_images_is_main
		ON images(is_main)
		WHERE deleted_at IS NULL AND is_main = true
	`).Error; err != nil {
		return fmt.Errorf("failed to create index on images.is_main: %w", err)
	}

	log.Println("Image table indexes applied successfully")
	return nil
}

// MigrateProjectImages migrates existing project image URLs to the new Image model
// This handles backward compatibility for projects with old-style images and main_image fields
func MigrateProjectImages(db *gorm.DB) error {
	log.Println("Starting project image data migration...")

	// Check if the old columns still exist
	var hasImagesColumn, hasMainImageColumn bool

	err := db.Raw(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_name = 'projects' AND column_name = 'images'
		)
	`).Scan(&hasImagesColumn).Error
	if err != nil {
		return fmt.Errorf("failed to check for images column: %w", err)
	}

	err = db.Raw(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_name = 'projects' AND column_name = 'main_image'
		)
	`).Scan(&hasMainImageColumn).Error
	if err != nil {
		return fmt.Errorf("failed to check for main_image column: %w", err)
	}

	if !hasImagesColumn && !hasMainImageColumn {
		log.Println("Old image columns not found, migration already completed or not needed")
		return nil
	}

	// Query to get all projects with their old image data
	type OldProjectData struct {
		ID        uint
		Images    []string // Will be populated from text[] column
		MainImage string
		OwnerID   string
	}

	var projects []OldProjectData

	// Build query based on which columns exist
	query := "SELECT id, owner_id"
	if hasImagesColumn {
		query += ", images"
	}
	if hasMainImageColumn {
		query += ", main_image"
	}
	query += " FROM projects WHERE deleted_at IS NULL"

	if err := db.Raw(query).Scan(&projects).Error; err != nil {
		log.Printf("Warning: failed to query projects for migration: %v", err)
		// Don't fail - might be running on fresh database
		return nil
	}

	migratedCount := 0
	errorCount := 0

	for _, project := range projects {
		// Skip if no images to migrate
		if len(project.Images) == 0 && project.MainImage == "" {
			continue
		}

		// Migrate main_image first (if exists)
		if project.MainImage != "" {
			image := map[string]interface{}{
				"url":           project.MainImage,
				"thumbnail_url": project.MainImage, // Use same URL for old images
				"file_name":     extractFilename(project.MainImage),
				"file_size":     0, // Unknown for old data
				"mime_type":     guessMimeType(project.MainImage),
				"alt":           "",
				"owner_id":      project.OwnerID,
				"type":          "image",
				"entity_id":     project.ID,
				"entity_type":   "project",
				"is_main":       true,
			}

			if err := db.Table("images").Create(image).Error; err != nil {
				log.Printf("Warning: failed to migrate main_image for project %d: %v", project.ID, err)
				errorCount++
			} else {
				migratedCount++
			}
		}

		// Migrate images array
		for _, imageURL := range project.Images {
			if imageURL == "" || imageURL == project.MainImage {
				continue // Skip empty or duplicate of main image
			}

			image := map[string]interface{}{
				"url":           imageURL,
				"thumbnail_url": imageURL, // Use same URL for old images
				"file_name":     extractFilename(imageURL),
				"file_size":     0, // Unknown for old data
				"mime_type":     guessMimeType(imageURL),
				"alt":           "",
				"owner_id":      project.OwnerID,
				"type":          "image",
				"entity_id":     project.ID,
				"entity_type":   "project",
				"is_main":       false,
			}

			if err := db.Table("images").Create(image).Error; err != nil {
				log.Printf("Warning: failed to migrate image for project %d: %v", project.ID, err)
				errorCount++
			} else {
				migratedCount++
			}
		}
	}

	log.Printf("Project image migration completed: %d images migrated, %d errors", migratedCount, errorCount)

	// Drop old columns if migration was successful and no errors
	if errorCount == 0 {
		if hasImagesColumn {
			log.Println("Dropping old 'images' column from projects table...")
			if err := db.Exec("ALTER TABLE projects DROP COLUMN IF EXISTS images").Error; err != nil {
				log.Printf("Warning: failed to drop images column: %v", err)
			} else {
				log.Println("Old 'images' column dropped successfully")
			}
		}

		if hasMainImageColumn {
			log.Println("Dropping old 'main_image' column from projects table...")
			if err := db.Exec("ALTER TABLE projects DROP COLUMN IF EXISTS main_image").Error; err != nil {
				log.Printf("Warning: failed to drop main_image column: %v", err)
			} else {
				log.Println("Old 'main_image' column dropped successfully")
			}
		}
	} else {
		log.Printf("Skipping column drop due to %d migration errors", errorCount)
	}

	return nil
}

// extractFilename extracts filename from URL
func extractFilename(url string) string {
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return "unknown"
}

// guessMimeType guesses MIME type from file extension
func guessMimeType(url string) string {
	lowerURL := strings.ToLower(url)
	if strings.HasSuffix(lowerURL, ".png") {
		return "image/png"
	} else if strings.HasSuffix(lowerURL, ".jpg") || strings.HasSuffix(lowerURL, ".jpeg") {
		return "image/jpeg"
	} else if strings.HasSuffix(lowerURL, ".webp") {
		return "image/webp"
	} else if strings.HasSuffix(lowerURL, ".gif") {
		return "image/gif"
	}
	return "image/jpeg" // Default
}
