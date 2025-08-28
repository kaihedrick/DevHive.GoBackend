package models

import (
	"time"

	"github.com/google/uuid"
)

// PaginationInfo provides pagination metadata for mobile API responses
type PaginationInfo struct {
	CurrentPage  int  `json:"currentPage"`
	TotalPages   int  `json:"totalPages"`
	TotalItems   int  `json:"totalItems"`
	ItemsPerPage int  `json:"itemsPerPage"`
	HasNext      bool `json:"hasNext"`
	HasPrev      bool `json:"hasPrev"`
}

// MobileProject represents a project optimized for mobile consumption
type MobileProject struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	OwnerID     uuid.UUID `json:"ownerId"`
	OwnerName   string    `json:"ownerName"`
	MemberCount int       `json:"memberCount"`
	SprintCount int       `json:"sprintCount"`
	TaskCount   int       `json:"taskCount"`
	IsOwner     bool      `json:"isOwner"`
	IsMember    bool      `json:"isMember"`
}

// MobileProjectsResponse represents the response for mobile projects list
type MobileProjectsResponse struct {
	Projects   []MobileProject `json:"projects"`
	Pagination PaginationInfo  `json:"pagination"`
}

// MobileProjectResponse represents the response for a single mobile project
type MobileProjectResponse struct {
	Project MobileProject `json:"project"`
}

// MobileSprint represents a sprint optimized for mobile consumption
type MobileSprint struct {
	ID                 uuid.UUID `json:"id"`
	Name               string    `json:"name"`
	Description        string    `json:"description"`
	Status             string    `json:"status"`
	StartDate          time.Time `json:"startDate"`
	EndDate            time.Time `json:"endDate"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
	TaskCount          int       `json:"taskCount"`
	CompletedTaskCount int       `json:"completedTaskCount"`
	Progress           float64   `json:"progress"`
}

// MobileSprintsResponse represents the response for mobile sprints list
type MobileSprintsResponse struct {
	Sprints    []MobileSprint `json:"sprints"`
	Pagination PaginationInfo `json:"pagination"`
}

// MobileMessage represents a message optimized for mobile consumption
type MobileMessage struct {
	ID         uuid.UUID `json:"id"`
	Content    string    `json:"content"`
	Type       string    `json:"type"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
	UserID     uuid.UUID `json:"userId"`
	UserName   string    `json:"userName"`
	UserAvatar string    `json:"userAvatar,omitempty"`
}

// MobileMessagesResponse represents the response for mobile messages list
type MobileMessagesResponse struct {
	Messages   []MobileMessage `json:"messages"`
	Pagination PaginationInfo  `json:"pagination"`
}
