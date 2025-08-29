package ws

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Add proper origin checking in production
	},
}

// HandleWS handles regular WebSocket connections
func HandleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	client := &Client{
		conn: conn,
		send: make(chan []byte, 256),
		hub:  GlobalHub,
	}

	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}

// HandleWSAuth handles authenticated WebSocket connections
func HandleWSAuth(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement authentication logic
	// For now, just handle as regular connection
	HandleWS(w, r)
}
