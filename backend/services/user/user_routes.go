package user

import (
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
    err := db.DB.QueryRow("SELECT id, username, email, profile_picture, gender, verified, created_at FROM users WHERE username = ?", username).Scan(&user.ID, &user.Username, &user.Email, &user.ProfilePicture, &user.Gender, &user.Verified, &user.CreatedAt)
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
    customClaims := claims.(*utils.CustomClaims)
    username := customClaims.Username

    rows, err := db.DB.Query("SELECT chat_id, user1, user2 FROM chats WHERE user1 = ? OR user2 = ?", username, username)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching user chats"})
        return
    }
    defer rows.Close()

    var chats []map[string]string
    for rows.Next() {
        var chatID, user1, user2 string
        if err := rows.Scan(&chatID, &user1, &user2); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Error scanning chat"})
            return
        }
        otherUser := user1
        if user1 == username {
            otherUser = user2
        }
        chats = append(chats, map[string]string{"chat_id": chatID, "user": otherUser})
    }

    c.JSON(http.StatusOK, gin.H{"chats": chats})
}
