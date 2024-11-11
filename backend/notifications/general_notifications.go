package notifications

import (
	"database/sql"
	"log"
	"net/http"
	"time"

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

    rows, err := db.DB.Query("SELECT id, user_id, type, message, is_read, created_at, follower_id FROM notifications WHERE user_id = ? ORDER BY created_at DESC", userID)
    if err != nil {
        log.Printf("Error fetching notifications: %v\n", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching notifications"})
        return
    }
    defer rows.Close()

    var notifications []map[string]interface{}
    for rows.Next() {
        var notification models.Notification
        var followerID sql.NullInt64
        var createdAt time.Time
        if err := rows.Scan(&notification.ID, &notification.UserID, &notification.Type, &notification.Message, &notification.IsRead, &createdAt, &followerID); err != nil {
            log.Printf("Error scanning notification: %v\n", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Error scanning notification"})
            return
        }
        formattedTime := createdAt.Format("2006-01-02 15:04:05")

        var profilePicture string
        if followerID.Valid {
            // Fetch profile picture for the user who followed (follower_id)
            err = db.DB.QueryRow("SELECT profile_picture FROM users WHERE id = ?", followerID.Int64).Scan(&profilePicture)
            if err != nil {
                log.Printf("Error fetching profile picture for user ID %d: %v\n", followerID.Int64, err)
                profilePicture = ""
            }
        } else {
            profilePicture = ""
        }

        notifications = append(notifications, map[string]interface{}{
            "id":              notification.ID,
            "user_id":         notification.UserID,
            "type":            notification.Type,
            "message":         notification.Message,
            "is_read":         notification.IsRead,
            "created_at":      formattedTime,
            "profile_picture": profilePicture,
        })
    }

    if err = rows.Err(); err != nil {
        log.Printf("Error after scanning rows: %v\n", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error after scanning rows"})
        return
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
