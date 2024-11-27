package chat

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/vaanskii/vansify/db"
	"github.com/vaanskii/vansify/models"
	"github.com/vaanskii/vansify/notifications"
	chatHub "github.com/vaanskii/vansify/services/chat/hub"
	"github.com/vaanskii/vansify/services/user"
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
	hub = chatHub.NewHub()
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

func ChatWsHandler(c *gin.Context) {
    chatID := c.Param("chatID")
    token := c.Query("token")
    log.Printf("Attempting to upgrade connection for chatID: %s with token: %s", chatID, token)

    token = strings.TrimSpace(token)
    log.Printf("Token after trimming: %s", token)

    // Verify JWT token
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

    senderUsername := claims.Username
    log.Printf("Token valid for user: %s", senderUsername)

    // Upgrade to WebSocket
    conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        log.Println("WebSocket Upgrade error:", err)
        return
    }
    log.Println("WebSocket connection established for chatID:", chatID)
    defer conn.Close()

    cm := user.ChatManagerInstance

    cm.AddUserToChat(chatID, senderUsername)
    defer cm.RemoveUserFromChat(chatID, senderUsername)

    hub.AddConnection(conn)
    defer hub.RemoveConnection(conn)

    // Retrieve chat users
    var chat models.Chat
    err = db.DB.QueryRow("SELECT user1, user2 FROM chats WHERE chat_id = ?", chatID).Scan(&chat.User1, &chat.User2)
    if err != nil {
        log.Println("Error querying chat:", err)
        return
    }

    // Determine recipient username
    var recipientUsername string
    if senderUsername == chat.User1 {
        recipientUsername = chat.User2
    } else {
        recipientUsername = chat.User1
    }

    for {
        messageType, p, err := conn.ReadMessage()
        if err != nil {
            log.Println("WebSocket ReadMessage error:", err)
            break
        }

        log.Printf("Received message: %s", string(p))

        var incomingMessage models.Message
        if err := json.Unmarshal(p, &incomingMessage); err != nil {
            log.Printf("Error decoding incoming message: %v", err)
            continue
        }

        incomingMessage.ChatID = chatID
        incomingMessage.Username = senderUsername
        incomingMessage.Status = "sent"
        log.Printf("Message initial status set to %s", incomingMessage.Status)

        // Fetch the profile picture URL for the sender
        var profilePicture string
        err = db.DB.QueryRow("SELECT profile_picture FROM users WHERE username = ?", senderUsername).Scan(&profilePicture)
        if err != nil {
            log.Println("Error querying profile picture:", err)
            profilePicture = ""
        }

        // Save message to database
        result, execErr := db.DB.Exec("INSERT INTO messages (chat_id, message, username, file_url, status) VALUES (?, ?, ?, ?, ?)",
            incomingMessage.ChatID, incomingMessage.Message, incomingMessage.Username, incomingMessage.FileURL, incomingMessage.Status)
        if execErr != nil {
            log.Printf("DB Exec error: %v", execErr)
            continue
        }

        messageID, err := result.LastInsertId()
        if err != nil {
            log.Printf("Error retrieving last insert ID: %v", err)
            continue
        }
        incomingMessage.ID = int(messageID)
        log.Printf("Message ID: %d saved with status: %s", incomingMessage.ID, incomingMessage.Status)

        // Check if the recipient is in the chat before updating the status
        if cm.IsUserInChat(chatID, recipientUsername) {
            incomingMessage.Status = "read"
            _, err = db.DB.Exec("UPDATE messages SET status = ? WHERE id = ?", incomingMessage.Status, incomingMessage.ID)
            if err != nil {
                log.Printf("Error updating message status to read for message ID %d: %v", incomingMessage.ID, err)
            } else {
                log.Printf("Message ID %d status updated to 'read' because recipient %s is in chat", incomingMessage.ID, recipientUsername)
            }
        } else {
            log.Printf("Recipient %s is not in chat. Message status remains as '%s'", recipientUsername, incomingMessage.Status)
            go UpdateStatusWhenUserBecomesActive(incomingMessage.ChatID, recipientUsername, incomingMessage.Username)
        }

        // Send the real message ID back to the sender first
        idMessage := map[string]interface{}{
            "id":     incomingMessage.ID,
            "status": incomingMessage.Status,
            "type":   "MESSAGE_ID",
        }
        idMessageBytes, _ := json.Marshal(idMessage)
        conn.WriteMessage(messageType, idMessageBytes)

        // Prepare full message to send to clients
        fullMessage := struct {
            models.Message
            ProfilePicture string `json:"profile_picture"`
            Receiver       string `json:"receiver"`
        }{
            Message:        incomingMessage,
            ProfilePicture: profilePicture,
            Receiver:       recipientUsername,
        }

        // Send the message back to the sender
        broadcastMessage, _ := json.Marshal(fullMessage)
        conn.WriteMessage(messageType, broadcastMessage)
        log.Printf("Sent message ID %d back to sender", incomingMessage.ID)

        // Broadcast the message to other connected clients
        hub.BroadcastMessage(conn, messageType, broadcastMessage)
        log.Printf("Broadcasted message ID %d to other clients", incomingMessage.ID)

        // Fetch the last message for notifications
        var lastMessage string
        err = db.DB.QueryRow("SELECT message FROM messages WHERE chat_id = ? ORDER BY created_at DESC LIMIT 1", chatID).Scan(&lastMessage)
        if err != nil {
            log.Println("Error querying last message:", err)
        }

        log.Printf("Last message: %s", lastMessage)

        var recipientID int
        err = db.DB.QueryRow("SELECT id FROM users WHERE username = ?", recipientUsername).Scan(&recipientID)
        if err != nil {
            log.Println("Error querying recipient ID:", err)
        } else {
            notifications.NotifyNewMessage(int64(recipientID), incomingMessage)
            chatUnreadCount, err := notifications.GetUnreadChatMessagesCount(int64(recipientID), chatID)
            if err != nil {
                log.Printf("Error getting unread message count for chat: %v", err)
            } else {
                totalUnreadCount, err := notifications.GetTotalUnreadMessageCount(int64(recipientID))
                if err != nil {
                    log.Printf("Error getting total unread message count: %v", err)
                } else {
                    chatNotificationMessage, _ := json.Marshal(map[string]interface{}{
                        "user_id":            recipientID,
                        "chat_id":            chatID,
                        "unread_count":       chatUnreadCount,
                        "total_unread_count": totalUnreadCount,
                        "message":            incomingMessage.Message,
                        "user":               senderUsername,
                        "profile_picture":    profilePicture,
                        "sender":             senderUsername,
                        "last_message_time":  time.Now().Format(time.RFC3339),
                        "last_message":       lastMessage,
                    })
                    notifications.GlobalNotificationHub.BroadcastNotification(chatNotificationMessage)
                }
            }
        }
    }
}


