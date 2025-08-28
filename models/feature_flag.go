package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// FeatureFlag represents a feature flag/toggle
type FeatureFlag struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Key         string    `json:"key" gorm:"uniqueIndex;not null;size:100"`
	Description string    `json:"description" gorm:"size:500"`
	Enabled     bool      `json:"enabled" gorm:"default:false"`
	Value       string    `json:"value,omitempty" gorm:"size:255"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName returns the table name for FeatureFlag
func (FeatureFlag) TableName() string {
	return "feature_flags"
}

// GetFeatureFlags retrieves all feature flags
func GetFeatureFlags(db *gorm.DB) ([]*FeatureFlag, error) {
	var flags []*FeatureFlag
	if err := db.Find(&flags).Error; err != nil {
		return nil, err
	}
	return flags, nil
}

// GetFeatureFlag retrieves a feature flag by key
func GetFeatureFlag(db *gorm.DB, key string) (*FeatureFlag, error) {
	var flag FeatureFlag
	if err := db.Where("key = ?", key).First(&flag).Error; err != nil {
		return nil, err
	}
	return &flag, nil
}

// CreateFeatureFlag creates a new feature flag
func CreateFeatureFlag(db *gorm.DB, key, description string, enabled bool, value string) (*FeatureFlag, error) {
	flag := &FeatureFlag{
		ID:          uuid.New(),
		Key:         key,
		Description: description,
		Enabled:     enabled,
		Value:       value,
	}

	if err := db.Create(flag).Error; err != nil {
		return nil, err
	}

	return flag, nil
}

// UpdateFeatureFlag updates an existing feature flag
func UpdateFeatureFlag(db *gorm.DB, key, description string, enabled bool, value string) (*FeatureFlag, error) {
	updates := map[string]interface{}{
		"description": description,
		"enabled":     enabled,
		"updated_at":  time.Now(),
	}

	if value != "" {
		updates["value"] = value
	}

	if err := db.Model(&FeatureFlag{}).Where("key = ?", key).Updates(updates).Error; err != nil {
		return nil, err
	}

	return GetFeatureFlag(db, key)
}

// DeleteFeatureFlag deletes a feature flag
func DeleteFeatureFlag(db *gorm.DB, key string) error {
	return db.Where("key = ?", key).Delete(&FeatureFlag{}).Error
}

// BeforeCreate is a GORM hook that runs before creating a feature flag
func (f *FeatureFlag) BeforeCreate(tx *gorm.DB) error {
	if f.ID == uuid.Nil {
		f.ID = uuid.New()
	}
	return nil
}

// BeforeUpdate is a GORM hook that runs before updating a feature flag
func (f *FeatureFlag) BeforeUpdate(tx *gorm.DB) error {
	f.UpdatedAt = time.Now()
	return nil
}
