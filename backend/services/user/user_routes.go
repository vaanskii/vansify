package user

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/vaanskii/vansify/db"
	"github.com/vaanskii/vansify/models"
	"github.com/vaanskii/vansify/utils"
)

// UserProfile holds the user profile details to be returned in the response
type UserProfile struct {
	ID              int64         `json:"id"`
	Username        string        `json:"username"`
	FollowersCount  int           `json:"followers_count"`
	FollowingsCount int           `json:"followings_count"`
	Followers       []Follower    `json:"followers"`
	Followings      []Following   `json:"followings"`
	ProfilePicture   string        `json:"profile_picture"`
	Gender 			string   	  `json:"gender"`
    OauthUser       bool          `json:"oauth_user"`
}

type Follower struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
}

type Following struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
}

// GetUserByUsername function geting user by username
func GetUserByUsername(c *gin.Context) {
    username := c.Param("username")
    var user models.User

    // Fetch user details by username
    err := db.DB.QueryRow("SELECT id, username, email, profile_picture, gender, verified, created_at, oauth_user FROM users WHERE username = ?", username).Scan(&user.ID, &user.Username, &user.Email, &user.ProfilePicture, &user.Gender, &user.Verified, &user.CreatedAt, &user.OauthUser)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
        return
    }

    // Prepare the user profile response
    profile := UserProfile{
        ID:             user.ID,
        Username:       user.Username,
        ProfilePicture:  user.ProfilePicture,
        Gender:         user.Gender,
        OauthUser:      user.OauthUser,
    }

    // Fetch follower count
    err = db.DB.QueryRow("SELECT COUNT(*) FROM followers WHERE following_id = ?", user.ID).Scan(&profile.FollowersCount)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching followers count"})
        return
    }

    // Fetch following count
    err = db.DB.QueryRow("SELECT COUNT(*) FROM followers WHERE follower_id = ?", user.ID).Scan(&profile.FollowingsCount)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching followings count"})
        return
    }

    // Fetch followers
    rows, err := db.DB.Query("SELECT u.id, u.username FROM followers f JOIN users u ON f.follower_id = u.id WHERE f.following_id = ?", user.ID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching followers"})
        return
    }
    defer rows.Close()
    for rows.Next() {
        var follower Follower
        if err := rows.Scan(&follower.ID, &follower.Username); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Error scanning follower"})
            return
        }
        profile.Followers = append(profile.Followers, follower)
    }

    // Fetch followings
    rows, err = db.DB.Query("SELECT u.id, u.username FROM followers f JOIN users u ON f.following_id = u.id WHERE f.follower_id = ?", user.ID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching followings"})
        return
    }
    defer rows.Close()
    for rows.Next() {
        var following Following
        if err := rows.Scan(&following.ID, &following.Username); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Error scanning following"})
            return
        }
        profile.Followings = append(profile.Followings, following)
    }

    c.JSON(http.StatusOK, profile)
}

func GetUserChats(c *gin.Context) {
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

    var userID int64
    err := db.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving user ID"})
        return
    }

    rows, err := db.DB.Query(`
        SELECT 
            c.chat_id, 
            c.user1, 
            c.user2, 
            COALESCE(c.deleted_for, '') AS deleted_for, 
            COALESCE(MAX(m.created_at), '') AS last_message_time,
            COALESCE((SELECT message FROM messages WHERE chat_id = c.chat_id AND (deleted_for IS NULL OR deleted_for NOT LIKE ?) ORDER BY created_at DESC LIMIT 1), '') AS last_message
        FROM chats c
        LEFT JOIN messages m ON c.chat_id = m.chat_id
        WHERE (c.user1 = ? OR c.user2 = ?) 
        GROUP BY c.chat_id, c.user1, c.user2, c.deleted_for
        HAVING last_message IS NOT NULL`, "%"+username+"%", username, username)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching user chats"})
        return
    }
    defer rows.Close()

    var chats []map[string]interface{}
    for rows.Next() {
        var chatID, user1, user2, deletedFor string
        var lastMessageTime, lastMessage sql.NullString

        if err := rows.Scan(&chatID, &user1, &user2, &deletedFor, &lastMessageTime, &lastMessage); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Error scanning chat"})
            return
        }

        if deletedFor != "" && contains(deletedFor, username) {
            continue
        }

        otherUser := user1
        if user1 == username {
            otherUser = user2
        }

        // Fetch the profile picture of the other user
        var profilePicture string
        err = db.DB.QueryRow("SELECT profile_picture FROM users WHERE username = ?", otherUser).Scan(&profilePicture)
        if err != nil {
            profilePicture = ""
        }

        // Get unread message count for each chat
        var unreadCount int
        err = db.DB.QueryRow("SELECT COUNT(*) FROM chat_notifications WHERE user_id = ? AND chat_id = ? AND is_read = false", userID, chatID).Scan(&unreadCount)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching unread count"})
            return
        }

        chats = append(chats, map[string]interface{}{
            "chat_id":            chatID,
            "user":               otherUser,
            "unread_count":       unreadCount,
            "last_message_time":  lastMessageTime.String,
            "profile_picture":    profilePicture,
            "last_message":       lastMessage.String, 
        })
    }

    c.JSON(http.StatusOK, gin.H{"chats": chats})
}

// Helper function to check if the username is in the deleted_for list
func contains(deletedFor, username string) bool {
    for _, u := range strings.Split(deletedFor, ",") {
        if u == username {
            return true
        }
    }
    return false
}

func GetActiveUsersHandler(c *gin.Context) {
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

    authenticatedUsername := customClaims.Username

    var activeUsers []struct {
        Username       string `json:"username"`
        ProfilePicture string `json:"profile_picture"`
    }

    rows, err := db.DB.Query("SELECT username, profile_picture FROM users WHERE active = true")
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    defer rows.Close()

    for rows.Next() {
        var user struct {
            Username       string `json:"username"`
            ProfilePicture string `json:"profile_picture"`
        }
        if err := rows.Scan(&user.Username, &user.ProfilePicture); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        if user.Username != authenticatedUsername {
            activeUsers = append(activeUsers, user)
        }
    }
    if err := rows.Err(); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"active_users": activeUsers})
}
