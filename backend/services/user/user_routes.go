package user

import (
	"log"
	"net/http"

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
        log.Println("Unauthorized: No claims found")
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
        log.Printf("Error retrieving user ID: %v\n", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving user ID"})
        return
    }

    rows, err := db.DB.Query(`
        SELECT 
            c.chat_id, 
            c.user1, 
            c.user2, 
            COALESCE(MAX(m.created_at), '') AS last_message_time,
            (SELECT message FROM messages WHERE chat_id = c.chat_id ORDER BY created_at DESC LIMIT 1) AS last_message
        FROM chats c
        LEFT JOIN messages m ON c.chat_id = m.chat_id
        WHERE (c.user1 = ? OR c.user2 = ?) AND c.chat_id IN (SELECT DISTINCT chat_id FROM messages)
        GROUP BY c.chat_id, c.user1, c.user2`, username, username)
    if err != nil {
        log.Printf("Error fetching user chats: %v\n", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching user chats"})
        return
    }
    defer rows.Close()

    var chats []map[string]interface{}
    for rows.Next() {
        var chatID, user1, user2, lastMessageTime, lastMessage string
        if err := rows.Scan(&chatID, &user1, &user2, &lastMessageTime, &lastMessage); err != nil {
            log.Printf("Error scanning chat: %v\n", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Error scanning chat"})
            return
        }

        otherUser := user1
        if user1 == username {
            otherUser = user2
        }

        // Fetch the profile picture of the other user
        var profilePicture string
        err = db.DB.QueryRow("SELECT profile_picture FROM users WHERE username = ?", otherUser).Scan(&profilePicture)
        if err != nil {
            log.Printf("Error querying profile picture for user %s: %v\n", otherUser, err)
            profilePicture = "" // Set to empty or a default value if needed
        }

        log.Printf("Other User: %s, Profile Picture: %s", otherUser, profilePicture)

        // Get unread message count for each chat
        var unreadCount int
        err = db.DB.QueryRow("SELECT COUNT(*) FROM chat_notifications WHERE user_id = ? AND chat_id = ? AND is_read = false", userID, chatID).Scan(&unreadCount)
        if err != nil {
            log.Printf("Error fetching unread count: %v\n", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching unread count"})
            return
        }

        chats = append(chats, map[string]interface{}{
            "chat_id":            chatID,
            "user":               otherUser,
            "unread_count":       unreadCount,
            "last_message_time":  lastMessageTime,
            "profile_picture":     profilePicture,
            "last_message":       lastMessage, 
        })
        
    }

    c.JSON(http.StatusOK, gin.H{"chats": chats})
}

