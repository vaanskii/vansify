package chat

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Hub struct {
	connections map[*websocket.Conn]bool
	mu          sync.Mutex
}

// NewHub creates a new Hub instance
func NewHub() *Hub {
	return &Hub{
		connections: make(map[*websocket.Conn]bool),
	}
}

// AddConnection adds a new connection to the Hub
func (h *Hub) AddConnection(conn *websocket.Conn) {
	h.mu.Lock()
	h.connections[conn] = true
	h.mu.Unlock()
}

// RemoveConnection removes a connection from the Hub
func (h *Hub) RemoveConnection(conn *websocket.Conn) {
	h.mu.Lock()
	delete(h.connections, conn)
	h.mu.Unlock()
}

// BroadcastMessage sends a message to all connected clients except the sender
func (h *Hub) BroadcastMessage(sender *websocket.Conn, messageType int, message []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for conn := range h.connections {
		if conn != sender {
			if err := conn.WriteMessage(messageType, message); err != nil {
				conn.Close()
				delete(h.connections, conn)
			}
		}
	}
}