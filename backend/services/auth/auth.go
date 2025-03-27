package auth

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lpernett/godotenv"
	"github.com/vaanskii/vansify/db"
	"github.com/vaanskii/vansify/models"
	"github.com/vaanskii/vansify/services/chat"
	activeUsers "github.com/vaanskii/vansify/services/user"

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

    // Set default profile picture if not provided
    if user.ProfilePicture == "" {
        user.ProfilePicture = "https://cdn.pixabay.com/photo/2015/10/05/22/37/blank-profile-picture-973460_1280.png"
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
    _, err = db.DB.Exec("INSERT INTO users (username, password, email, profile_picture, verified, oauth_user) VALUES (?, ?, ?, ?, ?, ?)", 
        user.Username, user.Password, user.Email, user.ProfilePicture, user.Verified, user.OauthUser)
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

    if !dbUser.Verified {
        c.JSON(http.StatusForbidden, gin.H{"error": "Please verify your email before logging in."})
        return
    }

    // Generate tokens
    accessToken, err := utils.GenerateAccessToken(request.Username, dbUser.Email)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating access token"})
        return
    }
    refreshToken, err := utils.GenerateRefreshToken(request.Username, dbUser.Email)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating refresh token"})
        return
    }

    if request.RememberMe {
        c.SetCookie("refresh_token", refreshToken, 7*24*3600, "/", "", false, true)
    }

    _, err = db.DB.Exec("UPDATE users SET active = true, last_active = NULL WHERE username = ?", dbUser.Username)
    if err != nil {
        // Log the error but don't stop the login process
        log.Println("Error updating user active status:", err)
    }

    c.JSON(http.StatusOK, gin.H{
        "access_token": accessToken,
        "refresh_token": refreshToken,
        "id": dbUser.ID,
        "username": dbUser.Username,
        "email": dbUser.Email,
        "oauth_user": dbUser.OauthUser,
        "active": true,
    })

    // Trigger status update for messages
    go func() {
        _, err := db.DB.Exec("UPDATE users SET active = true, last_active = NULL WHERE email = ?", dbUser.Email)
        if err != nil {
            log.Println("Error updating user active status:", err)
            return
        }

        // Fetch active users and broadcast
        activeUsers.FetchActiveUsersAndBroadcast(db.DB)

        // Update message statuses for all chats involving the user
        rows, err := db.DB.Query("SELECT chat_id, user1, user2 FROM chats WHERE user1 = ? OR user2 = ?", dbUser.Username, dbUser.Username)
        if err != nil {
            log.Println("Error querying chats for user:", err)
            return
        }
        defer rows.Close()

        for rows.Next() {
            var chatID string
            var user1 string
            var user2 string
            if err := rows.Scan(&chatID, &user1, &user2); err != nil {
                log.Println("Error scanning chat ID:", err)
                continue
            }

            // Determine the other user in the chat
            var otherUser string
            if dbUser.Username == user1 {
                otherUser = user2
            } else {
                otherUser = user1
            }

            go chat.UpdateStatusWhenUserBecomesActive(chatID, dbUser.Username, otherUser)
        }
    }()

    go activeUsers.FetchActiveUsersAndBroadcast(db.DB)
}

func LogoutUser(c *gin.Context) {
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

    // Update the user's active status to false and set last_active to current timestamp
    _, err := db.DB.Exec("UPDATE users SET active = ?, last_active = NOW() WHERE username = ?", false, username)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
        return
    }

    c.SetCookie("refresh_token", "", -1, "/", "", false, true)
    c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})

    // Broadcast active users update asynchronously
    go activeUsers.FetchActiveUsersAndBroadcast(db.DB)
}


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
    result, err := db.DB.Exec("DELETE FROM users WHERE username = ?", username)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting user account"})
        return
    }

    rowsAffected, _ := result.RowsAffected()
    if rowsAffected == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Account deleted successfully"})
}
