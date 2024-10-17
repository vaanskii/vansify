package services

import (
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
	followerUsername := claims.Subject
	followingUsername := c.Param("username")

	// Check if the following user exists
	var exists bool
	err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = ?)", followingUsername).Scan(&exists)
	if err != nil || !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User does not exist"})
		return
	}

	if followerUsername == followingUsername {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You cannot follow yourself"})
		return
	}

	// Check if the follow relationship already exists
	var followExists bool
	err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM followers WHERE follower_username = ? AND following_username = ?)", followerUsername, followingUsername).Scan(&followExists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking follow status"})
		return
	}

	if followExists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You are already following this user"})
		return
	}

	// Create a Follow relationship
	_, err = db.DB.Exec("INSERT INTO followers (follower_username, following_username) VALUES (?, ?)", followerUsername, followingUsername)
	if err != nil {
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
	followerUsername := claims.Subject
	followingUsername := c.Param("username")

	// Check if the follow relationship exists
	var followExists bool
	err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM followers WHERE follower_username = ? AND following_username = ?)", followerUsername, followingUsername).Scan(&followExists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking follow status"})
		return
	}

	if !followExists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You are not following this user"})
		return
	}

	// Remove the Follow relationship
	_, err = db.DB.Exec("DELETE FROM followers WHERE follower_username = ? AND following_username = ?", followerUsername, followingUsername)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unfollowing user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully unfollowed user"})
}
