package notifications

import (
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/vaanskii/vansify/utils"
)

var notificationUpgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin:     func(r *http.Request) bool { return true },
}

// A map to track connected users
var connectedUsers = make(map[string]*websocket.Conn)
var mu sync.Mutex

func NotificationWsHandler(c *gin.Context) {
    claims, exists := c.Get("claims")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }
    customClaims, ok := claims.(*utils.CustomClaims)
    if !ok {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }
    username := customClaims.Username

    conn, err := notificationUpgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        log.Println("WebSocket Upgrade error:", err)
        return
    }
    log.Printf("User %s connected for notifications", username)

    mu.Lock()
    connectedUsers[username] = conn
    mu.Unlock()
    defer func() {
        mu.Lock()
        delete(connectedUsers, username)
        mu.Unlock()
        log.Printf("User %s disconnected from notifications", username)
        conn.Close()
    }()

    GlobalNotificationHub.AddConnection(conn)
    defer GlobalNotificationHub.RemoveConnection(conn)

    for {
        messageType, p, err := conn.ReadMessage()
        if err != nil {
            log.Println("WebSocket ReadMessage error:", err)
            return
        }
        log.Printf("Received: %s", p)
        if err := conn.WriteMessage(messageType, p); err != nil {
            log.Println("WebSocket WriteMessage error:", err)
            return
        }
    }
}
