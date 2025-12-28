package broadcast

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
)

// Client handles broadcasting messages to WebSocket connections
type Client struct {
	lambdaClient *lambda.Client
	functionName string
	enabled      bool
}

// BroadcastEvent is the payload sent to the broadcaster Lambda
type BroadcastEvent struct {
	ProjectID           string      `json:"projectId"`
	Type                string      `json:"type"`
	Data                interface{} `json:"data"`
	ExcludeConnectionID string      `json:"exclude_connection_id,omitempty"`
}

var defaultClient *Client

// Init initializes the broadcast client (call during Lambda init)
func Init() {
	functionName := os.Getenv("BROADCASTER_FUNCTION_NAME")
	if functionName == "" {
		log.Println("BROADCASTER_FUNCTION_NAME not set, broadcasting disabled")
		defaultClient = &Client{enabled: false}
		return
	}

	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Printf("Failed to load AWS config for broadcaster: %v", err)
		defaultClient = &Client{enabled: false}
		return
	}

	defaultClient = &Client{
		lambdaClient: lambda.NewFromConfig(cfg),
		functionName: functionName,
		enabled:      true,
	}
	log.Printf("Broadcaster client initialized: %s", functionName)
}

// Send broadcasts a message to all connections subscribed to a project
func Send(ctx context.Context, projectID, eventType string, data interface{}) error {
	if defaultClient == nil || !defaultClient.enabled {
		return nil // Broadcasting disabled, silently skip
	}
	return defaultClient.Send(ctx, projectID, eventType, data, "")
}

// SendExcluding broadcasts a message to all connections except the specified one
func SendExcluding(ctx context.Context, projectID, eventType string, data interface{}, excludeConnID string) error {
	if defaultClient == nil || !defaultClient.enabled {
		return nil
	}
	return defaultClient.Send(ctx, projectID, eventType, data, excludeConnID)
}

// Send broadcasts a message to WebSocket connections
func (c *Client) Send(ctx context.Context, projectID, eventType string, data interface{}, excludeConnID string) error {
	if !c.enabled {
		return nil
	}

	event := BroadcastEvent{
		ProjectID:           projectID,
		Type:                eventType,
		Data:                data,
		ExcludeConnectionID: excludeConnID,
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	// Invoke broadcaster Lambda asynchronously (Event invocation type)
	_, err = c.lambdaClient.Invoke(ctx, &lambda.InvokeInput{
		FunctionName:   aws.String(c.functionName),
		Payload:        payload,
		InvocationType: "Event", // Async invocation
	})
	if err != nil {
		log.Printf("Failed to invoke broadcaster: %v", err)
		return err
	}

	log.Printf("Broadcast sent: project=%s, type=%s", projectID, eventType)
	return nil
}

// Common event types
const (
	EventTaskCreated     = "task_created"
	EventTaskUpdated     = "task_updated"
	EventTaskDeleted     = "task_deleted"
	EventSprintCreated   = "sprint_created"
	EventSprintUpdated   = "sprint_updated"
	EventSprintDeleted   = "sprint_deleted"
	EventMessageCreated  = "message_created"
	EventProjectUpdated  = "project_updated"
	EventMemberAdded     = "member_added"
	EventMemberRemoved   = "member_removed"
	EventCacheInvalidate = "cache_invalidate"
)
