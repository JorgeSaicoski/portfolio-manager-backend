package db

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Database struct {
	DB *gorm.DB
}

func NewDatabase() *Database {
	return &Database{}
}

func (d *Database) Initialize() error {
	dsn := d.buildDSN()

	var err error
	d.DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold:             time.Second,
				LogLevel:                  logger.Info,
				IgnoreRecordNotFoundError: true,
				Colorful:                  false,
			},
		),
	})

	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test connection
	sqlDB, err := d.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("Connected to PostgreSQL database")
	return nil
}

func (d *Database) Migrate() error {
	if d.DB == nil {
		return fmt.Errorf("database not initialized")
	}

	err := d.DB.AutoMigrate(
		&models.Portfolio{},
		&models.Section{},
		&models.SectionContent{},
		&models.Category{},
		&models.Project{},
	)

	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	log.Println("Database migration completed successfully")
	return nil
}

func (d *Database) buildDSN() string {
	host := d.getEnv("DB_HOST", "localhost")
	port := d.getEnv("DB_PORT", "5432")
	user := d.getEnv("DB_USER", "backend_user")
	password := d.getEnv("DB_PASSWORD", "backend_pass")
	dbname := d.getEnv("DB_NAME", "db")
	sslmode := d.getEnv("DB_SSLMODE", "disable")
	timezone := d.getEnv("DB_TIMEZONE", "UTC")

	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		host, user, password, dbname, port, sslmode, timezone)
}

func (d *Database) getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func (d *Database) Close() error {
	if d.DB == nil {
		return nil
	}

	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}

	return sqlDB.Close()
}
