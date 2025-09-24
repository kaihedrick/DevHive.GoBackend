package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"devhive-backend/internal/config"
	"devhive-backend/internal/http/router"
	"devhive-backend/internal/repo"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Setup logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Connect to database
	db, err := initDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Run migrations (temporarily disabled)
	// if err := runMigrations(db); err != nil {
	//	log.Fatal("Failed to run migrations:", err)
	// }

	// Create pgx pool for sqlc
	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to create pgx pool:", err)
	}
	defer pool.Close()

	// Initialize repository
	queries := repo.New(pool)

	// Setup router
	r := router.Setup(cfg, queries)

	// Start server
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Server shutdown failed: %v", err)
		}
	}()

	log.Printf("Starting server on port %s", cfg.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("Server failed:", err)
	}
}

func initDB(databaseURL string) (*sql.DB, error) {
	config, err := pgx.ParseConfig(databaseURL)
	if err != nil {
		return nil, err
	}

	db := stdlib.OpenDB(*config)
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
