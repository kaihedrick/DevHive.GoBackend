package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"

	"devhive-backend/internal/router"
	"devhive-backend/internal/ws"
	"devhive-backend/pkg/config"
	"devhive-backend/pkg/db"
)

func main() {
	if err := config.LoadEnv(); err != nil {
		log.Fatal("Error loading .env file:", err)
	}

	if err := db.InitDB(); err != nil {
		log.Fatal("Error initializing database:", err)
	}
	defer db.CloseDB()

	if err := config.InitFirebase(); err != nil {
		log.Printf("Warning: Firebase initialization failed: %v", err)
	}

	// Start WebSocket hub
	ws.StartWebSocketHub()

	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery(), gzip.Gzip(gzip.DefaultCompression))

	router.Register(r)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	log.Printf("Starting DevHive Backend on port %s", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("Failed to start server:", err)
	}
}
