package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Add proper origin checking in production
	},
}

// Client represents a connected WebSocket client
type Client struct {
	conn      *websocket.Conn
	userID    string
	projectID string
	send      chan []byte
	hub       *Hub
}

// Hub manages all WebSocket connections
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	Register   chan *Client // Exported for external registration
	unregister chan *Client
	mutex      sync.RWMutex
}

// Message represents a WebSocket message
type Message struct {
	Type      string      `json:"type"`
	Resource  string      `json:"resource,omitempty"` // Resource type: project, sprint, task, project_member
	Action    string      `json:"action,omitempty"`   // Action: INSERT, UPDATE, DELETE
	Data      interface{} `json:"data"`
	ProjectID string      `json:"projectId,omitempty"`
	UserID    string      `json:"userId,omitempty"`
}

// NewHub creates a new WebSocket hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mutex.Lock()
			h.clients[client] = true
			h.mutex.Unlock()
			log.Printf("Client registered. Total clients: %d", len(h.clients))

		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mutex.Unlock()
			log.Printf("Client unregistered. Total clients: %d", len(h.clients))

		case message := <-h.broadcast:
			h.mutex.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mutex.RUnlock()
		}
	}
}

// BroadcastToProject sends a message to all clients in a specific project
func (h *Hub) BroadcastToProject(projectID string, messageType string, data interface{}) {
	msg := Message{
		Type:      messageType,
		Data:      data,
		ProjectID: projectID,
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	h.mutex.RLock()
	for client := range h.clients {
		if client.projectID == projectID {
			select {
			case client.send <- msgBytes:
			default:
				close(client.send)
				delete(h.clients, client)
			}
		}
	}
	h.mutex.RUnlock()
}

// HandleConnections upgrades HTTP connections to WebSocket and manages client lifecycle
func HandleConnections(hub *Hub, w http.ResponseWriter, r *http.Request) {
	// Extract user ID and project ID from query parameters or headers
	userID := r.URL.Query().Get("user_id")
	projectID := r.URL.Query().Get("project_id")

	if userID == "" || projectID == "" {
		http.Error(w, "Missing user_id or project_id", http.StatusBadRequest)
		return
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := &Client{
		conn:      ws,
		userID:    userID,
		projectID: projectID,
		send:      make(chan []byte, 256),
		hub:       hub,
	}

	client.hub.Register <- client

	// Start goroutines for reading and writing
	go client.WritePump()
	go client.ReadPump()
}

// ReadPump pumps messages from the WebSocket connection to the hub (exported)
func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512) // Max message size
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket read error: %v", err)
			}
			break
		}

		// Handle incoming messages if needed
		var msg Message
		if err := json.Unmarshal(message, &msg); err == nil {
			// Process message based on type
			switch msg.Type {
			case "ping":
				// Send pong response
				pongMsg := Message{Type: "pong"}
				if pongBytes, err := json.Marshal(pongMsg); err == nil {
					c.send <- pongBytes
				}
			}
		}
	}
}

// WritePump pumps messages from the hub to the WebSocket connection (exported)
func (c *Client) WritePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Global hub instance
var GlobalHub = NewHub()

// StartWebSocketHub starts the global WebSocket hub
func StartWebSocketHub() {
	go GlobalHub.Run()
}

// BroadcastSprintUpdate broadcasts sprint updates to all clients in a project
func BroadcastSprintUpdate(projectID string, sprintData interface{}) {
	GlobalHub.BroadcastToProject(projectID, "sprint_update", sprintData)
}

// BroadcastProjectUpdate broadcasts project updates to all clients in a project
func BroadcastProjectUpdate(projectID string, projectData interface{}) {
	GlobalHub.BroadcastToProject(projectID, "project_update", projectData)
}

// BroadcastMessageUpdate broadcasts message updates to all clients in a project
func BroadcastMessageUpdate(projectID string, messageData interface{}) {
	GlobalHub.BroadcastToProject(projectID, "message_update", messageData)
}

// AuthenticatedHandleConnections handles WebSocket connections with JWT authentication
func AuthenticatedHandleConnections(hub *Hub, w http.ResponseWriter, r *http.Request) {
	// Extract JWT token from query parameter or header
	token := r.URL.Query().Get("token")
	if token == "" {
		token = r.Header.Get("Authorization")
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}
	}

	if token == "" {
		http.Error(w, "Missing authentication token", http.StatusUnauthorized)
		return
	}

	// Validate JWT token and extract user info
	userID, err := validateJWTToken(token)
	if err != nil {
		log.Printf("JWT validation error: %v", err)
		http.Error(w, "Invalid authentication token", http.StatusUnauthorized)
		return
	}

	// Extract project ID from query parameters
	projectID := r.URL.Query().Get("project_id")
	if projectID == "" {
		http.Error(w, "Missing project_id", http.StatusBadRequest)
		return
	}

	// Validate project access
	if !validateProjectAccess(userID, projectID) {
		http.Error(w, "Access denied to project", http.StatusForbidden)
		return
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := &Client{
		conn:      ws,
		userID:    userID,
		projectID: projectID,
		send:      make(chan []byte, 256),
		hub:       hub,
	}

	client.hub.Register <- client

	// Start goroutines for reading and writing
	go client.WritePump()
	go client.ReadPump()
}

// validateJWTToken validates a JWT token and returns the user ID
func validateJWTToken(tokenString string) (string, error) {
	// Import the JWT package and validate token
	// This would integrate with your existing JWT validation logic
	// For now, returning a placeholder implementation
	return "user-123", nil // Replace with actual JWT validation
}

// validateProjectAccess checks if a user has access to a project
func validateProjectAccess(userID, projectID string) bool {
	// This would integrate with your existing project access validation
	// For now, returning true as placeholder
	return true
}
