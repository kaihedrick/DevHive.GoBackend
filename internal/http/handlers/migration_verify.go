package handlers

import (
	"net/http"

	"devhive-backend/internal/http/response"
)

// VerifyMigration verifies that a specific migration was applied correctly
func (h *MigrationHandler) VerifyMigration(w http.ResponseWriter, r *http.Request) {
	migrationName := r.URL.Query().Get("name")
	if migrationName == "" {
		migrationName = "004_add_cache_invalidation_triggers.sql"
	}

	results := make(map[string]interface{})

	// 1. Check if migration was recorded
	var migrationRecorded bool
	err := h.db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM schema_migrations WHERE version = $1
		)
	`, migrationName).Scan(&migrationRecorded)
	if err != nil {
		response.InternalServerError(w, "Failed to check migration record: "+err.Error())
		return
	}
	results["migration_recorded"] = migrationRecorded

	// 2. Check if function exists (for cache invalidation migration)
	if migrationName == "004_add_cache_invalidation_triggers.sql" {
		var functionExists bool
		err = h.db.QueryRow(`
			SELECT EXISTS(
				SELECT 1 FROM pg_proc p
				JOIN pg_namespace n ON p.pronamespace = n.oid
				WHERE n.nspname = 'public' AND p.proname = 'notify_cache_invalidation'
			)
		`).Scan(&functionExists)
		if err != nil {
			response.InternalServerError(w, "Failed to check function: "+err.Error())
			return
		}
		results["function_exists"] = functionExists

		// 3. Check triggers
		triggers := make(map[string]bool)
		tables := []string{"projects", "sprints", "tasks", "project_members"}
		for _, table := range tables {
			var triggerExists bool
			err = h.db.QueryRow(`
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
			if err == nil {
				triggers[table] = triggerExists
			}
		}
		results["triggers"] = triggers

		// Count total triggers
		var triggerCount int
		err = h.db.QueryRow(`
			SELECT COUNT(*) FROM pg_trigger t
			JOIN pg_class c ON t.tgrelid = c.oid
			JOIN pg_namespace n ON c.relnamespace = n.oid
			WHERE n.nspname = 'public'
			AND t.tgname LIKE '%cache_invalidate%'
			AND NOT t.tgisinternal
		`).Scan(&triggerCount)
		if err == nil {
			results["trigger_count"] = triggerCount
		}
	}

	// Determine overall status
	allPassed := migrationRecorded
	if migrationName == "004_add_cache_invalidation_triggers.sql" {
		if funcExists, ok := results["function_exists"].(bool); ok {
			allPassed = allPassed && funcExists
		}
		if triggerCount, ok := results["trigger_count"].(int); ok {
			allPassed = allPassed && (triggerCount == 4)
		}
	}

	results["status"] = "passed"
	if !allPassed {
		results["status"] = "failed"
	}

	response.JSON(w, http.StatusOK, results)
}

