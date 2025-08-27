package controllers

import (
	"net/http"
	"strings"

	"devhive-backend/config"
	"devhive-backend/db"
	"devhive-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// AuthMiddleware creates a middleware for JWT authentication
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Check if the header starts with "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		// Extract the token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Parse and validate the token
		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.AppConfig.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Extract claims
		claims, ok := token.Claims.(*Claims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		// Parse user ID
		userID, err := uuid.Parse(claims.UserID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID in token"})
			c.Abort()
			return
		}

		// Get user from database
		user, err := models.GetUserByID(db.GetDB(), userID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			c.Abort()
			return
		}

		// Set user in context
		c.Set("user", user)
		c.Set("userID", userID)

		c.Next()
	}
}

// GetCurrentUser returns the current authenticated user from context
func GetCurrentUser(c *gin.Context) *models.User {
	user, exists := c.Get("user")
	if !exists {
		return nil
	}
	return user.(*models.User)
}

// GetCurrentUserID returns the current authenticated user ID from context
func GetCurrentUserID(c *gin.Context) uuid.UUID {
	userID, exists := c.Get("userID")
	if !exists {
		return uuid.Nil
	}
	return userID.(uuid.UUID)
}

// RequireProjectAccess creates a middleware that requires project access
func RequireProjectAccess(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := GetCurrentUserID(c)
		if userID == uuid.Nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		// Get project ID from URL parameter
		projectIDStr := c.Param("projectId")
		if projectIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID required"})
			c.Abort()
			return
		}

		projectID, err := uuid.Parse(projectIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
			c.Abort()
			return
		}

		// Check if user is a member of the project
		isMember, err := models.IsProjectMember(db.GetDB(), projectID, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			c.Abort()
			return
		}

		if !isMember {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to project"})
			c.Abort()
			return
		}

		// If role requirement is specified, check user's role
		if requiredRole != "" {
			userRole, err := models.GetProjectMemberRole(db.GetDB(), projectID, userID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
				c.Abort()
				return
			}

			// Check if user has required role
			if !hasRequiredRole(userRole, requiredRole) {
				c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
				c.Abort()
				return
			}
		}

		// Set project ID in context
		c.Set("projectID", projectID)

		c.Next()
	}
}

// hasRequiredRole checks if a user's role meets the required role
func hasRequiredRole(userRole, requiredRole string) bool {
	roleHierarchy := map[string]int{
		"viewer": 1,
		"member": 2,
		"admin":  3,
		"owner":  4,
	}

	userLevel, userExists := roleHierarchy[userRole]
	requiredLevel, requiredExists := roleHierarchy[requiredRole]

	if !userExists || !requiredExists {
		return false
	}

	return userLevel >= requiredLevel
}