func UpdateStatusWhenUserBecomesActive(chatID string, recipientUsername string, senderUsername string) {
    log.Printf("Updating status for chat %s, recipient %s, sender %s", chatID, recipientUsername, senderUsername)

    var isActive bool
    var lastActive sql.NullTime
    var becomingInactive bool
    err := db.DB.QueryRow("SELECT active, last_active, becoming_inactive FROM users WHERE username = ?", recipientUsername).Scan(&isActive, &lastActive, &becomingInactive)
    if err != nil {
        log.Printf("Error querying active status for user uswb %s: %v", recipientUsername, err)
        return
    }

    if isActive && !lastActive.Valid {
        rows, err := db.DB.Query("SELECT id FROM messages WHERE chat_id = ? AND username = ? AND status = 'sent'", chatID, senderUsername)
        if err != nil {
            log.Printf("Error querying messages for chat %s and sender %s: %v", chatID, senderUsername, err)
            return
        }
        defer rows.Close()

        messageIDs := []int{}
        for rows.Next() {
            var messageID int
            if err := rows.Scan(&messageID); err != nil {
                log.Printf("Error scanning message ID: %v", err)
                continue
            }
            messageIDs = append(messageIDs, messageID)
        }

        result, err := db.DB.Exec("UPDATE messages SET status = 'delivered' WHERE chat_id = ? AND username = ? AND status = 'sent'", chatID, senderUsername)
        if err != nil {
            log.Printf("Error updating message statuses for chat %s and user %s: %v", chatID, senderUsername, err)
        } else {
            rowsAffected, _ := result.RowsAffected()
            log.Printf("Updated %d message(s) to 'delivered' in chat %s for sender %s", rowsAffected, chatID, senderUsername)

            statusUpdateMessage := map[string]interface{}{
                "type":       "STATUS_UPDATE",
                "chat_id":    chatID,
                "username":   senderUsername,
                "status":     "delivered",
                "message_ids": messageIDs,
            }
            broadcastMessage, _ := json.Marshal(statusUpdateMessage)
            hub.BroadcastMessage(nil, websocket.TextMessage, broadcastMessage)
            log.Printf("Broadcasted status update for chat %s and user %s: %s", chatID, senderUsername, broadcastMessage)
        }
    }
}


