package chat

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/vaanskii/vansify/db"
	"github.com/vaanskii/vansify/models"
	notifications "github.com/vaanskii/vansify/notifications"
	"github.com/vaanskii/vansify/utils"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize: 1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	hub = NewHub()
)

func generateChatID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
        return "", err
    }
    return hex.EncodeToString(bytes), nil
}


// Creating chat if not exists
func CreateChat(c *gin.Context) {
    // Extract claims from context
    claims, exists := c.Get("claims")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }
    customClaims := claims.(*utils.CustomClaims)
    user1 := customClaims.Username

    // Bind JSON to chat object
    var chat models.Chat
    if err := c.ShouldBindJSON(&chat); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    chat.User1 = user1  // Automatically set user1 from the authenticated user

    // Check if user2 exists
    var user2Exists bool
    err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = ?)", chat.User2).Scan(&user2Exists)
    if err != nil || !user2Exists {
        c.JSON(http.StatusBadRequest, gin.H{"error": "User2 does not exist"})
        return
    }

    // Check if chat already exists
    var existingChat string
    err = db.DB.QueryRow("SELECT chat_id FROM chats WHERE (user1 = ? AND user2 = ?) OR (user1 = ? AND user2 = ?)",
        chat.User1, chat.User2, chat.User2, chat.User1).Scan(&existingChat)
    if err == nil {
        c.JSON(http.StatusOK, gin.H{"chat_id": existingChat})
        return
    } else if err != sql.ErrNoRows {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking existing chat"})
        return
    }

    // If chat does not exist, create a new one
    chatID, err := generateChatID()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating chat ID"})
        return
    }
    chat.ChatID = chatID
    _, execErr := db.DB.Exec("INSERT INTO chats (chat_id, user1, user2) VALUES (?, ?, ?)", chat.ChatID, chat.User1, chat.User2)
    if execErr != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error saving chat to database"})
        return
    }

    // Do not notify the other user until a message is sent
    c.JSON(http.StatusOK, gin.H{"chat_id": chat.ChatID})
}

