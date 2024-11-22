package notifications

import (
	"database/sql"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/vaanskii/vansify/db"
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

    // Call CheckAndDeliverMessages when a user connects
    CheckAndDeliverMessages(db.DB, username)

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


// Function to check if a user is connected
func IsUserConnected(username string) bool {
    mu.Lock()
    defer mu.Unlock()
    _, connected := connectedUsers[username]
    return connected
}


func UpdateMessageStatus(db *sql.DB, messageID int, status string) error {
    _, err := db.Exec("UPDATE messages SET status = ? WHERE id = ?", status, messageID)
    if err != nil {
        log.Printf("Error updating message status: %v", err)
        return err
    }
    return nil
}


func CheckAndDeliverMessages(db *sql.DB, recipientUsername string) {
    rows, err := db.Query("SELECT id, chat_id FROM messages WHERE status = 'sent' AND username != ?", recipientUsername)
    if err != nil {
        log.Printf("Error fetching messages: %v", err)
        return
    }
    defer rows.Close()

    for rows.Next() {
        var messageID int
        var chatID string
        if err := rows.Scan(&messageID, &chatID); err != nil {
            log.Printf("Error scanning message ID: %v", err)
            continue
        }

        // Fetch the other user in the chat
        var user1, user2 string
        err := db.QueryRow("SELECT user1, user2 FROM chats WHERE chat_id = ?", chatID).Scan(&user1, &user2)
        if err != nil {
            log.Printf("Error fetching chat users: %v", err)
            continue
        }

        var otherUser string
        if recipientUsername == user1 {
            otherUser = user2
        } else {
            otherUser = user1
        }

        // Check if the recipient (otherUser) is connected
        if IsUserConnected(otherUser) {
            if err := UpdateMessageStatus(db, messageID, "delivered"); err != nil {
                log.Printf("Error updating message status to delivered: %v", err)
            }
        }
    }
}
