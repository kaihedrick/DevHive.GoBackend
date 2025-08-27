package controllers

import (
	"net/http"
	"strconv"
	"time"

	"devhive-backend/db"
	"devhive-backend/internal/flags"
	"devhive-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// MobileSprint represents a simplified sprint structure for mobile clients
type MobileSprint struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	Status      string    `json:"status"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	DaysLeft    int       `json:"days_left"`
	Progress    float64   `json:"progress"`
}

// MobileProject represents a simplified project structure for mobile clients
type MobileProject struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	Status      string    `json:"status"`
	MemberCount int       `json:"member_count"`
	ActiveSprint *MobileSprint `json:"active_sprint,omitempty"`
}

// MobileMessage represents a simplified message structure for mobile clients
type MobileMessage struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	Sender    string    `json:"sender"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
}

// GetMobileSprints retrieves sprints optimized for mobile clients
func GetMobileSprints(c *gin.Context) {
	// Check if mobile API is enabled
	if !flags.IsEnabledGlobal("mobile_v2_api") {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Mobile API is not enabled"})
		return
	}

	userID := GetCurrentUserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	projectIDStr := c.Param("projectId")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Check if user has access to the project
	isMember, err := models.IsProjectMember(db.GetDB(), projectID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if !isMember {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to project"})
		return
	}

	// Get sprints from database
	sprints, err := models.GetSprints(db.GetDB(), projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve sprints"})
		return
	}

	// Convert to mobile format
	mobileSprints := make([]MobileSprint, len(sprints))
	for i, sprint := range sprints {
		mobileSprints[i] = convertToMobileSprint(sprint)
	}

	c.JSON(http.StatusOK, gin.H{
		"sprints": mobileSprints,
		"count":   len(mobileSprints),
	})
}

// GetMobileProjects retrieves projects optimized for mobile clients
func GetMobileProjects(c *gin.Context) {
	// Check if mobile API is enabled
	if !flags.IsEnabledGlobal("mobile_v2_api") {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Mobile API is not enabled"})
		return
	}

	userID := GetCurrentUserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get projects from database
	projects, err := models.GetProjects(db.GetDB(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve projects"})
		return
	}

	// Convert to mobile format
	mobileProjects := make([]MobileProject, len(projects))
	for i, project := range projects {
		mobileProjects[i] = convertToMobileProject(project)
	}

	c.JSON(http.StatusOK, gin.H{
		"projects": mobileProjects,
		"count":    len(mobileProjects),
	})
}

// GetMobileProject retrieves a specific project optimized for mobile clients
func GetMobileProject(c *gin.Context) {
	// Check if mobile API is enabled
	if !flags.IsEnabledGlobal("mobile_v2_api") {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Mobile API is not enabled"})
		return
	}

	userID := GetCurrentUserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	projectIDStr := c.Param("id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Check if user has access to the project
	isMember, err := models.IsProjectMember(db.GetDB(), projectID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if !isMember {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to project"})
		return
	}

	// Get project details
	project, err := models.GetProject(db.GetDB(), projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	// Get active sprint if any
	var activeSprint *MobileSprint
	if active, err := models.GetActiveSprint(db.GetDB(), projectID); err == nil && active != nil {
		sprint := convertToMobileSprint(active)
		activeSprint = &sprint
	}

	mobileProject := convertToMobileProject(project)
	mobileProject.ActiveSprint = activeSprint

	c.JSON(http.StatusOK, gin.H{
		"project": mobileProject,
	})
}

// GetMobileMessages retrieves messages optimized for mobile clients
func GetMobileMessages(c *gin.Context) {
	// Check if mobile API is enabled
	if !flags.IsEnabledGlobal("mobile_v2_api") {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Mobile API is not enabled"})
		return
	}

	userID := GetCurrentUserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	projectIDStr := c.Param("projectId")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Check if user has access to the project
	isMember, err := models.IsProjectMember(db.GetDB(), projectID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if !isMember {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to project"})
		return
	}

	// Get query parameters for pagination
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 50
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Get messages from database
	messages, err := models.GetMessages(db.GetDB(), projectID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve messages"})
		return
	}

	// Convert to mobile format
	mobileMessages := make([]MobileMessage, len(messages))
	for i, message := range messages {
		mobileMessages[i] = convertToMobileMessage(message)
	}

	c.JSON(http.StatusOK, gin.H{
		"messages": mobileMessages,
		"pagination": gin.H{
			"total":    len(messages),
			"limit":    limit,
			"offset":   offset,
			"has_more": len(messages) == limit,
		},
	})
}

// Helper functions to convert models to mobile format

func convertToMobileSprint(sprint *models.Sprint) MobileSprint {
	now := time.Now()
	daysLeft := 0
	progress := 0.0

	if sprint.Status == "active" {
		if sprint.EndDate.After(now) {
			daysLeft = int(sprint.EndDate.Sub(now).Hours() / 24)
		}
		
		// Calculate progress based on time elapsed
		totalDuration := sprint.EndDate.Sub(sprint.StartDate)
		elapsed := now.Sub(sprint.StartDate)
		if totalDuration > 0 && elapsed > 0 {
			progress = float64(elapsed) / float64(totalDuration) * 100
			if progress > 100 {
				progress = 100
			}
		}
	}

	return MobileSprint{
		ID:          sprint.ID.String(),
		Name:        sprint.Name,
		Description: sprint.Description,
		Status:      sprint.Status,
		StartDate:   sprint.StartDate,
		EndDate:     sprint.EndDate,
		DaysLeft:    daysLeft,
		Progress:    progress,
	}
}

func convertToMobileProject(project *models.Project) MobileProject {
	memberCount := 0
	if project.Members != nil {
		memberCount = len(project.Members)
	}

	return MobileProject{
		ID:          project.ID.String(),
		Name:        project.Name,
		Description: project.Description,
		Status:      project.Status,
		MemberCount: memberCount,
	}
}

func convertToMobileMessage(message *models.Message) MobileMessage {
	senderName := "Unknown"
	if message.Sender != nil {
		senderName = message.Sender.FirstName + " " + message.Sender.LastName
	}

	return MobileMessage{
		ID:        message.ID.String(),
		Content:   message.Content,
		Sender:    senderName,
		Type:      message.MessageType,
		CreatedAt: message.CreatedAt,
	}
}
