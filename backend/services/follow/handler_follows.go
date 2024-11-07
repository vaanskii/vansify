package follow

import (
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
    // Retrieve the claims from the context set by the middleware
    claims, exists := c.Get("claims")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "No claims found"})
        return
    }

    // Assuming claims is of type *utils.CustomClaims
    customClaims, ok := claims.(*utils.CustomClaims)
    if !ok {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
        return
    }

    // Extract username from claims
    followerUsername := customClaims.Username
    followingUsername := c.Param("username")

    // Get follower ID from follower username
    var followerID int64
    err := db.DB.QueryRow("SELECT id FROM users WHERE username = ?", followerUsername).Scan(&followerID)
    if err != nil {
        log.Printf("Error retrieving follower ID: %v\n", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving follower ID"})
        return
    }

    // Get following ID from following username
    var followingID int64
    err = db.DB.QueryRow("SELECT id FROM users WHERE username = ?", followingUsername).Scan(&followingID)
    if err != nil {
        log.Printf("Error retrieving following ID: %v\n", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": "User does not exist"})
        return
    }

    if followerID == followingID {
        c.JSON(http.StatusBadRequest, gin.H{"error": "You cannot follow yourself"})
        return
    }

    // Check if the follow relationship already exists
    var followExists bool
    err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM followers WHERE follower_id = ? AND following_id = ?)", followerID, followingID).Scan(&followExists)
    if err != nil {
        log.Printf("Error checking follow status: %v\n", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking follow status"})
        return
    }

    if followExists {
        c.JSON(http.StatusBadRequest, gin.H{"error": "You are already following this user"})
        return
    }

    // Create a follow relationship
    _, err = db.DB.Exec("INSERT INTO followers (follower_id, following_id) VALUES (?, ?)", followerID, followingID)
    if err != nil {
        log.Printf("Error creating follow relationship: %v\n", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error following user"})
        return
    }

    // Create follow notification
    message := followerUsername + " started following you"
    _, err = db.DB.Exec("INSERT INTO notifications (user_id, type, message) VALUES (?, ?, ?)", followingID, models.FollowNotificationType, message)
    if err != nil {
        log.Printf("Error creating follow notification: %v\n", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating notification"})
        return
    }

    // Broadcast notification count
    notificationCount, err := getUnreadNotificationCount(followingID)
    if err != nil {
        log.Printf("Error fetching unread notification count: %v\n", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching unread notification count"})
        return
    }

    notificationMessage := map[string]int{
        "unread_notification_count": notificationCount,
    }

    notificationJSON, err := json.Marshal(notificationMessage)
    if err != nil {
        log.Printf("Error marshalling notification count: %v\n", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error marshalling notification count"})
        return
    }

    notifications.GlobalNotificationHub.BroadcastNotification(notificationJSON)

    c.JSON(http.StatusOK, gin.H{"message": "Successfully followed user"})
}

func getUnreadNotificationCount(userID int64) (int, error) {
    var count int
    err := db.DB.QueryRow("SELECT COUNT(*) FROM notifications WHERE user_id = ? AND is_read = false", userID).Scan(&count)
    return count, err
}

func UnfollowUser(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No claims found"})
		return
	}

	customClaims, ok := claims.(*utils.CustomClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
		return
	}


	followerUsername := customClaims.Username // The username of the logged-in user
	followingUsername := c.Param("username") // The username of the user to unfollow

	// Get follower ID from follower username
	var followerID int64
	err := db.DB.QueryRow("SELECT id FROM users WHERE username = ?", followerUsername).Scan(&followerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving follower ID"})
		return
	}

	// Get following ID from following username
	var followingID int64
	err = db.DB.QueryRow("SELECT id FROM users WHERE username = ?", followingUsername).Scan(&followingID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User does not exist"})
		return
	}

	// Check if the follow relationship exists
	var followExists bool
	err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM followers WHERE follower_id = ? AND following_id = ?)", followerID, followingID).Scan(&followExists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking follow status"})
		return
	}

	if !followExists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You are not following this user"})
		return
	}

	// Remove the Follow relationship
	_, err = db.DB.Exec("DELETE FROM followers WHERE follower_id = ? AND following_id = ?", followerID, followingID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unfollowing user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully unfollowed user"})
}

func CheckFollowStatus(c *gin.Context) {
    followerUsername := c.Param("follower")
    followingUsername := c.Param("following")
    var followerID, followingID int64

    if err := db.DB.QueryRow("SELECT id FROM users WHERE username = ?", followerUsername).Scan(&followerID); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Follower not found"})
        return
    }
    if err := db.DB.QueryRow("SELECT id FROM users WHERE username = ?", followingUsername).Scan(&followingID); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Following user not found"})
        return
    }

    var followExists bool
    if err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM followers WHERE follower_id = ? AND following_id = ?)", followerID, followingID).Scan(&followExists); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking follow status"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"is_following": followExists})
}
func GetFollowers(c *gin.Context) {
    username := c.Param("username")
    var userID int64
    err := db.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
    if err != nil {
        log.Printf("Error retrieving user ID: %v\n", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving user ID"})
        return
    }
    rows, err := db.DB.Query("SELECT u.username, u.profile_picture FROM followers f JOIN users u ON f.follower_id = u.id WHERE f.following_id = ?", userID)
    if err != nil {
        log.Printf("Error retrieving followers: %v\n", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving followers"})
        return
    }
    defer rows.Close()

    var followers []gin.H
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
    var userID int64
    err := db.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
    if err != nil {
        log.Printf("Error retrieving user ID: %v\n", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving user ID"})
        return
    }
    rows, err := db.DB.Query("SELECT u.username, u.profile_picture FROM followers f JOIN users u ON f.following_id = u.id WHERE f.follower_id = ?", userID)
    if err != nil {
        log.Printf("Error retrieving followings: %v\n", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving followings"})
        return
    }
    defer rows.Close()

    var followings []gin.H
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
