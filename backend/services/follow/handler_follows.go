package follow

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vaanskii/vansify/db"
	"github.com/vaanskii/vansify/models"
	"github.com/vaanskii/vansify/notifications"
	"github.com/vaanskii/vansify/utils"
)

func FollowUser(c *gin.Context) {
    // Validate and extract claims
    claims, exists := c.Get("claims")
    if !exists || claims.(*utils.CustomClaims) == nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
        return
    }
    customClaims := claims.(*utils.CustomClaims)

    followerUsername := customClaims.Username
    followingUsername := c.Param("username")
    
    if followerUsername == followingUsername {
        c.JSON(http.StatusBadRequest, gin.H{"error": "You cannot follow yourself"})
        return
    }

    // Use a single query to get follower and following IDs
    var followerID, followingID int64
    err := db.DB.QueryRow(`
        SELECT f.id, u.id 
        FROM users f 
        JOIN users u ON u.username = ? 
        WHERE f.username = ?`, followingUsername, followerUsername).
        Scan(&followerID, &followingID)
    if err != nil {
        if err == sql.ErrNoRows {
            c.JSON(http.StatusBadRequest, gin.H{"error": "User does not exist"})
        } else {
            log.Printf("Error retrieving user IDs: %v\n", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving user IDs"})
        }
        return
    }

    // Check if the follow relationship already exists and create it if not
    tx, err := db.DB.Begin()
    if err != nil {
        log.Printf("Error starting transaction: %v\n", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error starting transaction"})
        return
    }

    var followExists bool
    err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM followers WHERE follower_id = ? AND following_id = ?)", followerID, followingID).Scan(&followExists)
    if err != nil {
        log.Printf("Error checking follow status: %v\n", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking follow status"})
        tx.Rollback()
        return
    }
    if followExists {
        c.JSON(http.StatusBadRequest, gin.H{"error": "You are already following this user"})
        tx.Rollback()
        return
    }

    // Create follow relationship and notification
    _, err = tx.Exec("INSERT INTO followers (follower_id, following_id) VALUES (?, ?)", followerID, followingID)
    if err != nil {
        log.Printf("Error creating follow relationship: %v\n", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error following user"})
        tx.Rollback()
        return
    }

    message := followerUsername + " started following you"
    _, err = tx.Exec("INSERT INTO notifications (user_id, type, message, follower_id) VALUES (?, ?, ?, ?)", followingID, models.FollowNotificationType, message, followerID)
    if err != nil {
        log.Printf("Error creating follow notification: %v\n", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating notification"})
        tx.Rollback()
        return
    }

    err = tx.Commit()
    if err != nil {
        log.Printf("Error committing transaction: %v\n", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error committing transaction"})
        return
    }

    // Get the unread notification count
    notificationCount, err := getUnreadNotificationCount(followingID)
    if err != nil {
        log.Printf("Error fetching unread notification count: %v\n", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching unread notification count"})
        return
    }

    notificationMessage := map[string]interface{}{
        "unread_notification_count": notificationCount,
        "sender":                    followerUsername,
        "receiver":                  followingUsername,
    }

    notificationJSON, err := json.Marshal(notificationMessage)
    if err != nil {
        log.Printf("Error marshalling notification count: %v\n", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error marshalling notification count"})
        return
    }

    notifications.GlobalNotificationHub.BroadcastNotification(followingUsername, notificationJSON)

    c.JSON(http.StatusOK, gin.H{"message": "Successfully followed user"})
}


func getUnreadNotificationCount(userID int64) (int, error) {
    var count int
    err := db.DB.QueryRow("SELECT COUNT(*) FROM notifications WHERE user_id = ? AND is_read = false", userID).Scan(&count)
    return count, err
}

func UnfollowUser(c *gin.Context) {
    claims, exists := c.Get("claims")
    if !exists || claims.(*utils.CustomClaims) == nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
        return
    }
    customClaims := claims.(*utils.CustomClaims)

    followerUsername := customClaims.Username
    followingUsername := c.Param("username")

    if followerUsername == followingUsername {
        c.JSON(http.StatusBadRequest, gin.H{"error": "You cannot unfollow yourself"})
        return
    }

    var followerID, followingID int64
    err := db.DB.QueryRow(`
        SELECT f.id, u.id 
        FROM users f 
        JOIN users u ON u.username = ? 
        WHERE f.username = ?`, followingUsername, followerUsername).Scan(&followerID, &followingID)
    if err != nil {
        if err == sql.ErrNoRows {
            c.JSON(http.StatusBadRequest, gin.H{"error": "User does not exist"})
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving user IDs"})
        }
        return
    }

    tx, err := db.DB.Begin()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error starting transaction"})
        return
    }

    var followExists bool
    err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM followers WHERE follower_id = ? AND following_id = ?)", followerID, followingID).Scan(&followExists)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking follow status"})
        tx.Rollback()
        return
    }

    if !followExists {
        c.JSON(http.StatusBadRequest, gin.H{"error": "You are not following this user"})
        tx.Rollback()
        return
    }

    _, err = tx.Exec("DELETE FROM followers WHERE follower_id = ? AND following_id = ?", followerID, followingID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unfollowing user"})
        tx.Rollback()
        return
    }

    err = tx.Commit()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error committing transaction"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Successfully unfollowed user"})
}

