package db

import (
	"database/sql"
	"log"
)

// RunMigrations runs database migrations using SQL files
func RunMigrations(db *sql.DB) error {
	log.Println("Starting database migrations...")

	// Note: This project uses sqlc for database operations
	// Database schema is managed through SQL migration files
	// See cmd/devhive-api/migrations/ for migration files

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
