package flags

import (
	"database/sql"
	"log"
	"sync"
	"time"
)

// FeatureFlag represents a feature flag in the system
type FeatureFlag struct {
	Key         string    `json:"key"`
	Enabled     bool      `json:"enabled"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Manager manages feature flags with caching
type Manager struct {
	db       *sql.DB
	cache    map[string]bool
	cacheTTL time.Duration
	lastSync time.Time
	mutex    sync.RWMutex
}

// NewManager creates a new feature flag manager
func NewManager(db *sql.DB) *Manager {
	manager := &Manager{
		db:       db,
		cache:    make(map[string]bool),
		cacheTTL: 5 * time.Minute, // Cache for 5 minutes
	}
	
	// Initial sync
	manager.syncFlags()
	
	// Start background sync
	go manager.backgroundSync()
	
	return manager
}

// IsEnabled checks if a feature flag is enabled
func (m *Manager) IsEnabled(key string) bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	// Check cache first
	if enabled, exists := m.cache[key]; exists {
		return enabled
	}
	
	// Fallback to database check
	return m.checkDatabase(key)
}

// IsEnabled checks if a feature flag is enabled (static function for backward compatibility)
func IsEnabled(db *sql.DB, key string) bool {
	var enabled bool
	err := db.QueryRow("SELECT enabled FROM feature_flags WHERE key = $1", key).Scan(&enabled)
	if err != nil {
		log.Printf("Error checking feature flag %s: %v", key, err)
		return false
	}
	return enabled
}

// checkDatabase checks the database for a feature flag
func (m *Manager) checkDatabase(key string) bool {
	var enabled bool
	err := m.db.QueryRow("SELECT enabled FROM feature_flags WHERE key = $1", key).Scan(&enabled)
	if err != nil {
		log.Printf("Error checking feature flag %s: %v", key, err)
		return false
	}
	return enabled
}

// syncFlags synchronizes all feature flags from the database
func (m *Manager) syncFlags() {
	rows, err := m.db.Query("SELECT key, enabled FROM feature_flags")
	if err != nil {
		log.Printf("Error syncing feature flags: %v", err)
		return
	}
	defer rows.Close()
	
	newCache := make(map[string]bool)
	for rows.Next() {
		var key string
		var enabled bool
		if err := rows.Scan(&key, &enabled); err != nil {
			log.Printf("Error scanning feature flag: %v", err)
			continue
		}
		newCache[key] = enabled
	}
	
	m.mutex.Lock()
	m.cache = newCache
	m.lastSync = time.Now()
	m.mutex.Unlock()
	
	log.Printf("Synced %d feature flags", len(newCache))
}

// backgroundSync runs periodic sync of feature flags
func (m *Manager) backgroundSync() {
	ticker := time.NewTicker(m.cacheTTL)
	defer ticker.Stop()
	
	for range ticker.C {
		m.syncFlags()
	}
}

// SetFlag enables or disables a feature flag
func (m *Manager) SetFlag(key string, enabled bool) error {
	_, err := m.db.Exec("UPDATE feature_flags SET enabled = $1 WHERE key = $2", enabled, key)
	if err != nil {
		return err
	}
	
	// Update cache immediately
	m.mutex.Lock()
	m.cache[key] = enabled
	m.mutex.Unlock()
	
	return nil
}

// CreateFlag creates a new feature flag
func (m *Manager) CreateFlag(key, description string, enabled bool) error {
	_, err := m.db.Exec(`
		INSERT INTO feature_flags (key, description, enabled) 
		VALUES ($1, $2, $3) 
		ON CONFLICT (key) DO UPDATE SET 
			description = EXCLUDED.description, 
			enabled = EXCLUDED.enabled
	`, key, description, enabled)
	
	if err != nil {
		return err
	}
	
	// Update cache
	m.mutex.Lock()
	m.cache[key] = enabled
	m.mutex.Unlock()
	
	return nil
}

// GetAllFlags returns all feature flags
func (m *Manager) GetAllFlags() ([]FeatureFlag, error) {
	rows, err := m.db.Query(`
		SELECT key, enabled, description, created_at, updated_at 
		FROM feature_flags 
		ORDER BY key
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var flags []FeatureFlag
	for rows.Next() {
		var flag FeatureFlag
		err := rows.Scan(&flag.Key, &flag.Enabled, &flag.Description, &flag.CreatedAt, &flag.UpdatedAt)
		if err != nil {
			log.Printf("Error scanning feature flag: %v", err)
			continue
		}
		flags = append(flags, flag)
	}
	
	return flags, nil
}

// Global manager instance
var GlobalManager *Manager

// InitGlobalManager initializes the global feature flag manager
func InitGlobalManager(db *sql.DB) {
	GlobalManager = NewManager(db)
}

// IsEnabledGlobal checks if a feature flag is enabled using the global manager
func IsEnabledGlobal(key string) bool {
	if GlobalManager == nil {
		return false
	}
	return GlobalManager.IsEnabled(key)
}
