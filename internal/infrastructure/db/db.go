package db

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	models2 "github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/models"
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

	// Get query timeout from environment (default: 30 seconds)
	queryTimeoutStr := os.Getenv("DB_QUERY_TIMEOUT")
	queryTimeout := 30 * time.Second
	if queryTimeoutStr != "" {
		if timeout, err := time.ParseDuration(queryTimeoutStr + "s"); err == nil {
			queryTimeout = timeout
		}
	}

	var err error
	d.DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		// Enable prepared statement caching for better performance
		PrepareStmt: true,

		// Query timeout configuration
		NowFunc: func() time.Time {
			return time.Now()
		},

		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold:             queryTimeout,
				LogLevel:                  getLogLevel(),
				IgnoreRecordNotFoundError: true,
				Colorful:                  false,
			},
		),
	})

	if err == nil {
		log.Printf("Database query timeout configured: %s", queryTimeout)
	}

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

	// Connection pool settings - configurable via environment variables
	maxIdleConns := d.getEnvInt("DB_MAX_IDLE_CONNS", 10)
	maxOpenConns := d.getEnvInt("DB_MAX_OPEN_CONNS", 100)
	connMaxLifetime := d.getEnvDuration("DB_CONN_MAX_LIFETIME", time.Hour)
	connMaxIdleTime := d.getEnvDuration("DB_CONN_MAX_IDLE_TIME", 10*time.Minute)

	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetConnMaxLifetime(connMaxLifetime)
	sqlDB.SetConnMaxIdleTime(connMaxIdleTime)

	log.Printf("Database connection pool configured: MaxIdle=%d, MaxOpen=%d, MaxLifetime=%s, MaxIdleTime=%s",
		maxIdleConns, maxOpenConns, connMaxLifetime, connMaxIdleTime)

	log.Println("Connected to PostgreSQL database")
	return nil
}

func (d *Database) Migrate() error {
	if d.DB == nil {
		return fmt.Errorf("database not initialized")
	}

	err := d.DB.AutoMigrate(
		&models2.Portfolio{},
		&models2.Section{},
		&models2.SectionContent{},
		&models2.Category{},
		&models2.Project{},
	)

	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	log.Println("Database migration completed successfully")

	// Apply performance indexes
	if err := ApplyPerformanceIndexes(d.DB); err != nil {
		return fmt.Errorf("failed to apply performance indexes: %w", err)
	}

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

func (d *Database) getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func (d *Database) getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
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

// getLogLevel returns the appropriate GORM log level based on environment
func getLogLevel() logger.LogLevel {
	logLevel := os.Getenv("DB_LOG_LEVEL")
	switch logLevel {
	case "silent":
		return logger.Silent
	case "error":
		return logger.Error
	case "warn":
		return logger.Warn
	case "info":
		return logger.Info
	default:
		// In production, use Error level; in development, use Info
		if os.Getenv("GIN_MODE") == "release" {
			return logger.Error
		}
		return logger.Info
	}
}
