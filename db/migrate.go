package db

import (
	"log"

	"devhive-backend/models"

	"gorm.io/gorm"
)

// AutoMigrate runs automatic database migrations
func AutoMigrate(db *gorm.DB) error {
	log.Println("Starting database migrations...")

	// Migrate all models
	err := db.AutoMigrate(
		&models.User{},
		&models.Project{},
		&models.ProjectMember{},
		&models.Sprint{},
		&models.Task{},
		&models.Message{},
		&models.PasswordReset{},
		&models.FeatureFlag{},
	)

	if err != nil {
		log.Printf("Migration failed: %v", err)
		return err
	}

	log.Println("Database migrations completed successfully")
	return nil
}

// CreateIndexes creates additional database indexes for performance
func CreateIndexes(db *gorm.DB) error {
	log.Println("Creating database indexes...")

	// User indexes
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)").Error; err != nil {
		log.Printf("Failed to create users email index: %v", err)
	}

	// Project indexes
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_projects_owner_id ON projects(project_owner_id)").Error; err != nil {
		log.Printf("Failed to create projects owner index: %v", err)
	}

	// Project member indexes
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_project_members_user_id ON project_members(user_id)").Error; err != nil {
		log.Printf("Failed to create project members user index: %v", err)
	}
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_project_members_project_id ON project_members(project_id)").Error; err != nil {
		log.Printf("Failed to create project members project index: %v", err)
	}

	// Sprint indexes
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_sprints_project_id ON sprints(project_id)").Error; err != nil {
		log.Printf("Failed to create sprints project index: %v", err)
	}

	// Task indexes
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_tasks_project_id ON tasks(project_id)").Error; err != nil {
		log.Printf("Failed to create tasks project index: %v", err)
	}
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_tasks_sprint_id ON tasks(sprint_id)").Error; err != nil {
		log.Printf("Failed to create tasks sprint index: %v", err)
	}
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_tasks_assignee_id ON tasks(assignee_id)").Error; err != nil {
		log.Printf("Failed to create tasks assignee index: %v", err)
	}

	// Message indexes
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_messages_project_id ON messages(project_id)").Error; err != nil {
		log.Printf("Failed to create messages project index: %v", err)
	}
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_messages_sender_id ON messages(sender_id)").Error; err != nil {
		log.Printf("Failed to create messages sender index: %v", err)
	}

	log.Println("Database indexes created successfully")
	return nil
}

// SeedData seeds the database with initial data if needed
func SeedData(db *gorm.DB) error {
	log.Println("Seeding database with initial data...")

	// Check if we already have users
	var userCount int64
	db.Model(&models.User{}).Count(&userCount)
	if userCount > 0 {
		log.Println("Database already has data, skipping seeding")
		return nil
	}

	// Create a default admin user
	defaultUser := &models.User{
		Email:     "admin@devhive.app",
		Username:  "admin",
		FirstName: "Admin",
		LastName:  "User",
		Active:    true,
	}

	if err := db.Create(defaultUser).Error; err != nil {
		log.Printf("Failed to create default user: %v", err)
		return err
	}

	// Create a default project
	defaultProject := &models.Project{
		Name:           "Welcome to DevHive",
		Description:    "This is your first project. Start by creating tasks and organizing your work!",
		ProjectOwnerID: defaultUser.ID,
	}

	if err := db.Create(defaultProject).Error; err != nil {
		log.Printf("Failed to create default project: %v", err)
		return err
	}

	// Add the admin user as a member of the default project
	projectMember := &models.ProjectMember{
		ProjectID: defaultProject.ID,
		UserID:    defaultUser.ID,
	}

	if err := db.Create(projectMember).Error; err != nil {
		log.Printf("Failed to create project member: %v", err)
		return err
	}

	log.Println("Database seeded successfully")
	return nil
}

// ResetDatabase drops all tables and recreates them (USE WITH CAUTION)
func ResetDatabase(db *gorm.DB) error {
	log.Println("WARNING: Resetting database - all data will be lost!")

	// Drop all tables
	if err := db.Migrator().DropTable(
		&models.User{},
		&models.Project{},
		&models.ProjectMember{},
		&models.Sprint{},
		&models.Task{},
		&models.Message{},
		&models.PasswordReset{},
		&models.FeatureFlag{},
	); err != nil {
		return err
	}

	// Recreate tables
	if err := AutoMigrate(db); err != nil {
		return err
	}

	// Seed with initial data
	if err := SeedData(db); err != nil {
		return err
	}

	log.Println("Database reset completed successfully")
	return nil
}
