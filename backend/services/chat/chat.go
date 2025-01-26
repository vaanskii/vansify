package chat

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/vaanskii/vansify/db"
	"github.com/vaanskii/vansify/models"
	"github.com/vaanskii/vansify/notifications/chat_notifications"
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

    token = strings.TrimSpace(token)

    // Verify JWT token
    parsedToken, err := utils.VerifyJWT(token)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }

    claims, ok := parsedToken.Claims.(*utils.CustomClaims)
    if !ok || !parsedToken.Valid {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }

    senderUsername := claims.Username

    // Upgrade to WebSocket
    conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        return
    }
    defer conn.Close()

    cm := user.ChatManagerInstance
    cm.AddUserToChat(chatID, senderUsername)
    defer cm.RemoveUserFromChat(chatID, senderUsername)

    hub.AddConnection(conn, senderUsername)
    defer hub.RemoveConnection(conn)

    // Retrieve chat users
    var chat models.Chat
    err = db.DB.QueryRow("SELECT user1, user2 FROM chats WHERE chat_id = ?", chatID).Scan(&chat.User1, &chat.User2)
    if err != nil {
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
            break
        }

        var incomingMessage models.Message
        if err := json.Unmarshal(p, &incomingMessage); err != nil {
            continue
        }

        incomingMessage.ChatID = chatID
        incomingMessage.Username = senderUsername
        incomingMessage.Status = "sending"

        // Check for duplicate message
        var duplicateMessageID int
        err = db.DB.QueryRow("SELECT id FROM messages WHERE chat_id = ? AND message = ? AND username = ? AND created_at = ?", 
            incomingMessage.ChatID, incomingMessage.Message, incomingMessage.Username, incomingMessage.CreatedAt).Scan(&duplicateMessageID)
        if err == nil {
            continue
        }

        var profilePicture string
        err = db.DB.QueryRow("SELECT profile_picture FROM users WHERE username = ?", senderUsername).Scan(&profilePicture)
        if err != nil {
            profilePicture = ""
        }

        // Reset the deleted_for field for the chat and messages
        _, err = db.DB.Exec("UPDATE chats SET deleted_for = NULL WHERE chat_id = ?", chatID)
        if err != nil {
        }

        // Save message to database with initial status 'sending'
        result, execErr := db.DB.Exec("INSERT INTO messages (chat_id, message, username, file_url, status, created_at) VALUES (?, ?, ?, ?, ?, ?)",
            incomingMessage.ChatID, incomingMessage.Message, incomingMessage.Username, incomingMessage.FileURL, incomingMessage.Status, incomingMessage.CreatedAt)
        if execErr != nil {
            continue
        }

        messageID, err := result.LastInsertId()
        if err != nil {
            continue
        }
        incomingMessage.ID = int(messageID)

        // Send the real message ID back to the sender first
        idMessage := map[string]interface{}{
            "id":     incomingMessage.ID,
            "status": incomingMessage.Status,
            "type":   "MESSAGE_ID",
        }
        idMessageBytes, _ := json.Marshal(idMessage)
        conn.WriteMessage(messageType, idMessageBytes)

        // Update message status to 'sent' after saving to DB
        incomingMessage.Status = "sent"
        _, err = db.DB.Exec("UPDATE messages SET status = ? WHERE id = ?", incomingMessage.Status, incomingMessage.ID)
        if err != nil {
        }

        // Check if the recipient is in the chat before updating the status
        if cm.IsUserInChat(chatID, recipientUsername) {
            incomingMessage.Status = "read"
            _, err = db.DB.Exec("UPDATE messages SET status = ? WHERE id = ?", incomingMessage.Status, incomingMessage.ID)
            if err == nil {
                // Send real-time update to the recipient
                statusUpdateMessage := map[string]interface{}{
                    "type":    "STATUS_UPDATE",
                    "chat_id": chatID,
                    "status":  "delivered",
                    "message_ids": []int{incomingMessage.ID},
                    "username": senderUsername,
                }
                broadcastMessage, _ := json.Marshal(statusUpdateMessage)
                hub.BroadcastMessage(nil, websocket.TextMessage, broadcastMessage)
                MarkChatNotificationsAsRead(c)
                log.Printf("Notifications marked as read for user: %s in chat: %s", recipientUsername, chatID)
            } else {
                log.Printf("Error marking message as read: %v", err)
            }
        } else {
            go UpdateStatusWhenUserBecomesActive(incomingMessage.ChatID, recipientUsername, incomingMessage.Username)
            log.Print("Status updated to delivered")

            // Send the delivered status update for messages
            statusUpdateMessage := map[string]interface{}{
                "type":    "STATUS_UPDATE",
                "chat_id": chatID,
                "status":  "delivered",
                "message_ids": []int{incomingMessage.ID},
                "username": senderUsername,
            }
            broadcastMessage, _ := json.Marshal(statusUpdateMessage)
            hub.BroadcastMessage(nil, websocket.TextMessage, broadcastMessage)
        }

        // Prepare full message to send to clients
        fullMessage := struct {
            models.Message
            ProfilePicture  string    `json:"profile_picture"`
            Receiver       string    `json:"receiver"`
            CreatedAt      string    `json:"created_at"`
        }{
            Message:        incomingMessage,
            ProfilePicture: profilePicture,
            Receiver:       recipientUsername,
            CreatedAt:      time.Now().Format(time.RFC3339),
        }

        // Marshal the full message
        broadcastMessage, _ := json.Marshal(fullMessage)

        // Send the message to the recipient only
        recipientConn := hub.GetConnectionByUsername(recipientUsername)
        if recipientConn != nil {
            recipientConn.WriteMessage(messageType, broadcastMessage)
        }

        // Send the message back to the sender
        conn.WriteMessage(messageType, broadcastMessage)

        // Fetch the last message for notifications
        var lastMessage string
        err = db.DB.QueryRow("SELECT message FROM messages WHERE chat_id = ? ORDER BY created_at DESC LIMIT 1", chatID).Scan(&lastMessage)
        if err != nil {
        }

        var recipientID int
        err = db.DB.QueryRow("SELECT id FROM users WHERE username = ?", recipientUsername).Scan(&recipientID)
        if err != nil {
        } else {
            if !cm.IsUserInChat(chatID, recipientUsername) { // Only send notifications if the recipient is not in the chat
                chat_notifications.NotifyNewMessage(int64(recipientID), incomingMessage)
                chatUnreadCount, err := chat_notifications.GetUnreadChatMessagesCount(int64(recipientID), chatID)
                if err == nil {
                    totalUnreadCount, err := chat_notifications.GetTotalUnreadMessageCount(int64(recipientID))
                    if err == nil {
                        chatNotificationMessage := map[string]interface{}{
                            "user_id":            recipientID,
                            "chat_id":            chatID,
                            "unread_count":       chatUnreadCount,
                            "total_unread_count": totalUnreadCount,
                            "message":            incomingMessage.Message,
                            "recipient":          recipientUsername,
                            "user":               senderUsername,
                            "profile_picture":     profilePicture,
                            "sender":             senderUsername,
                            "last_message_time":  time.Now().Format(time.RFC3339),
                            "last_message":       lastMessage,
                        }
                        chatNotificationJSON, _ := json.Marshal(chatNotificationMessage)
                        fmt.Print(chatNotificationJSON)
                        chat_notifications.ChatNotification.SendChatNotification(recipientUsername, chatNotificationJSON)
                    }
                }
            }
        }
    }
}

