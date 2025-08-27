package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"devhive-backend/config"

	_ "github.com/lib/pq"
)

var DB *sql.DB

// InitDB initializes the PostgreSQL database connection
func InitDB() error {
	// First try to use the Fly.io connection string
	dbURL := os.Getenv("ConnectionStringsDbConnection")

	// Fallback to individual environment variables if connection string not available
	if dbURL == "" {
		dbURL = config.GetDatabaseURL()
	}

	var err error
	DB, err = sql.Open("postgres", dbURL)
	if err != nil {
		return fmt.Errorf("error opening database: %v", err)
	}

	// Test the connection
	if err := DB.Ping(); err != nil {
		return fmt.Errorf("error connecting to database: %v", err)
	}

	// Set connection pool settings
	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(25)

	log.Println("Database connected successfully")

	// Run migrations if they exist
	if err := runMigrations(); err != nil {
		log.Printf("Warning: Error running migrations: %v", err)
	}

	return nil
}

// CloseDB closes the database connection
func CloseDB() {
	if DB != nil {
		DB.Close()
		log.Println("Database connection closed")
	}
}

// runMigrations runs the database schema migrations
func runMigrations() error {
	// Check if migrations directory exists
	if _, err := os.Stat("migrations"); os.IsNotExist(err) {
		log.Println("Migrations directory not found, skipping migrations")
		return nil
	}

	// Read and execute schema.sql
	schemaPath := "db/schema.sql"
	if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
		log.Println("Schema file not found, skipping schema creation")
		return nil
	}

	schemaBytes, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("error reading schema file: %v", err)
	}

	// Execute the schema
	if _, err := DB.Exec(string(schemaBytes)); err != nil {
		return fmt.Errorf("error executing schema: %v", err)
	}

	log.Println("Database schema created successfully")
	return nil
}

// GetDB returns the database instance
func GetDB() *sql.DB {
	return DB
}
