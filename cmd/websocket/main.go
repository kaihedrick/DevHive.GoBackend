package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/apigatewaymanagementapi"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/golang-jwt/jwt/v5"
)

var (
	dynamoClient *dynamodb.Client
	tableName    string
	jwtSecret    string
)

// Connection represents a WebSocket connection stored in DynamoDB
type Connection struct {
	ConnectionID string `dynamodbav:"connectionId"`
	UserID       string `dynamodbav:"userId"`
	ProjectID    string `dynamodbav:"projectId,omitempty"` // omitempty: DynamoDB GSI keys can't be empty strings
	ConnectedAt  string `dynamodbav:"connectedAt"`
	TTL          int64  `dynamodbav:"ttl"`
}

// WebSocketMessage represents an incoming WebSocket message
type WebSocketMessage struct {
	Action    string          `json:"action"`
	ProjectID string          `json:"projectId,omitempty"`
	Data      json.RawMessage `json:"data,omitempty"`
}

// BroadcastMessage represents a message to broadcast to clients
type BroadcastMessage struct {
	Type      string      `json:"type"`
	ProjectID string      `json:"projectId"`
	Data      interface{} `json:"data"`
	Timestamp string      `json:"timestamp"`
}

func init() {
	tableName = os.Getenv("CONNECTIONS_TABLE")
	jwtSecret = os.Getenv("JWT_SIGNING_KEY")

	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}

	dynamoClient = dynamodb.NewFromConfig(cfg)
}

func handler(ctx context.Context, request events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	routeKey := request.RequestContext.RouteKey
	connectionID := request.RequestContext.ConnectionID

	log.Printf("Route: %s, ConnectionID: %s", routeKey, connectionID)

	switch routeKey {
	case "$connect":
		return handleConnect(ctx, request)
	case "$disconnect":
		return handleDisconnect(ctx, connectionID)
	case "$default":
		return handleDefault(ctx, request)
	case "subscribe":
		return handleSubscribe(ctx, request)
	case "unsubscribe":
		return handleUnsubscribe(ctx, request)
	default:
		return handleDefault(ctx, request)
	}
}

func handleConnect(ctx context.Context, request events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	connectionID := request.RequestContext.ConnectionID

	// Extract token from query string or header
	token := request.QueryStringParameters["token"]
	if token == "" {
		if auth := request.Headers["Authorization"]; strings.HasPrefix(auth, "Bearer ") {
			token = strings.TrimPrefix(auth, "Bearer ")
		}
	}

	if token == "" {
		log.Printf("Connection rejected: no token provided")
		return events.APIGatewayProxyResponse{StatusCode: 401, Body: "Unauthorized: missing token"}, nil
	}

	// Validate JWT and extract user ID
	userID, err := validateToken(token)
	if err != nil {
		log.Printf("Connection rejected: invalid token: %v", err)
		return events.APIGatewayProxyResponse{StatusCode: 401, Body: "Unauthorized: invalid token"}, nil
	}

	// Get project ID from query string (optional at connect time)
	projectID := request.QueryStringParameters["projectId"]
	if projectID == "" {
		// Fallback to project_id for backward compatibility
		projectID = request.QueryStringParameters["project_id"]
	}

	// Store connection in DynamoDB
	conn := Connection{
		ConnectionID: connectionID,
		UserID:       userID,
		ProjectID:    projectID,
		ConnectedAt:  time.Now().UTC().Format(time.RFC3339),
		TTL:          time.Now().Add(24 * time.Hour).Unix(), // Auto-expire after 24 hours
	}

	item, err := attributevalue.MarshalMap(conn)
	if err != nil {
		log.Printf("Failed to marshal connection: %v", err)
		return events.APIGatewayProxyResponse{StatusCode: 500, Body: "Internal error"}, nil
	}

	_, err = dynamoClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      item,
	})
	if err != nil {
		log.Printf("Failed to store connection: %v", err)
		return events.APIGatewayProxyResponse{StatusCode: 500, Body: "Internal error"}, nil
	}

	log.Printf("Connection established: user=%s, project=%s, connectionId=%s", userID, projectID, connectionID)
	return events.APIGatewayProxyResponse{StatusCode: 200, Body: "Connected"}, nil
}

