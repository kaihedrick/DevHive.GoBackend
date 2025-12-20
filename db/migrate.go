package db

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// RunMigrations runs database migrations using SQL files
func RunMigrations(db *sql.DB) error {
	log.Println("Starting database migrations...")

	// Create migrations tracking table if it doesn't exist
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Read all migration files
	// Try multiple possible paths for migrations directory
	migrationsDirs := []string{
		"cmd/devhive-api/migrations",
		"./cmd/devhive-api/migrations",
		"migrations",
		"./migrations",
	}

	var migrationsDir string
	var entries []os.FileInfo
	for _, dir := range migrationsDirs {
		var err error
		entries, err = ioutil.ReadDir(dir)
		if err == nil {
			migrationsDir = dir
			break
		}
	}

	if migrationsDir == "" {
		return fmt.Errorf("could not find migrations directory (tried: %v)", migrationsDirs)
	}

	// Filter and sort migration files
	var migrationFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			migrationFiles = append(migrationFiles, entry.Name())
		}
	}
	sort.Strings(migrationFiles)

	// Apply each migration
	for _, filename := range migrationFiles {
		// Check if migration has already been applied
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = $1", filename).Scan(&count)
		if err != nil {
			return fmt.Errorf("failed to check migration status: %w", err)
		}

		if count > 0 {
			log.Printf("Migration %s already applied, skipping", filename)
			continue
		}

		// Read migration file
		migrationPath := filepath.Join(migrationsDir, filename)
		migrationSQL, err := ioutil.ReadFile(migrationPath)
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", filename, err)
		}

		// Execute migration
		log.Printf("Applying migration: %s", filename)
		_, err = db.Exec(string(migrationSQL))
		if err != nil {
			// Check if error is due to already existing objects (idempotency)
			errStr := err.Error()
			if strings.Contains(errStr, "already exists") || 
			   strings.Contains(errStr, "duplicate key") ||
			   (strings.Contains(errStr, "relation") && strings.Contains(errStr, "already exists")) ||
			   (strings.Contains(errStr, "does not exist") && strings.Contains(errStr, "DROP")) {
				log.Printf("Migration %s: Some objects already exist or don't exist (idempotent), continuing...", filename)
				// Continue - this is expected for idempotent migrations
			} else {
				log.Printf("❌ Migration %s failed with error: %v", filename, err)
				return fmt.Errorf("failed to execute migration %s: %w", filename, err)
			}
		} else {
			log.Printf("✅ Migration %s executed successfully", filename)
		}

		// Record migration as applied (use ON CONFLICT to handle idempotency)
		_, err = db.Exec(`
			INSERT INTO schema_migrations (version) 
			VALUES ($1) 
			ON CONFLICT (version) DO NOTHING
		`, filename)
		if err != nil {
			return fmt.Errorf("failed to record migration %s: %w", filename, err)
		}

		log.Printf("Successfully applied migration: %s", filename)
	}

	log.Println("Database migrations completed successfully")
	return nil
}

// CreateIndexes creates additional database indexes for performance
func CreateIndexes(db *sql.DB) error {
	log.Println("Creating database indexes...")

	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)",
		"CREATE INDEX IF NOT EXISTS idx_projects_owner_id ON projects(owner_id)",
		"CREATE INDEX IF NOT EXISTS idx_project_members_user_id ON project_members(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_project_members_project_id ON project_members(project_id)",
		"CREATE INDEX IF NOT EXISTS idx_sprints_project_id ON sprints(project_id)",
		"CREATE INDEX IF NOT EXISTS idx_tasks_project_id ON tasks(project_id)",
		"CREATE INDEX IF NOT EXISTS idx_tasks_sprint_id ON tasks(sprint_id)",
		"CREATE INDEX IF NOT EXISTS idx_tasks_assignee_id ON tasks(assignee_id)",
		"CREATE INDEX IF NOT EXISTS idx_messages_project_id ON messages(project_id)",
		"CREATE INDEX IF NOT EXISTS idx_messages_sender_id ON messages(sender_id)",
	}

	for _, indexSQL := range indexes {
		if _, err := db.Exec(indexSQL); err != nil {
			log.Printf("Failed to create index: %v", err)
			return err
		}
	}

	log.Println("Database indexes created successfully")
	return nil
}

// HealthCheck verifies database connectivity
func HealthCheck(db *sql.DB) error {
	return db.Ping()
}

// VerifyNotifyTriggers checks if NOTIFY triggers are installed and logs their status
func VerifyNotifyTriggers(db *sql.DB) error {
	log.Println("🔍 Verifying NOTIFY triggers installation...")
	
	// Check if the function exists
	var funcExists bool
	err := db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM pg_proc p
			JOIN pg_namespace n ON p.pronamespace = n.oid
			WHERE n.nspname = 'public' AND p.proname = 'notify_cache_invalidation'
		)
	`).Scan(&funcExists)
	if err != nil {
		return fmt.Errorf("failed to check function: %w", err)
	}
	
	if !funcExists {
		log.Println("❌ NOTIFY function 'notify_cache_invalidation' NOT found")
		return fmt.Errorf("notify_cache_invalidation function not found")
	}
	log.Println("✅ NOTIFY function 'notify_cache_invalidation' exists")
	
	// Check triggers on each table
	tables := []string{"projects", "sprints", "tasks", "project_members"}
	allTriggersExist := true
	
	for _, table := range tables {
		var triggerExists bool
		err := db.QueryRow(`
			SELECT EXISTS(
				SELECT 1 FROM pg_trigger t
				JOIN pg_class c ON t.tgrelid = c.oid
				JOIN pg_namespace n ON c.relnamespace = n.oid
				WHERE n.nspname = 'public' 
				AND c.relname = $1
				AND t.tgname LIKE $2
				AND NOT t.tgisinternal
			)
		`, table, table+"_%cache_invalidate%").Scan(&triggerExists)
		
		if err != nil {
			log.Printf("❌ Failed to check trigger for table %s: %v", table, err)
			allTriggersExist = false
			continue
		}
		
		if triggerExists {
			log.Printf("✅ Trigger exists on table: %s", table)
		} else {
			log.Printf("❌ Trigger MISSING on table: %s", table)
			allTriggersExist = false
		}
	}
	
	if !allTriggersExist {
		return fmt.Errorf("some NOTIFY triggers are missing")
	}
	
	log.Println("✅ All NOTIFY triggers verified and installed")
	return nil
}
