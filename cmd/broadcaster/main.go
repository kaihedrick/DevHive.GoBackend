package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/apigatewaymanagementapi"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

var (
	dynamoClient    *dynamodb.Client
	tableName       string
	wsEndpoint      string
)

// BroadcastEvent is the event received by this Lambda
type BroadcastEvent struct {
	ProjectID string      `json:"projectId"`
	Type      string      `json:"type"` // e.g., "task_created", "message_created", "cache_invalidate"
	Data      interface{} `json:"data"`
	ExcludeConnectionID string `json:"exclude_connection_id,omitempty"` // Don't send to this connection
}

// Connection represents a WebSocket connection from DynamoDB
type Connection struct {
	ConnectionID string `dynamodbav:"connectionId"`
	UserID       string `dynamodbav:"userId"`
	ProjectID    string `dynamodbav:"projectId"`
}

// OutgoingMessage is the message sent to WebSocket clients
type OutgoingMessage struct {
	Type      string      `json:"type"`
	ProjectID string      `json:"projectId"`
	Data      interface{} `json:"data"`
	Timestamp string      `json:"timestamp"`
}

func init() {
	tableName = os.Getenv("CONNECTIONS_TABLE")
	wsEndpoint = os.Getenv("WEBSOCKET_ENDPOINT")

	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}

	dynamoClient = dynamodb.NewFromConfig(cfg)
}

func handler(ctx context.Context, event BroadcastEvent) error {
	log.Printf("Broadcasting to project %s: type=%s", event.ProjectID, event.Type)

	// Query connections for this project using GSI
	connections, err := getProjectConnections(ctx, event.ProjectID)
	if err != nil {
		log.Printf("Failed to get connections: %v", err)
		return err
	}

	if len(connections) == 0 {
		log.Printf("No connections for project %s", event.ProjectID)
		return nil
	}

	// Create API Gateway Management API client
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return err
	}

	apiClient := apigatewaymanagementapi.NewFromConfig(cfg, func(o *apigatewaymanagementapi.Options) {
		o.BaseEndpoint = aws.String(wsEndpoint)
	})

	// Prepare message
	msg := OutgoingMessage{
		Type:      event.Type,
		ProjectID: event.ProjectID,
		Data:      event.Data,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	payload, _ := json.Marshal(msg)

	// Send to all connections
	var staleConnections []string
	successCount := 0

	for _, conn := range connections {
		// Skip excluded connection
		if conn.ConnectionID == event.ExcludeConnectionID {
			continue
		}

		_, err := apiClient.PostToConnection(ctx, &apigatewaymanagementapi.PostToConnectionInput{
			ConnectionId: aws.String(conn.ConnectionID),
			Data:         payload,
		})
		if err != nil {
			// Connection is stale, mark for cleanup
			log.Printf("Failed to send to %s: %v", conn.ConnectionID, err)
			staleConnections = append(staleConnections, conn.ConnectionID)
		} else {
			successCount++
		}
	}

	// Clean up stale connections
	for _, connID := range staleConnections {
		_, _ = dynamoClient.DeleteItem(ctx, &dynamodb.DeleteItemInput{
			TableName: aws.String(tableName),
			Key: map[string]types.AttributeValue{
				"connectionId": &types.AttributeValueMemberS{Value: connID},
			},
		})
	}

	log.Printf("Broadcast complete: %d delivered, %d stale removed", successCount, len(staleConnections))
	return nil
}

func getProjectConnections(ctx context.Context, projectID string) ([]Connection, error) {
	// Query using GSI on projectId
	result, err := dynamoClient.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(tableName),
		IndexName:              aws.String("projectId-index"),
		KeyConditionExpression: aws.String("projectId = :pid"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pid": &types.AttributeValueMemberS{Value: projectID},
		},
	})
	if err != nil {
		return nil, err
	}

	var connections []Connection
	err = attributevalue.UnmarshalListOfMaps(result.Items, &connections)
	return connections, err
}

func main() {
	lambda.Start(handler)
}
