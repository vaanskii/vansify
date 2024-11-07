package notifications

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vaanskii/vansify/db"
	"github.com/vaanskii/vansify/models"
	"github.com/vaanskii/vansify/utils"
)

func GetNotifications(c *gin.Context) {
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

    rows, err := db.DB.Query("SELECT id, user_id, type, message, is_read, created_at FROM notifications WHERE user_id = ?", userID)
    if err != nil {
        log.Printf("Error fetching notifications: %v\n", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching notifications"})
        return
    }
    defer rows.Close()

    var notifications []models.Notification
    for rows.Next() {
        var notification models.Notification
        if err := rows.Scan(&notification.ID, &notification.UserID, &notification.Type, &notification.Message, &notification.IsRead, &notification.CreatedAt); err != nil {
            log.Printf("Error scanning notification: %v\n", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Error scanning notification"})
            return
        }
        notifications = append(notifications, notification)
    }

    c.JSON(http.StatusOK, gin.H{"notifications": notifications})
}

func GetUnreadNotificationCount(c *gin.Context) {
    claims, exists := c.Get("claims")
    if (!exists) {
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

    var count int
    err = db.DB.QueryRow("SELECT COUNT(*) FROM notifications WHERE user_id = ? AND is_read = false", userID).Scan(&count)
    if err != nil {
        log.Printf("Error fetching unread notification count: %v\n", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching unread notification count"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"unread_count": count})
}

func MarkNotificationAsRead(c *gin.Context) {
    claims, exists := c.Get("claims")
    if (!exists) {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }
    customClaims, ok := claims.(*utils.CustomClaims)
    if (!ok) {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }
    notificationID := c.Param("notificationID")
    _, err := db.DB.Exec("UPDATE notifications SET is_read = true WHERE id = ? AND user_id = (SELECT id FROM users WHERE username = ?)", notificationID, customClaims.Username)
    if err != nil {
        log.Printf("Error marking notification as read: %v\n", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error marking notification as read"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Notification marked as read"})
}

func DeleteNotification(c *gin.Context) {
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

    notificationID := c.Param("notificationID")
    _, err := db.DB.Exec("DELETE FROM notifications WHERE id = ? AND user_id = (SELECT id FROM users WHERE username = ?)", notificationID, customClaims.Username)
    if err != nil {
        log.Printf("Error deleting notification: %v\n", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting notification"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Notification deleted"})
}