func UpdateStatusWhenUserBecomesActive(chatID string, recipientUsername string, senderUsername string) {
    var isActive bool
    var lastActive sql.NullTime
    var becomingInactive bool
    err := db.DB.QueryRow("SELECT active, last_active, becoming_inactive FROM users WHERE username = ?", recipientUsername).Scan(&isActive, &lastActive, &becomingInactive)
    if err != nil {
        return
    }

    if isActive && !lastActive.Valid {
        rows, err := db.DB.Query("SELECT id FROM messages WHERE chat_id = ? AND username = ? AND status = 'sent'", chatID, senderUsername)
        if err != nil {
            return
        }
        defer rows.Close()

        messageIDs := []int{}
        for rows.Next() {
            var messageID int
            if err := rows.Scan(&messageID); err != nil {
                continue
            }
            messageIDs = append(messageIDs, messageID)
        }

        _, err = db.DB.Exec("UPDATE messages SET status = 'delivered' WHERE chat_id = ? AND username = ? AND status = 'sent'", chatID, senderUsername)
        if err != nil {
        } else {
            statusUpdateMessage := map[string]interface{}{
                "type":       "STATUS_UPDATE",
                "chat_id":    chatID,
                "username":   senderUsername,
                "status":     "delivered",
                "message_ids": messageIDs,
            }
            broadcastMessage, _ := json.Marshal(statusUpdateMessage)
            hub.BroadcastMessage(nil, websocket.TextMessage, broadcastMessage)
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
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving user ID"})
        return
    }

    // Update message statuses to "read" only if the user is the recipient
    _, err = db.DB.Exec("UPDATE messages SET status = 'read' WHERE chat_id = ? AND username = ?", chatID, recipientUsername)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating message statuses to read"})
        return
    }

    // Delete notifications for the chat
    _, err = db.DB.Exec("DELETE FROM chat_notifications WHERE user_id = ? AND chat_id = ?", userID, chatID)
    if err != nil {
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

    c.JSON(http.StatusOK, gin.H{"message": "Messages marked as read and notifications deleted"})
    log.Printf("Notifications Deleted")
}

func GetChatHistory(c *gin.Context) {
    chatID := c.Param("chatID")

    // Get the authenticated user from the claims
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

    // Fetch the profile picture for the authenticated user
    var profilePicture string
    err := db.DB.QueryRow("SELECT profile_picture FROM users WHERE username = ?", username).Scan(&profilePicture)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching user profile picture"})
        return
    }

    // Get limit and offset from query parameters
    limit, err := strconv.Atoi(c.DefaultQuery("limit", "20"))
    if err != nil || limit <= 0 {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit specified"})
        return
    }

    offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
    if err != nil || offset < 0 {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offset specified"})
        return
    }

    // Fetch messages for the chat with pagination, including messages marked as deleted for the user
    rows, err := db.DB.Query("SELECT id, chat_id, message, username, file_url, created_at, status FROM messages WHERE chat_id = ? AND (deleted_for IS NULL OR deleted_for NOT LIKE ?) ORDER BY created_at ASC LIMIT ? OFFSET ?", chatID, "%"+username+"%", limit, offset)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching chat history"})
        return
    }
    defer rows.Close()

    var messages []map[string]interface{}
    for rows.Next() {
        var message models.Message
        var createdAt time.Time
        if err := rows.Scan(&message.ID, &message.ChatID, &message.Message, &message.Username, &message.FileURL, &createdAt, &message.Status); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Error scanning message"})
            return
        }
        formattedTime := createdAt.Format(time.RFC3339)
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
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }

    customClaims, ok := claims.(*utils.CustomClaims)
    if !ok {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
        return
    }

    username := customClaims.Username
    var chat models.Chat

    // Ensure the user is part of the chat
    err := db.DB.QueryRow("SELECT user1, user2 FROM chats WHERE chat_id = ?", chatID).Scan(&chat.User1, &chat.User2)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error querying chat"})
        return
    }

    if username != chat.User1 && username != chat.User2 {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }

    // Permanently delete all messages in the chat
    _, err = db.DB.Exec("DELETE FROM messages WHERE chat_id = ?", chatID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting messages"})
        return
    }

    _, err = db.DB.Exec("DELETE FROM chat_notifications WHERE chat_id = ? AND is_read = FALSE", chatID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting messages"})
        return
    }

    // Mark the chat as deleted for both users
    _, err = db.DB.Exec(`
        UPDATE chats 
        SET deleted_for = 
            CASE 
                WHEN deleted_for IS NULL OR deleted_for = '' THEN ?
                ELSE CONCAT(deleted_for, ',', ?)
            END 
        WHERE chat_id = ?`, 
        chat.User1+","+chat.User2, chat.User1+","+chat.User2, chatID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error marking chat as deleted"})
        return
    }

    // Broadcast the deletion of the chat to connected clients
    deleteChat := map[string]interface{}{
        "type":    "CHAT_DELETED",
        "chat_id": chatID,
    }
    broadcastMessage, _ := json.Marshal(deleteChat)
    hub.BroadcastMessage(nil, websocket.TextMessage, broadcastMessage)

    c.JSON(http.StatusOK, gin.H{"message": "Chat and messages deleted successfully for both users"})
}

