package ws

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

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
	Resource  string      `json:"resource,omitempty"`  // Resource type: project, sprint, task, project_member
	Action    string      `json:"action,omitempty"`    // Action: INSERT, UPDATE, DELETE
	Data      interface{} `json:"data"`
	ProjectID string      `json:"projectId,omitempty"`
	UserID    string      `json:"userId,omitempty"`
}

// GlobalHub is the global WebSocket hub instance
var GlobalHub *Hub

// StartWebSocketHub initializes and starts the global WebSocket hub
func StartWebSocketHub() {
	GlobalHub = NewHub()
	go GlobalHub.Run()
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
	totalClients := len(h.clients)
	matchingClients := 0
	clientsToNotify := []string{} // Track user IDs for logging

	for client := range h.clients {
		if client.projectID == projectID {
			matchingClients++
			clientsToNotify = append(clientsToNotify, client.userID)
			select {
			case client.send <- msgBytes:
				// Message sent successfully
			default:
				close(client.send)
				delete(h.clients, client)
			}
		}
	}
	h.mutex.RUnlock()

	log.Printf("Broadcast: type=%s, project=%s, total_clients=%d, matching=%d, users=%v",
		messageType, projectID, totalClients, matchingClients, clientsToNotify)
}

// GetProjectConnections returns connection status for a specific project
func (h *Hub) GetProjectConnections(projectID string) (int, int, []string) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	totalClients := len(h.clients)
	matchingClients := 0
	userIDs := []string{}

	for client := range h.clients {
		if client.projectID == projectID {
			matchingClients++
			userIDs = append(userIDs, client.userID)
		}
	}

	return totalClients, matchingClients, userIDs
}

// BroadcastToAll sends a message to all connected clients
func (h *Hub) BroadcastToAll(messageType string, data interface{}) {
	msg := Message{
		Type: messageType,
		Data: data,
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	h.mutex.RLock()
	for client := range h.clients {
		select {
		case client.send <- msgBytes:
		default:
			close(client.send)
			delete(h.clients, client)
		}
	}
	h.mutex.RUnlock()
}

// ReadPump pumps messages from the WebSocket connection to the hub (exported)
func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		// Only close if WritePump hasn't already closed it
		// WritePump is responsible for sending close frame
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			// Check if this is a normal close
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNormalClosure) {
				log.Printf("WebSocket read error: %v", err)
			}
			// Break on any error (including normal close)
			break
		}

		// Handle the message
		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			continue
		}

		// Process the message based on type
		c.handleMessage(msg)
	}
}

// WritePump pumps messages from the hub to the WebSocket connection (exported)
func (c *Client) WritePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		// WritePump is responsible for sending close frame
		// ReadPump will handle cleanup after we close the send channel
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// Channel closed - send close frame and exit
				// This happens when hub unregisters the client
				c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				// Connection error - ReadPump will handle cleanup
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				// Write error - ReadPump will handle cleanup
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				// Ping failed - connection is dead, ReadPump will handle cleanup
				return
			}
		}
	}
}

// handleMessage processes incoming WebSocket messages
func (c *Client) handleMessage(msg Message) {
	switch msg.Type {
	case "join_project":
		c.projectID = msg.ProjectID
		log.Printf("Client joined project: %s", msg.ProjectID)
	case "leave_project":
		c.projectID = ""
		log.Printf("Client left project")
	case "init", "ping", "pong":
		// Protocol control messages - silently accept
		// These are used for connection health checks and initialization
		return
	default:
		log.Printf("Unknown message type: %s", msg.Type)
	}
}

// NewClient creates a new WebSocket client (exported helper function)
func NewClient(conn *websocket.Conn, userID string, projectID string, hub *Hub) *Client {
	return &Client{
		conn:      conn,
		userID:    userID,
		projectID: projectID,
		send:      make(chan []byte, 256),
		hub:       hub,
	}
}
