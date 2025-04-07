package user

import (
	"database/sql"
	"log"
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
    err := db.DB.QueryRow("SELECT id, username, email, profile_picture, verified, created_at, oauth_user FROM users WHERE username = ?", username).Scan(&user.ID, &user.Username, &user.Email, &user.ProfilePicture, &user.Verified, &user.CreatedAt, &user.OauthUser)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
        return
    }

    // Prepare the user profile response
    profile := UserProfile{
        ID:             user.ID,
        Username:       user.Username,
        ProfilePicture: user.ProfilePicture,
        OauthUser:      user.OauthUser,
    }

    // Fetch follower count and following count in one query using subqueries
    err = db.DB.QueryRow(`
        SELECT 
            (SELECT COUNT(*) FROM followers WHERE following_id = ?) AS followers_count,
            (SELECT COUNT(*) FROM followers WHERE follower_id = ?) AS followings_count
        `, user.ID, user.ID).Scan(&profile.FollowersCount, &profile.FollowingsCount)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching followers and followings count"})
        return
    }

    // Fetch followers and followings in one query using UNION
    query := `
        SELECT u.id, u.username, 'follower' AS relation FROM followers f JOIN users u ON f.follower_id = u.id WHERE f.following_id = ?
        UNION ALL
        SELECT u.id, u.username, 'following' AS relation FROM followers f JOIN users u ON f.following_id = u.id WHERE f.follower_id = ?
    `
    rows, err := db.DB.Query(query, user.ID, user.ID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching followers and followings"})
        return
    }
    defer rows.Close()

    for rows.Next() {
        var id int
        var username, relation string
        if err := rows.Scan(&id, &username, &relation); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Error scanning follower/following"})
            return
        }
        if relation == "follower" {
            profile.Followers = append(profile.Followers, Follower{ID: id, Username: username})
        } else {
            profile.Followings = append(profile.Followings, Following{ID: id, Username: username})
        }
    }

    if err := rows.Err(); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error iterating through followers/followings"})
        return
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

    query := `
        SELECT 
            c.chat_id, 
            c.user1, 
            c.user2, 
            COALESCE(c.deleted_for, '') AS deleted_for, 
            COALESCE(MAX(m.created_at), '') AS last_message_time,
            COALESCE((SELECT message FROM messages WHERE chat_id = c.chat_id AND (deleted_for IS NULL OR deleted_for NOT LIKE ?) ORDER BY created_at DESC LIMIT 1), '') AS last_message,
            (SELECT profile_picture FROM users u WHERE u.username = CASE WHEN c.user1 = ? THEN c.user2 ELSE c.user1 END) AS profile_picture,
            (SELECT COUNT(*) FROM chat_notifications WHERE user_id = ? AND chat_id = c.chat_id AND is_read = false) AS unread_count
        FROM chats c
        LEFT JOIN messages m ON c.chat_id = m.chat_id
        WHERE (c.user1 = ? OR c.user2 = ?)
        GROUP BY c.chat_id, c.user1, c.user2, c.deleted_for
        HAVING last_message IS NOT NULL`
    
    rows, err := db.DB.Query(query, "%"+username+"%", username, userID, username, username)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching user chats"})
        return
    }
    defer rows.Close()

    var chats []map[string]interface{}
    for rows.Next() {
        var chatID, user1, user2, deletedFor, profilePicture string
        var lastMessageTime, lastMessage sql.NullString
        var unreadCount int

        if err := rows.Scan(&chatID, &user1, &user2, &deletedFor, &lastMessageTime, &lastMessage, &profilePicture, &unreadCount); err != nil {
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

        // Append 'Z' to indicate UTC time
        lastMessageTimeStr := lastMessageTime.String
        if lastMessageTimeStr != "" {
            lastMessageTimeStr += "Z"
        }

        chats = append(chats, map[string]interface{}{
            "chat_id":            chatID,
            "user":               otherUser,
            "unread_count":       unreadCount,
            "last_message_time":  lastMessageTimeStr,
            "profile_picture":    profilePicture,
            "last_message":       lastMessage.String, 
        })
        log.Print("Chats", chats)
    }

    if err := rows.Err(); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error iterating through chats"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"chats": chats})
}

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

    rows, err := db.DB.Query(`
        SELECT u.username, u.profile_picture 
        FROM users u
        JOIN chats c ON (c.user1 = u.username OR c.user2 = u.username)
        WHERE u.active = true AND u.username != ? AND (c.user1 = ? OR c.user2 = ?)`,
        authenticatedUsername, authenticatedUsername, authenticatedUsername)
    
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving active users"})
        return
    }
    defer rows.Close()

    var activeUsers []struct {
        Username       string `json:"username"`
        ProfilePicture string `json:"profile_picture"`
    }

    for rows.Next() {
        var user struct {
            Username       string `json:"username"`
            ProfilePicture string `json:"profile_picture"`
        }
        if err := rows.Scan(&user.Username, &user.ProfilePicture); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Error scanning active user"})
            return
        }
        activeUsers = append(activeUsers, user)
    }

    if err := rows.Err(); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error iterating through active users"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"active_users": activeUsers})
}
