package chat_notifications

import (
	"sync"

	"github.com/gorilla/websocket"
)

type ChatNotificationHub struct {
	connections map[string]*websocket.Conn
	mu 			sync.Mutex
}

func NewChatNotificationHub() *ChatNotificationHub {
    return &ChatNotificationHub{connections: make(map[string]*websocket.Conn)}
}

func (h *ChatNotificationHub) AddConnection(conn *websocket.Conn, username string) {
    h.mu.Lock()
    h.connections[username] = conn
    h.mu.Unlock()
}

func (h *ChatNotificationHub) RemoveConnection(conn *websocket.Conn) {
    h.mu.Lock()
    for user, userConn := range h.connections {
        if userConn == conn {
            delete(h.connections, user)
            break
        }
    }
    h.mu.Unlock()
}

func (h *ChatNotificationHub) SendChatNotification(username string, message []byte) {
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

var ChatNotification = NewChatNotificationHub()