func handleDisconnect(ctx context.Context, connectionID string) (events.APIGatewayProxyResponse, error) {
	// Remove connection from DynamoDB
	_, err := dynamoClient.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"connectionId": &types.AttributeValueMemberS{Value: connectionID},
		},
	})
	if err != nil {
		log.Printf("Failed to delete connection: %v", err)
	}

	log.Printf("Connection closed: connectionId=%s", connectionID)
	return events.APIGatewayProxyResponse{StatusCode: 200, Body: "Disconnected"}, nil
}

func handleSubscribe(ctx context.Context, request events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	connectionID := request.RequestContext.ConnectionID

	var msg WebSocketMessage
	if err := json.Unmarshal([]byte(request.Body), &msg); err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 400, Body: "Invalid message format"}, nil
	}

	if msg.ProjectID == "" {
		return events.APIGatewayProxyResponse{StatusCode: 400, Body: "project_id required"}, nil
	}

	// Update connection with project ID
	_, err := dynamoClient.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"connectionId": &types.AttributeValueMemberS{Value: connectionID},
		},
		UpdateExpression: aws.String("SET projectId = :pid"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pid": &types.AttributeValueMemberS{Value: msg.ProjectID},
		},
	})
	if err != nil {
		log.Printf("Failed to update connection project: %v", err)
		return events.APIGatewayProxyResponse{StatusCode: 500, Body: "Internal error"}, nil
	}

	log.Printf("Subscribed to project: connectionId=%s, projectId=%s", connectionID, msg.ProjectID)
	return events.APIGatewayProxyResponse{StatusCode: 200, Body: "Subscribed"}, nil
}

func handleUnsubscribe(ctx context.Context, request events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	connectionID := request.RequestContext.ConnectionID

	// Clear project ID from connection
	_, err := dynamoClient.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"connectionId": &types.AttributeValueMemberS{Value: connectionID},
		},
		UpdateExpression: aws.String("REMOVE projectId"),
	})
	if err != nil {
		log.Printf("Failed to clear project from connection: %v", err)
		return events.APIGatewayProxyResponse{StatusCode: 500, Body: "Internal error"}, nil
	}

	log.Printf("Unsubscribed from project: connectionId=%s", connectionID)
	return events.APIGatewayProxyResponse{StatusCode: 200, Body: "Unsubscribed"}, nil
}

func handleDefault(ctx context.Context, request events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Echo back or handle custom messages
	log.Printf("Received message: %s", request.Body)

	var msg WebSocketMessage
	if err := json.Unmarshal([]byte(request.Body), &msg); err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 400, Body: "Invalid message format"}, nil
	}

	// Handle ping/pong for keep-alive
	if msg.Action == "ping" {
		return sendToConnection(ctx, request.RequestContext, request.RequestContext.ConnectionID, map[string]string{
			"action": "pong",
		})
	}

	return events.APIGatewayProxyResponse{StatusCode: 200, Body: "OK"}, nil
}

func validateToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})
	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", fmt.Errorf("invalid token claims")
	}

	// Extract user_id from claims (handle both "sub" and "user_id" fields)
	if userID, ok := claims["user_id"].(string); ok {
		return userID, nil
	}
	if sub, ok := claims["sub"].(string); ok {
		return sub, nil
	}

	return "", fmt.Errorf("user_id not found in token")
}

func sendToConnection(ctx context.Context, reqCtx events.APIGatewayWebsocketProxyRequestContext, connectionID string, data interface{}) (events.APIGatewayProxyResponse, error) {
	endpoint := fmt.Sprintf("https://%s/%s", reqCtx.DomainName, reqCtx.Stage)

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500}, err
	}

	apiClient := apigatewaymanagementapi.NewFromConfig(cfg, func(o *apigatewaymanagementapi.Options) {
		o.BaseEndpoint = aws.String(endpoint)
	})

	payload, _ := json.Marshal(data)

	_, err = apiClient.PostToConnection(ctx, &apigatewaymanagementapi.PostToConnectionInput{
		ConnectionId: aws.String(connectionID),
		Data:         payload,
	})
	if err != nil {
		log.Printf("Failed to send to connection %s: %v", connectionID, err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, err
	}

	return events.APIGatewayProxyResponse{StatusCode: 200}, nil
}

func main() {
	lambda.Start(handler)
}