func DeleteUserMessages(c *gin.Context) {
    chatID := c.Param("chatID")

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
    var chatUserCount int
    err := db.DB.QueryRow("SELECT COUNT(*) FROM chats WHERE chat_id = ? AND (user1 = ? OR user2 = ?)", chatID, username, username).Scan(&chatUserCount)
    if err != nil || chatUserCount == 0 {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }

    // Mark all messages as deleted for the user
    _, err = db.DB.Exec(`
        UPDATE messages 
        SET deleted_for = 
            CASE 
                WHEN deleted_for IS NULL OR deleted_for = '' THEN ?
                ELSE CONCAT_WS(',', deleted_for, ?)
            END 
        WHERE chat_id = ? AND (deleted_for IS NULL OR deleted_for NOT LIKE ?)`, 
        username, username, chatID, "%"+username+"%")
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error marking messages as deleted"})
        return
    }

    // Mark the chat as deleted for the user
    _, err = db.DB.Exec(`
        UPDATE chats 
        SET deleted_for = 
            CASE 
                WHEN deleted_for IS NULL OR deleted_for = '' THEN ?
                ELSE CONCAT(deleted_for, ',', ?)
            END 
        WHERE chat_id = ? AND (deleted_for IS NULL OR deleted_for NOT LIKE ?)`, 
        username, username, chatID, "%"+username+"%")
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error marking chat as deleted"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Messages and chat deleted successfully for user"})
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

    // Fetch the last message and last message time
    var lastMessage, lastMessageTime string
    err = db.DB.QueryRow("SELECT message, created_at FROM messages WHERE chat_id = ? ORDER BY created_at DESC LIMIT 1", message.ChatID).Scan(&lastMessage, &lastMessageTime)
    if err != nil {
        lastMessage = ""
        lastMessageTime = ""
    }

    // Identify the recipient of the chat
    var recipientUsername string
    var chat models.Chat
    err = db.DB.QueryRow("SELECT user1, user2 FROM chats WHERE chat_id = ?", message.ChatID).Scan(&chat.User1, &chat.User2)
    if err == nil {
        if username == chat.User1 {
            recipientUsername = chat.User2
        } else {
            recipientUsername = chat.User1
        }
    }

    // Fetch recipient userID
    var recipientUserID int64
    err = db.DB.QueryRow("SELECT id FROM users WHERE username = ?", recipientUsername).Scan(&recipientUserID)
    if err == nil {
        // Delete chat notification for the recipient if exists
        _, err = db.DB.Exec("DELETE FROM chat_notifications WHERE user_id = ? AND chat_id = ?", recipientUserID, message.ChatID)
        if err == nil {
            // Call GetTotalUnreadMessageCount to update the unread message count for the recipient
            totalUnreadCount, err := chat_notifications.GetTotalUnreadMessageCount(recipientUserID)
            if err == nil {
                // Broadcast deletion to all connected clients
                deleteMessage := map[string]interface{}{
                    "type":              "MESSAGE_DELETED",
                    "message_id":        messageID,
                    "chat_id":           message.ChatID,
                    "last_message":      lastMessage,
                    "last_message_time": lastMessageTime,
                    "status":            message.Status,
                    "total_unread_count": totalUnreadCount,
                }
                broadcastMessage, _ := json.Marshal(deleteMessage)
                hub.BroadcastMessage(nil, websocket.TextMessage, broadcastMessage)

                c.JSON(http.StatusOK, gin.H{"message": "Message deleted successfully"})
                return
            }
        }
    }

    c.JSON(http.StatusOK, gin.H{"message": "Message deleted successfully"})
}
