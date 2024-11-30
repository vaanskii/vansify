package notifications

import (
	"sync"

	"github.com/gorilla/websocket"
)

type NotificationHub struct {
    connections map[string]*websocket.Conn
    mu          sync.Mutex
}

// NewNotificationHub creates a new NotificationHub instance
func NewNotificationHub() *NotificationHub {
    return &NotificationHub{connections: make(map[string]*websocket.Conn)}
}

// AddConnection adds a new connection to the NotificationHub
func (h *NotificationHub) AddConnection(conn *websocket.Conn, username string) {
    h.mu.Lock()
    h.connections[username] = conn
    h.mu.Unlock()
}

// RemoveConnection removes a connection from the NotificationHub
func (h *NotificationHub) RemoveConnection(conn *websocket.Conn) {
    h.mu.Lock()
    for user, userConn := range h.connections {
        if userConn == conn {
            delete(h.connections, user)
            break
        }
    }
    h.mu.Unlock()
}

// BroadcastNotification sends a notification to all connected clients
func (h *NotificationHub) BroadcastNotification(username string, message []byte) {
    h.mu.Lock()
    conn, exists := h.connections[username]
    h.mu.Unlock()
    if exists {
        if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
            conn.Close()
            h.RemoveConnection(conn)
        }
    }
}

var GlobalNotificationHub = NewNotificationHub()
