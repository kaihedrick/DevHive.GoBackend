package db

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"devhive-backend/internal/ws"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// NotifyListener handles PostgreSQL NOTIFY events for cache invalidation
type NotifyListener struct {
	databaseURL string
	hub         *ws.Hub
	conn        *pgx.Conn
	ctx         context.Context
	cancel      context.CancelFunc
}

// CacheInvalidationPayload represents the notification payload from PostgreSQL
type CacheInvalidationPayload struct {
	Resource  string `json:"resource"`
	ID        string `json:"id"`
	Action    string `json:"action"`
	ProjectID string `json:"projectId"`
	Timestamp string `json:"timestamp"`
}

// StartNotifyListener creates a dedicated connection and starts listening for NOTIFY events
// CRITICAL: Uses dedicated connection, NOT from pool
func StartNotifyListener(databaseURL string, hub *ws.Hub) {
	log.Println("🔧 StartNotifyListener called - initializing NOTIFY listener...")
	if databaseURL == "" {
		log.Println("❌ ERROR: databaseURL is empty! NOTIFY listener cannot start.")
		return
	}
	if hub == nil {
		log.Println("❌ ERROR: hub is nil! NOTIFY listener cannot start.")
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	listener := &NotifyListener{
		databaseURL: databaseURL,
		hub:         hub,
		ctx:         ctx,
		cancel:      cancel,
	}

	log.Println("🔧 Starting NOTIFY listener goroutine...")
	go listener.listen()
	log.Println("✅ StartNotifyListener completed - goroutine launched")
}

// listen establishes connection and listens for notifications
func (l *NotifyListener) listen() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("❌ PANIC in NOTIFY listener: %v", r)
		}
		if l.conn != nil {
			l.conn.Close(context.Background())
		}
		log.Println("❌ NOTIFY listener goroutine exiting")
	}()

	log.Println("🚀 NOTIFY listener goroutine started")

	backoff := time.Second
	maxBackoff := 30 * time.Second

	for {
		log.Printf("🔄 NOTIFY listener attempting to connect to database...")
		// Create dedicated connection (NOT from pool)
		conn, err := pgx.Connect(l.ctx, l.databaseURL)
		if err != nil {
			log.Printf("❌ Failed to connect for NOTIFY listener: %v. Retrying in %v...", err, backoff)
			select {
			case <-l.ctx.Done():
				return
			case <-time.After(backoff):
				backoff = min(backoff*2, maxBackoff)
				continue
			}
		}

		l.conn = conn
		backoff = time.Second // Reset backoff on successful connection

		// Listen to the cache_invalidate channel
		_, err = conn.Exec(l.ctx, "LISTEN cache_invalidate")
		if err != nil {
			log.Printf("Failed to LISTEN on cache_invalidate: %v", err)
			conn.Close(l.ctx)
			select {
			case <-l.ctx.Done():
				return
			case <-time.After(backoff):
				continue
			}
		}

		log.Println("✅ NOTIFY listener started, listening on 'cache_invalidate' channel")
		log.Printf("✅ NOTIFY listener connection established successfully at %s", time.Now().Format(time.RFC3339))

		// Notify clients of reconnection (if this is a reconnect)
		if l.conn != nil {
			l.hub.BroadcastToAll("reconnect", map[string]string{"reason": "notify_reconnect"})
		}

		// Main loop: wait for notifications
		log.Println("⏳ NOTIFY listener waiting for notifications on 'cache_invalidate' channel...")
		log.Printf("⏳ NOTIFY listener ready at %s - any PostgreSQL NOTIFY events will be logged here", time.Now().Format(time.RFC3339))
		for {
			notification, err := conn.WaitForNotification(l.ctx)
			if err != nil {
				if err == context.Canceled {
					log.Println("NOTIFY listener context canceled, shutting down")
					return
				}
				log.Printf("❌ Error waiting for notification: %v. Reconnecting...", err)
				conn.Close(l.ctx)
				break // Break inner loop to reconnect
			}

			// Handle notification
			log.Printf("📨 Notification received at %s", time.Now().Format(time.RFC3339))
			l.handleNotification(notification)
		}
	}
}

// handleNotification parses the notification payload and broadcasts to WebSocket hub
func (l *NotifyListener) handleNotification(notification *pgconn.Notification) {
	log.Printf("🔔 RAW NOTIFY received from PostgreSQL: channel=%s, payload=%s", notification.Channel, notification.Payload)

	var payload CacheInvalidationPayload
	if err := json.Unmarshal([]byte(notification.Payload), &payload); err != nil {
		log.Printf("❌ Failed to parse notification payload: %v. Payload: %s", err, notification.Payload)
		return
	}

	log.Printf("✅ NOTIFY received: resource=%s, action=%s, project_id=%s, id=%s",
		payload.Resource, payload.Action, payload.ProjectID, payload.ID)

	// Validate payload has required fields
	if payload.ProjectID == "" {
		log.Printf("ERROR: Notification missing project_id, skipping. Full payload: %+v", payload)
		return
	}

	// Special logging for project_members to help debug
	if payload.Resource == "project_members" {
		log.Printf("Project member notification: action=%s, project_id=%s, member_id=%s",
			payload.Action, payload.ProjectID, payload.ID)

		// Log WebSocket connection status BEFORE broadcasting
		total, matching, users := l.hub.GetProjectConnections(payload.ProjectID)
		log.Printf("WebSocket status BEFORE broadcast for project %s: total_clients=%d, matching=%d, users=%v",
			payload.ProjectID, total, matching, users)
	}

	// Build message data
	messageData := map[string]interface{}{
		"resource":   payload.Resource,
		"id":         payload.ID,
		"action":     payload.Action,
		"project_id": payload.ProjectID,
		"timestamp":  payload.Timestamp,
	}

	// For project deletions and creations, broadcast to ALL clients (not just project-scoped)
	// This ensures all users' project lists are invalidated
	// Note: Resource is now normalized to singular ('project', not 'projects')
	if payload.Resource == "project" && (payload.Action == "DELETE" || payload.Action == "INSERT") {
		l.hub.BroadcastToAll("cache_invalidate", messageData)
		log.Printf("Broadcasted project %s cache invalidation to all clients: project_id=%s", payload.Action, payload.ProjectID)
	} else if payload.Resource == "project_members" {
		// Broadcast member changes to ALL clients to ensure owners see updates
		// even if they're not actively viewing the project WebSocket
		l.hub.BroadcastToAll("cache_invalidate", messageData)
		log.Printf("Broadcasted project_members %s cache invalidation to all clients: project_id=%s", payload.Action, payload.ProjectID)

		// Log WebSocket connection status AFTER broadcasting
		total, matching, users := l.hub.GetProjectConnections(payload.ProjectID)
		log.Printf("WebSocket status AFTER broadcast for project %s: total_clients=%d, matching=%d, users=%v",
			payload.ProjectID, total, matching, users)
	} else {
		// For other changes, broadcast project-scoped
		l.hub.BroadcastToProject(payload.ProjectID, "cache_invalidate", messageData)
		log.Printf("Broadcasted cache invalidation: resource=%s, action=%s, project_id=%s", payload.Resource, payload.Action, payload.ProjectID)
	}
}

// Stop stops the NOTIFY listener
func (l *NotifyListener) Stop() {
	if l.cancel != nil {
		l.cancel()
	}
	if l.conn != nil {
		l.conn.Close(context.Background())
	}
}

// min returns the minimum of two durations
func min(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}
