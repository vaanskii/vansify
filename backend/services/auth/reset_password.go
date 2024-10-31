package auth

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/lpernett/godotenv"
	"github.com/vaanskii/vansify/db"
	"github.com/vaanskii/vansify/models"
	"github.com/vaanskii/vansify/utils"
	"gopkg.in/gomail.v2"
)

// SendResetPasswordEmail sends the reset password email
func sendResetPasswordEmail(email string, link string) error {
    godotenv.Load()
    m := gomail.NewMessage()
    m.SetHeader("From", os.Getenv("SMTP_USER"))
    m.SetHeader("To", email)
    m.SetHeader("Subject", "Reset Password")
    m.SetBody("text/html", "Please reset your password by clicking this link: <a href='" + link + "'>here</a>")
    port, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
    if err != nil {
        return err
    }
    d := gomail.NewDialer(os.Getenv("SMTP_SERVER"), port, os.Getenv("SMTP_USER"), os.Getenv("SMTP_PASS"))
    return d.DialAndSend(m)
}

// ForgotPassword handles sending a reset password email
func ForgotPassword(c *gin.Context) {
    godotenv.Load()
    var request struct {
        Email string `json:"email"`
    }
    if err := c.ShouldBindJSON(&request); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
        return
    }

    var user models.User
    err := db.DB.QueryRow("SELECT id, email FROM users WHERE email = ?", request.Email).Scan(&user.ID, &user.Email)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "User not found"})
        return
    }

    claims := &utils.CustomClaims{
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(10 * time.Second)),
            Subject:   strconv.FormatInt(user.ID, 10),
        },
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating token"})
        return
    }

    // Extract the origin from the request
    frontendURL := c.Request.Header.Get("Origin")
    if frontendURL == "" {
        frontendURL = "http://localhost:5173" 
    }

    resetLink := frontendURL + "/reset-password?token=" + tokenString
    err = sendResetPasswordEmail(user.Email, resetLink)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error sending email"})
        return
    }
    c.JSON(http.StatusOK, gin.H{"message": "Password reset email sent"})
}

// ResetPassword handles updating the user's password
func ResetPassword(c *gin.Context) {
    var request struct {
        Token       string `json:"token"`
        NewPassword string `json:"new_password"`
    }

    if err := c.ShouldBindJSON(&request); err != nil {
        log.Printf("Error binding JSON: %v\n", err) // Logging the error
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
        return
    }

    log.Printf("Received token: %s, new password: %s\n", request.Token, request.NewPassword)

    token, err := jwt.ParseWithClaims(request.Token, &utils.CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
        return []byte(os.Getenv("JWT_SECRET")), nil
    })
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
        return
    }

    claims, ok := token.Claims.(*utils.CustomClaims)
    if !ok || !token.Valid {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
        return
    }

    // Check if the token has already been used
    var tokenCount int
    err = db.DB.QueryRow("SELECT COUNT(*) FROM used_tokens WHERE token = ?", request.Token).Scan(&tokenCount)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking token"})
        return
    }
    if tokenCount > 0 {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Token has already been used"})
        return
    }

    userID, err := strconv.ParseInt(claims.Subject, 10, 64)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error parsing user ID"})
        return
    }

    var user models.User
    err = db.DB.QueryRow("SELECT id, password FROM users WHERE id = ?", userID).Scan(&user.ID, &user.Password)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving user"})
        return
    }

    // Check if the new password is the same as the current password
    if user.CheckPassword(request.NewPassword) {
        c.JSON(http.StatusBadRequest, gin.H{"error": "New password must be different from the current password"})
        return
    }

    user.Password = request.NewPassword
    if err := user.HashPassword(); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error hashing password"})
        return
    }

    _, err = db.DB.Exec("UPDATE users SET password = ? WHERE id = ?", user.Password, userID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating password"})
        return
    }

    // Mark the token as used
    _, err = db.DB.Exec("INSERT INTO used_tokens (token) VALUES (?)", request.Token)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error invalidating token"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Password successfully reset"})
}