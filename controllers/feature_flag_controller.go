package controllers

import (
	"net/http"

	"devhive-backend/db"
	"devhive-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// FeatureFlagRequest represents the request to create/update a feature flag
type FeatureFlagRequest struct {
	Key         string `json:"key" binding:"required" example:"new_ui"`
	Description string `json:"description" binding:"required" example:"Enable new UI design"`
	Enabled     bool   `json:"enabled" example:"true"`
	Value       string `json:"value,omitempty" example:"v2"`
}

// BulkUpdateFeatureFlagRequest represents the request to bulk update feature flags
type BulkUpdateFeatureFlagRequest struct {
	Flags []FeatureFlagRequest `json:"flags" binding:"required"`
}

// GetFeatureFlags retrieves all feature flags
// @Summary Get all feature flags
// @Description Retrieves all feature flags (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.FeatureFlag "List of feature flags"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 403 {object} models.ErrorResponse "Access denied"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /admin/feature-flags [get]
func GetFeatureFlags(c *gin.Context) {
	userID := GetCurrentUserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "User not authenticated",
		})
		return
	}

	// TODO: Check if user is admin
	// For now, allow any authenticated user

	flags, err := models.GetFeatureFlags(db.GetDB())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to retrieve feature flags",
		})
		return
	}

	c.JSON(http.StatusOK, flags)
}

// GetFeatureFlag retrieves a specific feature flag
// @Summary Get feature flag
// @Description Retrieves a specific feature flag by key (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Param key path string true "Feature flag key"
// @Security BearerAuth
// @Success 200 {object} models.FeatureFlag "Feature flag details"
// @Failure 400 {object} models.ErrorResponse "Bad request"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 403 {object} models.ErrorResponse "Access denied"
// @Failure 404 {object} models.ErrorResponse "Feature flag not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /admin/feature-flags/{key} [get]
func GetFeatureFlag(c *gin.Context) {
	userID := GetCurrentUserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "User not authenticated",
		})
		return
	}

	// TODO: Check if user is admin
	// For now, allow any authenticated user

	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Feature flag key is required",
		})
		return
	}

	flag, err := models.GetFeatureFlag(db.GetDB(), key)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "Feature flag not found",
		})
		return
	}

	c.JSON(http.StatusOK, flag)
}

// CreateFeatureFlag creates a new feature flag
// @Summary Create feature flag
// @Description Creates a new feature flag (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Param flag body FeatureFlagRequest true "Feature flag to create"
// @Security BearerAuth
// @Success 201 {object} models.FeatureFlag "Feature flag created successfully"
// @Failure 400 {object} models.ErrorResponse "Bad request"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 403 {object} models.ErrorResponse "Access denied"
// @Failure 409 {object} models.ErrorResponse "Feature flag already exists"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /admin/feature-flags [post]
func CreateFeatureFlag(c *gin.Context) {
	userID := GetCurrentUserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "User not authenticated",
		})
		return
	}

	// TODO: Check if user is admin
	// For now, allow any authenticated user

	var req FeatureFlagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid request body",
		})
		return
	}

	flag, err := models.CreateFeatureFlag(db.GetDB(), req.Key, req.Description, req.Enabled, req.Value)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to create feature flag",
		})
		return
	}

	c.JSON(http.StatusCreated, flag)
}

// UpdateFeatureFlag updates an existing feature flag
// @Summary Update feature flag
// @Description Updates an existing feature flag (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Param key path string true "Feature flag key"
// @Param flag body FeatureFlagRequest true "Updated feature flag data"
// @Security BearerAuth
// @Success 200 {object} models.FeatureFlag "Feature flag updated successfully"
// @Failure 400 {object} models.ErrorResponse "Bad request"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 403 {object} models.ErrorResponse "Access denied"
// @Failure 404 {object} models.ErrorResponse "Feature flag not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /admin/feature-flags/{key} [put]
func UpdateFeatureFlag(c *gin.Context) {
	userID := GetCurrentUserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "User not authenticated",
		})
		return
	}

	// TODO: Check if user is admin
	// For now, allow any authenticated user

	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Feature flag key is required",
		})
		return
	}

	var req FeatureFlagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid request body",
		})
		return
	}

	flag, err := models.UpdateFeatureFlag(db.GetDB(), key, req.Description, req.Enabled, req.Value)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "Feature flag not found",
		})
		return
	}

	c.JSON(http.StatusOK, flag)
}

// DeleteFeatureFlag deletes a feature flag
// @Summary Delete feature flag
// @Description Deletes a feature flag (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Param key path string true "Feature flag key"
// @Security BearerAuth
// @Success 204 "Feature flag deleted successfully"
// @Failure 400 {object} models.ErrorResponse "Bad request"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 403 {object} models.ErrorResponse "Access denied"
// @Failure 404 {object} models.ErrorResponse "Feature flag not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /admin/feature-flags/{key} [delete]
func DeleteFeatureFlag(c *gin.Context) {
	userID := GetCurrentUserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "User not authenticated",
		})
		return
	}

	// TODO: Check if user is admin
	// For now, allow any authenticated user

	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Feature flag key is required",
		})
		return
	}

	err := models.DeleteFeatureFlag(db.GetDB(), key)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "Feature flag not found",
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// BulkUpdateFeatureFlags bulk updates multiple feature flags
// @Summary Bulk update feature flags
// @Description Bulk updates multiple feature flags (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Param flags body BulkUpdateFeatureFlagRequest true "Feature flags to update"
// @Security BearerAuth
// @Success 200 {array} models.FeatureFlag "Feature flags updated successfully"
// @Failure 400 {object} models.ErrorResponse "Bad request"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 403 {object} models.ErrorResponse "Access denied"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /admin/feature-flags/bulk-update [post]
func BulkUpdateFeatureFlags(c *gin.Context) {
	userID := GetCurrentUserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "User not authenticated",
		})
		return
	}

	// TODO: Check if user is admin
	// For now, allow any authenticated user

	var req BulkUpdateFeatureFlagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid request body",
		})
		return
	}

	var updatedFlags []*models.FeatureFlag
	for _, flagReq := range req.Flags {
		flag, err := models.UpdateFeatureFlag(db.GetDB(), flagReq.Key, flagReq.Description, flagReq.Enabled, flagReq.Value)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: "Failed to update feature flag: " + flagReq.Key,
			})
			return
		}
		updatedFlags = append(updatedFlags, flag)
	}

	c.JSON(http.StatusOK, updatedFlags)
}