func CheckFollowStatus(c *gin.Context) {
    followerUsername := c.Param("follower")
    followingUsername := c.Param("following")

    var followerID, followingID int64
    err := db.DB.QueryRow(`
        SELECT f.id, u.id 
        FROM users f 
        JOIN users u ON u.username = ? 
        WHERE f.username = ?`, followingUsername, followerUsername).Scan(&followerID, &followingID)
    if err != nil {
        if err == sql.ErrNoRows {
            c.JSON(http.StatusBadRequest, gin.H{"error": "User does not exist"})
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving user IDs"})
        }
        return
    }

    var follows bool
    var followedBy bool

    // Check if the followerUsername follows followingUsername
    err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM followers WHERE follower_id = ? AND following_id = ?)", followerID, followingID).Scan(&follows)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking follow status"})
        return
    }

    // Check if the followingUsername follows followerUsername (mutual follow check)
    err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM followers WHERE follower_id = ? AND following_id = ?)", followingID, followerID).Scan(&followedBy)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking mutual follow status"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "is_following": follows,
        "is_followed_by": followedBy,
    })
}

func GetFollowers(c *gin.Context) {
    username := c.Param("username")
    
    var followers []gin.H
    query := `
        SELECT u.username, u.profile_picture 
        FROM followers f 
        JOIN users u ON f.follower_id = u.id 
        JOIN users u2 ON f.following_id = u2.id 
        WHERE u2.username = ?`
    
    rows, err := db.DB.Query(query, username)
    if err != nil {
        log.Printf("Error retrieving followers: %v\n", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving followers"})
        return
    }
    defer rows.Close()

    for rows.Next() {
        var username, profilePicture string
        if err := rows.Scan(&username, &profilePicture); err != nil {
            log.Printf("Error scanning follower: %v\n", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving followers"})
            return
        }
        followers = append(followers, gin.H{"username": username, "profile_picture": profilePicture})
    }

    if err := rows.Err(); err != nil {
        log.Printf("Error iterating through followers: %v\n", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving followers"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"followers": followers})
}

func GetFollowing(c *gin.Context) {
    username := c.Param("username")
    
    var followings []gin.H
    query := `
        SELECT u.username, u.profile_picture 
        FROM followers f 
        JOIN users u ON f.following_id = u.id 
        JOIN users u2 ON f.follower_id = u2.id 
        WHERE u2.username = ?`
    
    rows, err := db.DB.Query(query, username)
    if err != nil {
        log.Printf("Error retrieving followings: %v\n", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving followings"})
        return
    }
    defer rows.Close()

    for rows.Next() {
        var username, profilePicture string
        if err := rows.Scan(&username, &profilePicture); err != nil {
            log.Printf("Error scanning following user: %v\n", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving followings"})
            return
        }
        followings = append(followings, gin.H{"username": username, "profile_picture": profilePicture})
    }

    if err := rows.Err(); err != nil {
        log.Printf("Error iterating through followings: %v\n", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving followings"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"followings": followings})
}
