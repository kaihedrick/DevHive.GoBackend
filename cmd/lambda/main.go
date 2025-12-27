package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"devhive-backend/db"
	"devhive-backend/internal/broadcast"
	"devhive-backend/internal/config"
	"devhive-backend/internal/http/router"
	"devhive-backend/internal/repo"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

var httpAdapter *httpadapter.HandlerAdapterV2

func init() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Connect to database (Neon)
	database, err := initDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Run database migrations
	if err := db.RunMigrations(database); err != nil {
		log.Printf("Warning: Migration failed: %v", err)
	}

	// Create pgx pool for sqlc
	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to create pgx pool:", err)
	}

	// Initialize repository
	queries := repo.New(pool)

	// Initialize broadcast client for WebSocket notifications
	broadcast.Init()

	// Setup router (pass nil for hub since WebSockets aren't supported in Lambda)
	// Real-time updates are handled via AWS API Gateway WebSocket API + broadcaster Lambda
	r := router.Setup(cfg, queries, database, nil)

	// Create the HTTP adapter for API Gateway HTTP API (v2)
	httpAdapter = httpadapter.NewV2(http.Handler(r))
}

func Handler(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	return httpAdapter.ProxyWithContext(ctx, req)
}

func main() {
	lambda.Start(Handler)
}

func initDB(databaseURL string) (*sql.DB, error) {
	config, err := pgx.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	database := stdlib.OpenDB(*config)
	if err := database.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Lower connection pool for Lambda (each invocation should be lightweight)
	database.SetMaxOpenConns(5)
	database.SetMaxIdleConns(2)

	return database, nil
}
