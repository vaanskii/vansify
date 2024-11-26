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

    // Call CheckUndeliveredMessagesForUser when a user connects
    // log.Printf("Checking undelivered messages for user %s", username)
    // CheckUndeliveredMessagesForUser(db.DB, username, conn)

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
// func IsUserConnected(username string) bool {
//     mu.Lock()
//     defer mu.Unlock()
//     _, connected := connectedUsers[username]
//     return connected
// }

// // Function to update message status and broadcast it
// func UpdateMessageStatus(db *sql.DB, messageID int, status string, messageType int) error {
//     _, err := db.Exec("UPDATE messages SET status = ? WHERE id = ?", status, messageID)
//     if err != nil {
//         log.Printf("Error updating message status: %v", err)
//         return err
//     }

//     // Broadcast status update
//     statusUpdate := map[string]interface{}{
//         "message_id": messageID,
//         "status":     status,
//     }
//     statusUpdateMessage, _ := json.Marshal(statusUpdate)
//     hub.GlobalChatHub.BroadcastMessage(nil, messageType, statusUpdateMessage)

//     return nil
// }

// func CheckUndeliveredMessagesForUser(db *sql.DB, username string, conn *websocket.Conn) {
//     log.Printf("Checking undelivered messages for user %s", username)
//     rows, err := db.Query("SELECT id, chat_id, username FROM messages WHERE status = 'sent' AND chat_id IN (SELECT chat_id FROM chats WHERE user1 = ? OR user2 = ?)", username, username)
//     if err != nil {
//         log.Printf("Error fetching messages: %v", err)
//         return
//     }
//     defer rows.Close()

//     for rows.Next() {
//         var messageID int
//         var chatID string
//         var senderUsername string
//         if err := rows.Scan(&messageID, &chatID, &senderUsername); err != nil {
//             log.Printf("Error scanning message ID: %v", err)
//             continue
//         }

//         log.Printf("Processing message ID %d in chat %s from %s to %s", messageID, chatID, senderUsername, username)

//         // Update message status to 'delivered'
//         err := UpdateMessageStatus(db, messageID, "delivered", websocket.TextMessage)
//         if err != nil {
//             log.Printf("Error updating message status: %v", err)
//         }

//         // Send notification to the frontend
//         reconnectionMessage := map[string]interface{}{
//             "type":       "USER_RECONNECTED",
//             "message_id": messageID,
//             "status":     "delivered",
//             "chat_id":    chatID,
//             "user":       username,
//         }
//         reconnectionMessageBytes, _ := json.Marshal(reconnectionMessage)
//         conn.WriteMessage(websocket.TextMessage, reconnectionMessageBytes)
//         log.Printf("Sent reconnection message for message ID %d to frontend", messageID)
//     }
// }

// func CheckAndDeliverMessages(db *sql.DB, recipientUsername string, conn *websocket.Conn) {
//     log.Printf("Checking undelivered messages for user %s", recipientUsername)
//     rows, err := db.Query("SELECT id, chat_id FROM messages WHERE status = 'sent' AND username != ?", recipientUsername)
//     if err != nil {
//         log.Printf("Error fetching messages: %v", err)
//         return
//     }
//     defer rows.Close()

//     for rows.Next() {
//         var messageID int
//         var chatID string
//         if err := rows.Scan(&messageID, &chatID); err != nil {
//             log.Printf("Error scanning message ID: %v", err)
//             continue
//         }

//         log.Printf("Processing message ID %d in chat %s", messageID, chatID)

//         // Fetch the other user in the chat
//         var user1, user2 string
//         err = db.QueryRow("SELECT user1, user2 FROM chats WHERE chat_id = ?", chatID).Scan(&user1, &user2)
//         if err != nil {
//             log.Printf("Error fetching chat users: %v", err)
//             continue
//         }

//         var senderUsername string
//         if recipientUsername == user1 {
//             senderUsername = user2
//         } else {
//             senderUsername = user1
//         }

//         log.Printf("Sender user: %s", senderUsername)
//         log.Printf("Recipient user: %s", recipientUsername)

//         // Check if the recipient (recipientUsername) is connected
//         if IsUserConnected(recipientUsername) {
//             log.Printf("Recipient %s is connected. Updating message ID %d to 'delivered'", recipientUsername, messageID)
//             err := UpdateMessageStatus(db, messageID, "delivered", websocket.TextMessage)
//             if err != nil {
//                 log.Printf("Error updating message status: %v", err)
//             }
//             // Send notification to the frontend
//             reconnectionMessage := map[string]interface{}{
//                 "type":       "USER_RECONNECTED",
//                 "message_id": messageID,
//                 "status":     "delivered",
//                 "chat_id":    chatID,
//                 "user":       recipientUsername,
//             }
//             reconnectionMessageBytes, _ := json.Marshal(reconnectionMessage)
//             conn.WriteMessage(websocket.TextMessage, reconnectionMessageBytes)
//             log.Printf("Sent reconnection message for message ID %d to frontend", messageID)
//         } else {
//             log.Printf("Recipient %s is not connected. Message ID %d remains 'sent'", recipientUsername, messageID)
//         }
//     }
// }
