package main

import (
	"fmt"
	"log"
	"os"

	"devhive-backend/db"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		// Default to Neon production
		databaseURL = "postgresql://neondb_owner:npg_EoarkRfZM5t2@ep-shy-waterfall-aflt9iog-pooler.c-2.us-west-2.aws.neon.tech/neondb?sslmode=require"
	}

	log.Printf("Connecting to database...")

	config, err := pgx.ParseConfig(databaseURL)
	if err != nil {
		log.Fatalf("Failed to parse database config: %v", err)
	}

	database := stdlib.OpenDB(*config)
	if err := database.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	defer database.Close()

	log.Println("Connected to database successfully")

	// Run migrations
	if err := db.RunMigrations(database); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	// Create indexes
	if err := db.CreateIndexes(database); err != nil {
		log.Printf("Warning: Index creation had issues: %v", err)
	}

	// Verify NOTIFY triggers
	if err := db.VerifyNotifyTriggers(database); err != nil {
		log.Printf("Warning: Trigger verification had issues: %v", err)
	}

	fmt.Println("\n✅ All migrations completed successfully!")
}
