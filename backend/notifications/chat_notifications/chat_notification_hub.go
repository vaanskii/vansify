package chat_notifications

import (
	"sync"

	"github.com/gorilla/websocket"
)

type ChatNotificationHub struct {
    connections map[string]*websocket.Conn
    mu          sync.RWMutex
}

func NewChatNotificationHub() *ChatNotificationHub {
    return &ChatNotificationHub{
        connections: make(map[string]*websocket.Conn),
    }
}

func (h *ChatNotificationHub) AddConnection(conn *websocket.Conn, username string) {
    h.mu.Lock()
    defer h.mu.Unlock()
    h.connections[username] = conn
}

func (h *ChatNotificationHub) RemoveConnection(conn *websocket.Conn) {
    h.mu.Lock()
    defer h.mu.Unlock()
    for user, userConn := range h.connections {
        if userConn == conn {
            delete(h.connections, user)
            break
        }
    }
}

func (h *ChatNotificationHub) SendChatNotification(username string, message []byte) {
    h.mu.RLock()
    conn, exists := h.connections[username]
    h.mu.RUnlock()
    if exists {
        if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
            conn.Close()
            h.RemoveConnection(conn)
        }
    }
}

var ChatNotification = NewChatNotificationHub()