// WsHandler which is handling ws connections
func ChatWsHandler(c *gin.Context) {
    chatID := c.Param("chatID")
    token := c.Query("token")
    log.Printf("Attempting to upgrade connection for chatID: %s with token: %s", chatID, token)

    token = strings.TrimSpace(token)
    log.Printf("Token after trimming: %s", token)

    parsedToken, err := utils.VerifyJWT(token)
    if err != nil {
        log.Println("Invalid token:", err)
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }

    claims, ok := parsedToken.Claims.(*utils.CustomClaims)
    if !ok || !parsedToken.Valid {
        log.Println("Invalid claims or token not valid")
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }

    log.Printf("Token valid for user: %s", claims.Username)

    conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        log.Println("WebSocket Upgrade error:", err)
        return
    }
    log.Println("WebSocket connection established for chatID:", chatID)
    defer conn.Close()

    hub.AddConnection(conn)
    defer hub.RemoveConnection(conn)

    var chat models.Chat
    err = db.DB.QueryRow("SELECT user1, user2 FROM chats WHERE chat_id = ?", chatID).Scan(&chat.User1, &chat.User2)
    if err != nil {
        log.Println("Error querying chat:", err)
        return
    }

    for {
        messageType, p, err := conn.ReadMessage()
        if err != nil {
            log.Println("WebSocket ReadMessage error:", err)
            break
        }

        var incomingMessage models.Message
        if err := json.Unmarshal(p, &incomingMessage); err != nil {
            log.Println("Error decoding incoming message:", err)
            continue
        }

        incomingMessage.ChatID = chatID
        incomingMessage.Username = claims.Username

        // Fetch the profile picture URL for the user
        var profilePicture string
        err = db.DB.QueryRow("SELECT profile_picture FROM users WHERE username = ?", incomingMessage.Username).Scan(&profilePicture)
        if err != nil {
            log.Println("Error querying profile picture:", err)
            profilePicture = ""
        }

        // Save message to database
        result, execErr := db.DB.Exec("INSERT INTO messages (chat_id, message, username) VALUES (?, ?, ?)", incomingMessage.ChatID, incomingMessage.Message, incomingMessage.Username)
        if execErr != nil {
            log.Println("DB Exec error:", execErr)
            continue
        }

        messageID, err := result.LastInsertId()
        if err != nil {
            log.Println("Error retrieving last insert ID:", err)
            continue
        }
        incomingMessage.ID = int(messageID)
        fmt.Printf("Message ID: %d\n", incomingMessage.ID)

        // Create a message structure that includes the profile picture and real ID
        fullMessage := struct {
            models.Message
            ProfilePicture string `json:"profile_picture"`
        }{
            Message:        incomingMessage,
            ProfilePicture: profilePicture,
        }

        // Send the message back to the sender only
        broadcastMessage, _ := json.Marshal(fullMessage)
        conn.WriteMessage(messageType, broadcastMessage)

        // Broadcast the message to all connected clients except the sender
        hub.BroadcastMessage(conn, messageType, broadcastMessage)

        // Determine recipient and notify
        var recipientUsername string
        if claims.Username == chat.User1 {
            recipientUsername = chat.User2
        } else {
            recipientUsername = chat.User1
        }

        var recipientID int
        err = db.DB.QueryRow("SELECT id FROM users WHERE username = ?", recipientUsername).Scan(&recipientID)
        if err != nil {
            log.Println("Error querying recipient ID:", err)
        } else {
            notifications.NotifyNewMessage(int64(recipientID), incomingMessage)
            log.Printf("Notification sent to user ID: %d for message: %s", recipientID, incomingMessage.Message)

            // Get unread count for this specific chat
            chatUnreadCount, err := notifications.GetUnreadChatMessagesCount(int64(recipientID), chatID)
            if err != nil {
                log.Printf("Error getting unread message count for chat: %v", err)
            } else {
                totalUnreadCount, err := notifications.GetTotalUnreadMessageCount(int64(recipientID))
                if err != nil { log.Printf("Error getting total unread message count: %v", err) 
            } else {
                    // Broadcast the chat-specific unread count notification
                    chatNotificationMessage, _ := json.Marshal(map[string]interface{}{
                        "user_id":           recipientID,
                        "chat_id":           chatID,
                        "unread_count":      chatUnreadCount,
                        "total_unread_count": totalUnreadCount,
                        "message":           incomingMessage.Message,
                        "user":              claims.Username,
                        "profile_picture":   profilePicture,
                        "sender":            claims.Username,
                        "last_message_time": time.Now().Format(time.RFC3339),
                    })
                    notifications.GlobalNotificationHub.BroadcastNotification(chatNotificationMessage)
                }
            }
        }
    }
}


