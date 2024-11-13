package auth

import (
	"database/sql"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lpernett/godotenv"
	"github.com/vaanskii/vansify/db"
	"github.com/vaanskii/vansify/models"
	"github.com/vaanskii/vansify/utils"
	"gopkg.in/gomail.v2"
)

func sendVerificationEmail(c *gin.Context, email string, token string) error {
    godotenv.Load()
    m := gomail.NewMessage()
    m.SetHeader("From", os.Getenv("SMTP_USER"))
    m.SetHeader("To", email)
    m.SetHeader("Subject", "Email Verification")

    // Extract the origin from the request
    frontendURL := c.Request.Header.Get("Origin")
    if frontendURL == "" {
        frontendURL = "http://localhost:5173" // Fallback to a default value if origin is not available
    }

    verificationLink := frontendURL + "/verify?token=" + token
    m.SetBody("text/html", "Please verify your email by clicking this link: <a href='" + verificationLink + "'>Verify Email</a>")

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

    if user.Gender != "male" && user.Gender != "female" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid gender value"})
        return
    }

    if user.Gender == "male" {
        user.ProfilePicture = "assets/images/man-picture.jpg"
    } else {
        user.ProfilePicture = "assets/images/woman-picture.jpg"
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
    _, err = db.DB.Exec("INSERT INTO users (username, password, email, profile_picture, gender, verified) VALUES (?, ?, ?, ?, ?, ?)", 
        user.Username, user.Password, user.Email, user.ProfilePicture, user.Gender, user.Verified)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error saving user to database"})
        return
    }

    // Generate a verification token
    token := GenerateVerificationToken(user.Email)

    // Send verification email
    if err := sendVerificationEmail(c, user.Email, token); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error sending verification email"})
        return
    }

    c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully! Please verify your email."})
}


// LoginUser handles user login
func LoginUser(c *gin.Context) {
    var request struct {
        Username   string `json:"username"`
        Password   string `json:"password"`
        RememberMe bool   `json:"remember_me"`
    }
    if err := c.ShouldBindJSON(&request); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
        return
    }
    // Check if user exists
    row := db.DB.QueryRow("SELECT id, username, email, password, verified FROM users WHERE username = ?", request.Username)
    var dbUser models.User
    if err := row.Scan(&dbUser.ID, &dbUser.Username, &dbUser.Email, &dbUser.Password, &dbUser.Verified); err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
        return
    }
    // Check password
    if !dbUser.CheckPassword(request.Password) {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
        return
    }
    if !dbUser.Verified { // Check if verified
        c.JSON(http.StatusForbidden, gin.H{"error": "Please verify your email before logging in."})
        return
    }
    // Generate tokens
    accessToken, err := utils.GenerateAccessToken(request.Username)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating access token"})
        return
    }
    refreshToken, err := utils.GenerateRefreshToken(request.Username)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating refresh token"})
        return
    }

    // Set refresh token in a cookie if "Remember Me" is checked
    if request.RememberMe {
        c.SetCookie("refresh_token", refreshToken, 7*24*3600, "/", "", false, true)
    }

    // Include user details in the response
    c.JSON(http.StatusOK, gin.H{
        "access_token": accessToken,
        "refresh_token": refreshToken,
        "id": dbUser.ID,
        "username": dbUser.Username,
        "email": dbUser.Email,
    })
}

// Delete user function
func DeleteUser(c *gin.Context) {
    // Retrieve the claims from the context set by the middleware
    claims, exists := c.Get("claims")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "No claims found"})
        return
    }

    // Assuming claims is of type *utils.CustomClaims
    customClaims, ok := claims.(*utils.CustomClaims)
    if !ok {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
        return
    }

    // Get username from claims
    username := customClaims.Username

    // Delete the user from the database
    _, err := db.DB.Exec("DELETE FROM users WHERE username = ?", username)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting user account"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Account deleted successfully"})
}
