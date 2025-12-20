package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"devhive-backend/internal/config"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Connect to database
	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	ctx := context.Background()

	fmt.Println("🔍 Testing Migration: 004_add_cache_invalidation_triggers.sql\n")

	// 1. Check if migration was recorded
	fmt.Println("1. Checking if migration was recorded in schema_migrations...")
	var count int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM schema_migrations WHERE version = $1", "004_add_cache_invalidation_triggers.sql").Scan(&count)
	if err != nil {
		log.Fatal("Failed to query schema_migrations:", err)
	}
	if count > 0 {
		fmt.Println("   ✅ Migration recorded in schema_migrations\n")
	} else {
		fmt.Println("   ❌ Migration NOT found in schema_migrations\n")
		os.Exit(1)
	}

	// 2. Check if function exists
	fmt.Println("2. Checking if notify_cache_invalidation() function exists...")
	var funcExists bool
	err = db.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM pg_proc p
			JOIN pg_namespace n ON p.pronamespace = n.oid
			WHERE n.nspname = 'public' AND p.proname = 'notify_cache_invalidation'
		)
	`).Scan(&funcExists)
	if err != nil {
		log.Fatal("Failed to check function:", err)
	}
	if funcExists {
		fmt.Println("   ✅ Function notify_cache_invalidation() exists\n")
	} else {
		fmt.Println("   ❌ Function notify_cache_invalidation() NOT found\n")
		os.Exit(1)
	}

	// 3. Check triggers on each table
	tables := []string{"projects", "sprints", "tasks", "project_members"}
	triggerName := "cache_invalidate"

	for _, table := range tables {
		fmt.Printf("3.%d Checking trigger on %s table...\n", len(tables)-len(tables)+1, table)
		var triggerExists bool
		err = db.QueryRowContext(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM pg_trigger t
				JOIN pg_class c ON t.tgrelid = c.oid
				JOIN pg_namespace n ON c.relnamespace = n.oid
				WHERE n.nspname = 'public' 
				AND c.relname = $1
				AND t.tgname LIKE $2
			)
		`, table, table+"_%"+triggerName+"%").Scan(&triggerExists)
		if err != nil {
			log.Printf("   ⚠️  Failed to check trigger on %s: %v\n", table, err)
			continue
		}
		if triggerExists {
			fmt.Printf("   ✅ Trigger exists on %s table\n\n", table)
		} else {
			fmt.Printf("   ❌ Trigger NOT found on %s table\n\n", table)
			os.Exit(1)
		}
	}

	// 4. Get trigger details
	fmt.Println("4. Getting trigger details...")
	rows, err := db.QueryContext(ctx, `
		SELECT 
			c.relname AS table_name,
			t.tgname AS trigger_name,
			CASE 
				WHEN t.tgenabled = 'O' THEN 'enabled'
				ELSE 'disabled'
			END AS status
		FROM pg_trigger t
		JOIN pg_class c ON t.tgrelid = c.oid
		JOIN pg_namespace n ON c.relnamespace = n.oid
		WHERE n.nspname = 'public'
		AND t.tgname LIKE '%cache_invalidate%'
		AND NOT t.tgisinternal
		ORDER BY c.relname, t.tgname
	`)
	if err != nil {
		log.Fatal("Failed to query triggers:", err)
	}
	defer rows.Close()

	fmt.Println("   Triggers found:")
	for rows.Next() {
		var tableName, triggerName, status string
		if err := rows.Scan(&tableName, &triggerName, &status); err != nil {
			log.Fatal("Failed to scan trigger:", err)
		}
		fmt.Printf("   - %s on %s (%s)\n", triggerName, tableName, status)
	}
	if err := rows.Err(); err != nil {
		log.Fatal("Error iterating triggers:", err)
	}

	fmt.Println("\n✅ All migration checks passed! Triggers are correctly installed.")
}



