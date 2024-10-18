package auth

import (
	"database/sql"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lpernett/godotenv"
	"github.com/vaanskii/vansify/db"
	"github.com/vaanskii/vansify/models"
	"github.com/vaanskii/vansify/utils"
	"gopkg.in/gomail.v2"
)

func sendVerificationEmail(email string, token string) error {
	godotenv.Load()
	m := gomail.NewMessage()
	m.SetHeader("From", os.Getenv("SMTP_USER"))
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Email Verification")
	m.SetBody("text/html", "Please verify your email by clicking this link: <a href='http://localhost:8080/v1/verify?token="+token+"'>Verify Email</a>")

	// Convert SMTP_PORT from string to int
	port, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil {
		return err 
	}

	d := gomail.NewDialer(os.Getenv("SMTP_SERVER"), port, os.Getenv("SMTP_USER"), os.Getenv("SMTP_PASS"))

	// Send the email
	if err := d.DialAndSend(m); err != nil {
		return err
	}

	return nil
}

// RegisterUser handles user registration
func RegisterUser(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate that everything is provided
	if user.Username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username is required"})
		return
	}
	
	if user.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password is required"})
		return
	} else if len(user.Password) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password should be at least 8 characters long"})
		return
	}

	if user.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
		return
	}

	// Check if username already exists
	var existingUsername string
	err := db.DB.QueryRow("SELECT username FROM users WHERE username = ?", user.Username).Scan(&existingUsername)
	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username already exists. Please choose another one."})
		return
	} else if err != sql.ErrNoRows {
		// Handle potential database error
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking existing username"})
		return
	}

	// Check if email already exists
	var existingEmail string
	err = db.DB.QueryRow("SELECT email FROM users WHERE email = ?", user.Email).Scan(&existingEmail)
	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email already exists. Please use another one."})
		return
	} else if err != sql.ErrNoRows {
		// Handle potential database error
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking existing email"})
		return
	}

	// Hash the user's password
	if err := user.HashPassword(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error hashing password"})
		return
	}

	user.Verified = false

	// Save user to the database
	_, err = db.DB.Exec("INSERT INTO users (username, password, email, verified) VALUES (?, ?, ?, ?)", user.Username, user.Password, user.Email, user.Verified)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error saving user to database"})
		return
	}

	// Generate a verification token
	token := GenerateVerificationToken(user.Email)

	// Send verification email
	if err := sendVerificationEmail(user.Email, token); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error sending verification email"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully! Please verify your email."})
}


// LoginUser handles user login
func LoginUser(c *gin.Context) {
    var user models.User
    if err := c.ShouldBindJSON(&user); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Check if user exists
    row := db.DB.QueryRow("SELECT id, password, verified FROM users WHERE username = ?", user.Username) 
    var dbUser models.User
    if err := row.Scan(&dbUser.ID, &dbUser.Password, &dbUser.Verified); err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
        return
    }

    // Check password
    if !dbUser.CheckPassword(user.Password) {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
        return
    }

    if !dbUser.Verified { // Check if verified
        c.JSON(http.StatusForbidden, gin.H{"error": "Please verify your email before logging in."})
        return
    }

    // Generate JWT
    token, err := utils.GenerateJWT(user.Username)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating token"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"token": token})
}

// Delete user function
func DeleteUser(c *gin.Context) {
	// Extract the token from the Authorization header
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization token is required"})
		return
	}

	// Remove the "Bearer " prefix from the token
	tokenParts := strings.Split(tokenString, " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
		return
	}

	tokenString = tokenParts[1]

	// Validate the token
	claims, err := utils.ValidateJWT(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}
	// Get user id from claims
	username := claims.Subject
	
	// Delete the user from the database
	_, err = db.DB.Exec("DELETE FROM users WHERE username = ?", username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting user account"})
		return 
	}

	c.JSON(http.StatusOK, gin.H{"message": "Account deleted successfully"})
}