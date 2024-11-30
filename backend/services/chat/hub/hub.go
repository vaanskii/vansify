package hub

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Hub struct {
    connections map[*websocket.Conn]string
    mu          sync.Mutex
}

// NewHub creates a new Hub instance
func NewHub() *Hub {
    return &Hub{
        connections: make(map[*websocket.Conn]string),
    }
}

// AddConnection adds a new connection to the Hub
func (h *Hub) AddConnection(conn *websocket.Conn, username string) {
    h.mu.Lock()
    h.connections[conn] = username 
    h.mu.Unlock()
}

// RemoveConnection removes a connection from the Hub
func (h *Hub) RemoveConnection(conn *websocket.Conn) {
    h.mu.Lock()
    delete(h.connections, conn)
    h.mu.Unlock()
}

// GetConnectionByUsername returns the connection associated with the given username
func (hub *Hub) GetConnectionByUsername(username string) *websocket.Conn {
    hub.mu.Lock()
    defer hub.mu.Unlock()
    for conn, user := range hub.connections {
        if user == username {
            return conn
        }
    }
    return nil
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
