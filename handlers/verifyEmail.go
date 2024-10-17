package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/vaanskii/vansify/db"
)

func VerifyEmail(c *gin.Context){
	token := c.Query("token")
	parts := strings.Split(token, ":")
	if len(parts) != 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid token"})
		return
	}

	email := parts[1]
	// Find the user by email and update the verified status
	_, err := db.DB.Exec("UPDATE users SET verified = ? WHERE email =?", true, email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error verifying email"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Email verified successfully! You can now log in."})
}