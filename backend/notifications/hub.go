package notifications

import (
	"sync"

	"github.com/gorilla/websocket"
)

type NotificationHub struct {
    connections map[*websocket.Conn]bool
    mu          sync.Mutex
}

// NewNotificationHub creates a new NotificationHub instance
func NewNotificationHub() *NotificationHub {
    return &NotificationHub{connections: make(map[*websocket.Conn]bool)}
}

// AddConnection adds a new connection to the NotificationHub
func (h *NotificationHub) AddConnection(conn *websocket.Conn) {
    h.mu.Lock()
    h.connections[conn] = true
    h.mu.Unlock()
}

// RemoveConnection removes a connection from the NotificationHub
func (h *NotificationHub) RemoveConnection(conn *websocket.Conn) {
    h.mu.Lock()
    delete(h.connections, conn)
    h.mu.Unlock()
}

// BroadcastNotification sends a notification to all connected clients
func (h *NotificationHub) BroadcastNotification(message []byte) {
    h.mu.Lock()
    defer h.mu.Unlock()
    for conn := range h.connections {
        if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
            conn.Close()
            delete(h.connections, conn)
        }
    }
}

var GlobalNotificationHub = NewNotificationHub()
