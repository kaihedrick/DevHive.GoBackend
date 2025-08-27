package controllers

import (
	"net/http"

	"devhive-backend/internal/flags"

	"github.com/gin-gonic/gin"
)

// GetFeatureFlags returns all feature flags
func GetFeatureFlags(c *gin.Context) {
	flags, err := flags.GlobalManager.GetAllFlags()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve feature flags"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"flags": flags,
		"count": len(flags),
	})
}

// GetFeatureFlag returns a specific feature flag
func GetFeatureFlag(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Feature flag key is required"})
		return
	}

	enabled := flags.IsEnabledGlobal(key)
	c.JSON(http.StatusOK, gin.H{
		"key":     key,
		"enabled": enabled,
	})
}

// UpdateFeatureFlag updates a feature flag
func UpdateFeatureFlag(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Feature flag key is required"})
		return
	}

	var request struct {
		Enabled     *bool   `json:"enabled"`
		Description *string `json:"description"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Update enabled status if provided
	if request.Enabled != nil {
		if err := flags.GlobalManager.SetFlag(key, *request.Enabled); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update feature flag"})
			return
		}
	}

	// Update description if provided
	if request.Description != nil {
		if err := flags.GlobalManager.CreateFlag(key, *request.Description, flags.IsEnabledGlobal(key)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update feature flag description"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Feature flag updated successfully",
		"key":     key,
		"enabled": flags.IsEnabledGlobal(key),
	})
}

// CreateFeatureFlag creates a new feature flag
func CreateFeatureFlag(c *gin.Context) {
	var request struct {
		Key         string `json:"key" binding:"required"`
		Description string `json:"description"`
		Enabled     bool   `json:"enabled"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := flags.GlobalManager.CreateFlag(request.Key, request.Description, request.Enabled); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create feature flag"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Feature flag created successfully",
		"key":     request.Key,
		"enabled": request.Enabled,
	})
}

// DeleteFeatureFlag deletes a feature flag
func DeleteFeatureFlag(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Feature flag key is required"})
		return
	}

	// Note: This would require adding a DeleteFlag method to the flags package
	// For now, we'll disable the flag instead
	if err := flags.GlobalManager.SetFlag(key, false); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to disable feature flag"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Feature flag disabled successfully",
		"key":     key,
	})
}

// BulkUpdateFeatureFlags updates multiple feature flags at once
func BulkUpdateFeatureFlags(c *gin.Context) {
	var request struct {
		Flags []struct {
			Key     string `json:"key" binding:"required"`
			Enabled bool   `json:"enabled"`
		} `json:"flags" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	results := make([]gin.H, 0, len(request.Flags))
	for _, flag := range request.Flags {
		if err := flags.GlobalManager.SetFlag(flag.Key, flag.Enabled); err != nil {
			results = append(results, gin.H{
				"key":     flag.Key,
				"success": false,
				"error":   err.Error(),
			})
		} else {
			results = append(results, gin.H{
				"key":     flag.Key,
				"success": true,
				"enabled": flag.Enabled,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"results": results,
		"total":   len(request.Flags),
	})
}
