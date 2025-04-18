package chat_notifications

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vaanskii/vansify/db"
	"github.com/vaanskii/vansify/models"
	"github.com/vaanskii/vansify/utils"
)

func NotifyNewMessage(userID int64, message models.Message) {
    log.Printf("Inserting notification for userID: %d, message: %s", userID, message.Message)
    _, err := db.DB.Exec("INSERT INTO chat_notifications (user_id, message, chat_id, is_read) VALUES (?, ?, ?, false)", userID, message.Message, message.ChatID)
    if err != nil {
        log.Printf("Error saving notification: %v", err)
        return
    }

    // Get the unread message count for this user in the specific chat
    var chatUnreadCount int
    err = db.DB.QueryRow("SELECT COUNT(*) FROM chat_notifications WHERE user_id = ? AND chat_id = ? AND is_read = false", userID, message.ChatID).Scan(&chatUnreadCount)
    if err != nil {
        log.Printf("Error getting unread message count for chat: %v", err)
        return
    }
    // Fetch the last message for notifications
    var lastMessage string
    err = db.DB.QueryRow("SELECT message FROM messages WHERE chat_id = ? ORDER BY created_at DESC LIMIT 1", message.ChatID).Scan(&lastMessage)
    if err != nil {
        log.Printf("Error getting last message: %v", err)
        return
    }

    // Broadcast the notification to all connected clients
    notificationMessage, _ := json.Marshal(map[string]interface{}{
        "user_id":           userID,
        "chat_id":           message.ChatID,
        "message":           message.Message,
        "user":              message.Username,
        "sender":            message.Username,
        "last_message_time": time.Now().UTC().Format(time.RFC3339),
        "last_message":      lastMessage,
    })
    ChatNotification.SendChatNotification(message.Username, notificationMessage)
}


func GetTotalUnreadMessageCount(userID int64) (int, error) {
    var count int
    err := db.DB.QueryRow("SELECT COUNT(*) FROM chat_notifications WHERE user_id = ? AND is_read = false", userID).Scan(&count)
    return count, err
}


func GetUnreadChatMessagesCount(userID int64, chatID string) (int, error) {
    var count int
    err := db.DB.QueryRow("SELECT COUNT(*) FROM chat_notifications WHERE user_id = ? AND chat_id = ? AND is_read = false", userID, chatID).Scan(&count)
    return count, err
}

func GetUnreadChatNotifications(c *gin.Context) {
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
    var userID int64
    err := db.DB.QueryRow("SELECT id FROM users WHERE username = ?", customClaims.Username).Scan(&userID)
    if err != nil {
        log.Printf("Error retrieving user ID: %v\n", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving user ID"})
        return
    }
    count, err := GetTotalUnreadMessageCount(userID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching notifications"})
        return
    }
    c.JSON(http.StatusOK, gin.H{"unread_count": count})
}
