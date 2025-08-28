package controllers

import (
	"net/http"
	"strconv"

	"devhive-backend/models"
	"devhive-backend/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// MobileController handles mobile-specific API endpoints
type MobileController struct {
	projectService services.ProjectService
	sprintService  services.SprintService
	messageService services.MessageService
	userService    services.UserService
}

// NewMobileController creates a new mobile controller instance
func NewMobileController(
	projectService services.ProjectService,
	sprintService services.SprintService,
	messageService services.MessageService,
	userService services.UserService,
) *MobileController {
	return &MobileController{
		projectService: projectService,
		sprintService:  sprintService,
		messageService: messageService,
		userService:    userService,
	}
}

// GetMobileProjects godoc
// @Summary Get projects for mobile app
// @Description Retrieves a list of projects optimized for mobile consumption
// @Tags mobile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 20, max: 100)"
// @Param search query string false "Search term for project names"
// @Success 200 {object} models.MobileProjectsResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /mobile/v2/projects [get]
func (mc *MobileController) GetMobileProjects(c *gin.Context) {
	// Get query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	search := c.Query("search")

	// Validate pagination
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "User not authenticated",
		})
		return
	}

	// Get projects with mobile optimization
	projects, total, err := mc.projectService.GetProjectsForMobile(userID.(string), page, limit, search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to retrieve projects",
		})
		return
	}

	// Calculate pagination info
	totalPages := (total + limit - 1) / limit
	hasNext := page < totalPages
	hasPrev := page > 1

	c.JSON(http.StatusOK, models.MobileProjectsResponse{
		Projects: projects,
		Pagination: models.PaginationInfo{
			CurrentPage:  page,
			TotalPages:   totalPages,
			TotalItems:   total,
			ItemsPerPage: limit,
			HasNext:      hasNext,
			HasPrev:      hasPrev,
		},
	})
}

// GetMobileProject godoc
// @Summary Get project details for mobile app
// @Description Retrieves detailed project information optimized for mobile consumption
// @Tags mobile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Project ID"
// @Success 200 {object} models.MobileProjectResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /mobile/v2/projects/{id} [get]
func (mc *MobileController) GetMobileProject(c *gin.Context) {
	projectID := c.Param("id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Project ID is required",
		})
		return
	}

	// Parse project ID
	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid project ID format",
		})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "User not authenticated",
		})
		return
	}

	// Get project with mobile optimization
	project, err := mc.projectService.GetProjectForMobile(projectUUID, userID.(string))
	if err != nil {
		if err.Error() == "project not found" {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: "Project not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to retrieve project",
		})
		return
	}

	c.JSON(http.StatusOK, models.MobileProjectResponse{
		Project: *project,
	})
}

// GetMobileSprints godoc
// @Summary Get sprints for mobile app
// @Description Retrieves a list of sprints for a specific project optimized for mobile consumption
// @Tags mobile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Project ID"
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 20, max: 100)"
// @Param status query string false "Filter by sprint status (active, completed, planned)"
// @Success 200 {object} models.MobileSprintsResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /mobile/v2/projects/{id}/sprints [get]
func (mc *MobileController) GetMobileSprints(c *gin.Context) {
	projectID := c.Param("id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Project ID is required",
		})
		return
	}

	// Parse project ID
	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid project ID format",
		})
		return
	}

	// Get query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	status := c.Query("status")

	// Validate pagination
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "User not authenticated",
		})
		return
	}

	// Get sprints with mobile optimization
	sprints, total, err := mc.sprintService.GetSprintsForMobile(projectUUID, userID.(string), page, limit, status)
	if err != nil {
		if err.Error() == "project not found" {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: "Project not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to retrieve sprints",
		})
		return
	}

	// Calculate pagination info
	totalPages := (total + limit - 1) / limit
	hasNext := page < totalPages
	hasPrev := page > 1

	c.JSON(http.StatusOK, models.MobileSprintsResponse{
		Sprints: sprints,
		Pagination: models.PaginationInfo{
			CurrentPage:  page,
			TotalPages:   totalPages,
			TotalItems:   total,
			ItemsPerPage: limit,
			HasNext:      hasNext,
			HasPrev:      hasPrev,
		},
	})
}

// GetMobileMessages godoc
// @Summary Get messages for mobile app
// @Description Retrieves a list of messages for a specific project optimized for mobile consumption
// @Tags mobile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Project ID"
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 20, max: 100)"
// @Param search query string false "Search term for message content"
// @Success 200 {object} models.MobileMessagesResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /mobile/v2/projects/{id}/messages [get]
func (mc *MobileController) GetMobileMessages(c *gin.Context) {
	projectID := c.Param("id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Project ID is required",
		})
		return
	}

	// Parse project ID
	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid project ID format",
		})
		return
	}

	// Get query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	search := c.Query("search")

	// Validate pagination
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "User not authenticated",
		})
		return
	}

	// Get messages with mobile optimization
	messages, total, err := mc.messageService.GetMessagesForMobile(projectUUID, userID.(string), page, limit, search)
	if err != nil {
		if err.Error() == "project not found" {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: "Project not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to retrieve messages",
		})
		return
	}

	// Calculate pagination info
	totalPages := (total + limit - 1) / limit
	hasNext := page < totalPages
	hasPrev := page > 1

	c.JSON(http.StatusOK, models.MobileMessagesResponse{
		Messages: messages,
		Pagination: models.PaginationInfo{
			CurrentPage:  page,
			TotalPages:   totalPages,
			TotalItems:   total,
			ItemsPerPage: limit,
			HasNext:      hasNext,
			HasPrev:      hasPrev,
		},
	})
}
