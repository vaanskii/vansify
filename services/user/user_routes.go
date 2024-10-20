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
	err := db.DB.QueryRow("SELECT id, username, email, verified, created_at FROM users WHERE username = ?", username).
		Scan(&user.ID, &user.Username, &user.Email, &user.Verified, &user.CreatedAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Prepare the user profile response
	profile := UserProfile{
		ID: 			user.ID,
		Username:       user.Username,
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

// Func GetCurrentUser showing user which we are logged in now
func GetCurrentUser(c *gin.Context) {
    claims, exists := c.Get("claims")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
        return
    }

    customClaims := claims.(*utils.CustomClaims)
    c.JSON(http.StatusOK, gin.H{"username": customClaims.Username})
}


// GetUserChats function fetching logged in users chats
func GetUserChats(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}
	customClaims := claims.(*utils.CustomClaims)
	username := customClaims.Username

	rows, err := db.DB.Query("SELECT user1, user2 FROM chats WHERE user1 = ? OR user2 = ?", username, username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching user chats"})
		return
	}
	defer rows.Close()

	var userChats []string
	for rows.Next() {
		var user1, user2 string
		if err := rows.Scan(&user1, &user2); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error scanning chat"})
            return
		}

		// check if we only return the other user's info
		if user1 == username {
			userChats = append(userChats, user2)
		} else {
			userChats = append(userChats, user1)
		}
	}
	c.JSON(http.StatusOK, gin.H{"chats": userChats})
}