func MarkChatNotificationsAsRead(c *gin.Context) {
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
    chatID := c.Param("chatID")

    // Ensure user is part of the chat
    var chat models.Chat
    err := db.DB.QueryRow("SELECT user1, user2 FROM chats WHERE chat_id = ?", chatID).Scan(&chat.User1, &chat.User2)
    if err != nil {
        log.Printf("Error querying chat: %v\n", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error querying chat"})
        return
    }

    // Determine recipient username
    var recipientUsername string
    if username == chat.User1 {
        recipientUsername = chat.User2
    } else {
        recipientUsername = chat.User1
    }

    var userID int64
    err = db.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
    if err != nil {
        log.Printf("Error retrieving user ID: %v\n", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving user ID"})
        return
    }

    // Update message statuses to "read" only if the user is the recipient
    _, err = db.DB.Exec("UPDATE messages SET status = 'read' WHERE chat_id = ? AND username = ?", chatID, recipientUsername)
    if err != nil {
        log.Printf("Error updating message statuses to read: %v\n", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating message statuses to read"})
        return
    }

    // Delete notifications for the chat
    _, err = db.DB.Exec("DELETE FROM chat_notifications WHERE user_id = ? AND chat_id = ?", userID, chatID)
    if err != nil {
        log.Printf("Error deleting notifications for chat: %v\n", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting notifications for chat"})
        return
    }

    // Broadcast the status update to all connected clients
    statusUpdateMessage := map[string]interface{}{
        "type":    "STATUS_UPDATE_READ",
        "chat_id": chatID,
        "status":  "read",
        "username": username,
    }
    broadcastMessage, _ := json.Marshal(statusUpdateMessage)
    hub.BroadcastMessage(nil, websocket.TextMessage, broadcastMessage)
    log.Printf("Broadcasted status update for chat %s: %s", chatID, broadcastMessage)

    c.JSON(http.StatusOK, gin.H{"message": "Messages marked as read and notifications deleted"})
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

    // Get limit and offset from query parameters
    limit, err := strconv.Atoi(c.DefaultQuery("limit", "20"))
    if err != nil || limit <= 0 {
        log.Println("Invalid limit specified")
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit specified"})
        return
    }

    offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
    if err != nil || offset < 0 {
        log.Println("Invalid offset specified")
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offset specified"})
        return
    }

    // Fetch messages for the chat with pagination
    rows, err := db.DB.Query("SELECT id, chat_id, message, username, file_url, created_at, status FROM messages WHERE chat_id = ? LIMIT ? OFFSET ?", chatID, limit, offset)
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
        if err := rows.Scan(&message.ID, &message.ChatID, &message.Message, &message.Username, &message.FileURL, &createdAt, &message.Status); err != nil {
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
            "profile_picture":  profilePicture,
            "file_url":         message.FileURL,
            "status":          message.Status, 
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

// Function to delete the chat from the application and Google Drive
func DeleteChat(c *gin.Context) {
    chatID := c.Param("chatID")

    claims, exists := c.Get("claims")
    if (!exists) {
        log.Println("No claims found")
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }

    customClaims, ok := claims.(*utils.CustomClaims)
    if (!ok) {
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

    // Broadcast the chat deletion to connected clients
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
    err := db.DB.QueryRow("SELECT id, chat_id, username, status FROM messages WHERE id = ?", messageID).Scan(&message.ID, &message.ChatID, &message.Username, &message.Status)
    if err != nil {
        log.Println("Query error: Incorrect message ID:", err)
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Incorrect message ID"})
        return
    }
    if message.Username != username {
        log.Println("Unauthorized: User does not have permission to delete this message")
        c.JSON(http.StatusUnauthorized, gin.H{"error": "You do not have permission to delete this message"})
        return
    }

    _, err = db.DB.Exec("DELETE FROM messages WHERE id = ?", messageID)
    if err != nil {
        log.Println("Database error: Error deleting message:", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting message"})
        return
    }

    log.Println("Message deleted successfully, Message ID:", messageID)

    // Fetch the last message and last message time
    var lastMessage, lastMessageTime string
    err = db.DB.QueryRow("SELECT message, created_at FROM messages WHERE chat_id = ? ORDER BY created_at DESC LIMIT 1", message.ChatID).Scan(&lastMessage, &lastMessageTime)
    if err != nil {
        log.Println("Query error: Error fetching last message:", err)
        lastMessage = ""
        lastMessageTime = ""
    }
    log.Printf("Last message after deletion: '%s' at '%s'", lastMessage, lastMessageTime)

    // Identify the recipient of the chat
    var recipientUsername string
    var chat models.Chat
    err = db.DB.QueryRow("SELECT user1, user2 FROM chats WHERE chat_id = ?", message.ChatID).Scan(&chat.User1, &chat.User2)
    if err != nil {
        log.Println("Error querying chat users:", err)
    } else {
        if username == chat.User1 {
            recipientUsername = chat.User2
        } else {
            recipientUsername = chat.User1
        }
    }

    // Fetch recipient userID
    var recipientUserID int64
    err = db.DB.QueryRow("SELECT id FROM users WHERE username = ?", recipientUsername).Scan(&recipientUserID)
    if err != nil {
        log.Println("Error fetching recipient user ID:", err)
    }

    // Delete chat notification for the recipient if exists
    _, err = db.DB.Exec("DELETE FROM chat_notifications WHERE user_id = ? AND chat_id = ?", recipientUserID, message.ChatID)
    if err != nil {
        log.Println("Error deleting chat notification for recipient:", err)
    } else {
        log.Println("Chat notification for recipient deleted successfully")
    }

    // Call GetTotalUnreadMessageCount to update the unread message count for the recipient
    totalUnreadCount, err := notifications.GetTotalUnreadMessageCount(recipientUserID)
    if err != nil {
        log.Println("Error getting total unread message count for recipient:", err)
    } else {
        log.Printf("Total unread message count for recipient: %d", totalUnreadCount)
    }

    // Broadcast deletion to all connected clients
    deleteMessage := map[string]interface{}{
        "type":              "MESSAGE_DELETED",
        "message_id":        messageID,
        "chat_id":           message.ChatID,
        "last_message":      lastMessage,
        "last_message_time": lastMessageTime,
        "status":            message.Status,
    }
    broadcastMessage, _ := json.Marshal(deleteMessage)
    log.Printf("Broadcasting delete message: %v", deleteMessage)
    hub.BroadcastMessage(nil, websocket.TextMessage, broadcastMessage)

    c.JSON(http.StatusOK, gin.H{"message": "Message deleted successfully"})
}

