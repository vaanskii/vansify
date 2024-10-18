package follow

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/vaanskii/vansify/db"
	"github.com/vaanskii/vansify/utils"
)

func FollowUser(c *gin.Context) {
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization token is required"})
		return
	}
	tokenParts := strings.Split(tokenString, " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
		return
	}
	tokenString = tokenParts[1]

	claims, err := utils.ValidateJWT(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	followerUsername := claims.Subject // The username of the logged-in user
	followingUsername := c.Param("username") // The username of the user to follow

	// Get follower ID from follower username
	var followerID int64
	err = db.DB.QueryRow("SELECT id FROM users WHERE username = ?", followerUsername).Scan(&followerID)
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

	// Create a Follow relationship
	_, err = db.DB.Exec("INSERT INTO followers (follower_id, following_id) VALUES (?, ?)", followerID, followingID)
	if err != nil {
		log.Printf("Error creating follow relationship: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error following user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Successfully followed user"})
}


func UnfollowUser(c *gin.Context) {
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization token is required"})
		return
	}

	tokenParts := strings.Split(tokenString, " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
		return
	}

	tokenString = tokenParts[1]

	claims, err := utils.ValidateJWT(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}
	followerUsername := claims.Subject // The username of the logged-in user
	followingUsername := c.Param("username") // The username of the user to unfollow

	// Get follower ID from follower username
	var followerID int64
	err = db.DB.QueryRow("SELECT id FROM users WHERE username = ?", followerUsername).Scan(&followerID)
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
