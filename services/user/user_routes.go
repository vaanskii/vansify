package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vaanskii/vansify/db"
	"github.com/vaanskii/vansify/models"
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