func GetChatHistory(c *gin.Context) {
    chatID := c.Param("chatID")
    chattingUser := c.Query("user")

    if chattingUser == "" {
        log.Println("No user specified in the query parameters")
        c.JSON(http.StatusBadRequest, gin.H{"error": "No user specified in the query parameters"})
        return
    }

    log.Printf("Fetching profile picture for user: %s", chattingUser)

    // Fetch the profile picture for the chatting user
    var profilePicture string
    err := db.DB.QueryRow("SELECT profile_picture FROM users WHERE username = ?", chattingUser).Scan(&profilePicture)
    if err != nil {
        log.Println("Error fetching user profile picture:", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching user profile picture"})
        return
    }
    log.Printf("Profile Picture for %s: %s", chattingUser, profilePicture)

    // Fetch messages for the chat
    rows, err := db.DB.Query("SELECT id, chat_id, message, username, created_at FROM messages WHERE chat_id = ?", chatID)
    if err != nil {
        log.Println("Error fetching chat history:", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching chat history"})
        return
    }
    defer rows.Close()

    var messages []map[string]interface{}
    for rows.Next() {
        var message models.Message
        var createdAt time.Time
        if err := rows.Scan(&message.ID, &message.ChatID, &message.Message, &message.Username, &createdAt); err != nil {
            log.Println("Error scanning message:", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Error scanning message"})
            return
        }
        formattedTime := createdAt.Format("2006-01-02 15:04:05")
        messages = append(messages, map[string]interface{}{
            "id":              message.ID,
            "chat_id":         message.ChatID,
            "message":         message.Message,
            "username":        message.Username,
            "created_at":      formattedTime,
            "profile_picture": profilePicture,
        })
    }
    log.Println("Fetched Messages:", messages)
    c.JSON(http.StatusOK, messages)
}


func CheckChatExists(c *gin.Context) {
	user1 := c.Param("user1")
	user2 := c.Param("user2")
  
	var chatID string
	err := db.DB.QueryRow("SELECT chat_id FROM chats WHERE (user1 = ? AND user2 = ?) OR (user1 = ? AND user2 = ?)", user1, user2, user2, user1).Scan(&chatID)
	if err != nil {
	  if err == sql.ErrNoRows {
		c.JSON(http.StatusOK, gin.H{"chat_id": ""})
	  } else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking chat existence"})
	  }
	  return
	}
  
	c.JSON(http.StatusOK, gin.H{"chat_id": chatID})
}

func DeleteChat(c *gin.Context) {
    chatID := c.Param("chatID")

    claims, exists := c.Get("claims")
    if !exists {
        log.Println("No claims found")
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }

    customClaims, ok := claims.(*utils.CustomClaims)
    if !ok {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
        return
    }

    username := customClaims.Username
    var chatUserCount int

    // Ensure the user is part of the chat to delete it
    err := db.DB.QueryRow("SELECT COUNT(*) FROM chats WHERE chat_id = ? AND (user1 = ? OR user2 = ?)", chatID, username, username).Scan(&chatUserCount)
    if err != nil || chatUserCount == 0 {
        log.Printf("Chat not found or user is not part of the chat. Error: %v", err)
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }

    // Delete the chat from the database
    _, err = db.DB.Exec("DELETE FROM chats WHERE chat_id = ?", chatID)
    if err != nil {
        log.Printf("Error deleting chat: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting chat"})
        return
    }

    deleteChat := map[string]interface{}{
        "type":    "CHAT_DELETED",
        "chat_id": chatID,
    }
    broadcastMessage, _ := json.Marshal(deleteChat)
    log.Printf("Broadcasting delete chat: %s", broadcastMessage)
    hub.BroadcastMessage(nil, websocket.TextMessage, broadcastMessage)

    c.JSON(http.StatusOK, gin.H{"message": "Chat deleted successfully"})
}

func DeleteMessage(c *gin.Context) {
    claims, exists := c.Get("claims")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }
    customClaims, ok := claims.(*utils.CustomClaims)
    if !ok {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
        return
    }
    username := customClaims.Username
    messageID := c.Param("messageID")

    var message models.Message
    err := db.DB.QueryRow("SELECT id, username FROM messages WHERE id = ?", messageID).Scan(&message.ID, &message.Username)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Incorrect message ID"})
        return
    }
    if message.Username != username {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "You do not have permission to delete this message"})
        return
    }

    _, err = db.DB.Exec("DELETE FROM messages WHERE id = ?", messageID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting message"})
        return
    }

    // Broadcast deletion to all connected clients
    deleteMessage := map[string]interface{}{
        "type":      "MESSAGE_DELETED",
        "message_id": messageID,
    }
    broadcastMessage, _ := json.Marshal(deleteMessage)
    log.Printf("Broadcasting delete message: %s", broadcastMessage)
    hub.BroadcastMessage(nil, websocket.TextMessage, broadcastMessage) 

    c.JSON(http.StatusOK, gin.H{"message": "Message deleted successfully"})
}
