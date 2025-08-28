package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"devhive-backend/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDB initializes the PostgreSQL database connection using GORM
func InitDB() error {
	// First try to use the Fly.io DATABASE_URL (standard format)
	dbURL := os.Getenv("DATABASE_URL")

	// Fallback to Fly.io connection string (with double underscores)
	if dbURL == "" {
		dbURL = os.Getenv("ConnectionStrings__DbConnection")
	}

	// Fallback to single underscore version
	if dbURL == "" {
		dbURL = os.Getenv("ConnectionStringsDbConnection")
	}

	// Fallback to individual environment variables if connection string not available
	if dbURL == "" {
		dbURL = config.GetDatabaseURL()
	}

	// Configure GORM logger
	gormLogger := logger.Default.LogMode(logger.Info)
	if os.Getenv("ENV") == "production" {
		gormLogger = logger.Default.LogMode(logger.Error)
	}

	// Open database with GORM
	var err error
	DB, err = gorm.Open(postgres.Open(dbURL), &gorm.Config{
		Logger: gormLogger,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return fmt.Errorf("error opening database: %v", err)
	}

	// Get the underlying sql.DB for connection pool settings
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("error getting underlying sql.DB: %v", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(25)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Test the connection
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("error connecting to database: %v", err)
	}

	log.Println("Database connected successfully with GORM")

	// Run GORM migrations
	if err := runGORMMigrations(); err != nil {
		log.Printf("Warning: Error running GORM migrations: %v", err)
	}

	return nil
}

// CloseDB closes the database connection
func CloseDB() {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			log.Printf("Error getting underlying sql.DB: %v", err)
			return
		}
		sqlDB.Close()
		log.Println("Database connection closed")
	}
}

// runGORMMigrations runs the database schema migrations using GORM
func runGORMMigrations() error {
	log.Println("Running GORM migrations...")

	// Run auto-migrations
	if err := AutoMigrate(DB); err != nil {
		return fmt.Errorf("error running auto-migrations: %v", err)
	}

	// Create indexes for performance
	if err := CreateIndexes(DB); err != nil {
		log.Printf("Warning: Error creating indexes: %v", err)
	}

	// Seed initial data if needed
	if err := SeedData(DB); err != nil {
		log.Printf("Warning: Error seeding data: %v", err)
	}

	log.Println("GORM migrations completed successfully")
	return nil
}

// GetDB returns the GORM database instance
func GetDB() *gorm.DB {
	return DB
}

// GetRawDB returns the underlying sql.DB instance (useful for raw SQL queries)
func GetRawDB() (*sql.DB, error) {
	if DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	return DB.DB()
}
