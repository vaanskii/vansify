package auth

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
	"github.com/lpernett/godotenv"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
	"github.com/vaanskii/vansify/db"
	"github.com/vaanskii/vansify/models"
	"github.com/vaanskii/vansify/utils"
)

const (
	key    = "RandomString"
	MaxAge = 86400 * 30
	IsProd = false
)

type ContextKey string

func InitGoogleAuth() {
	log.Println("Initializing Google authentication...")
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file:", err)
	}

	googleClientID := os.Getenv("GOOGLE_CLIENT_ID")
	googleClientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")

	store := sessions.NewCookieStore([]byte(key))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   MaxAge,
		HttpOnly: true,
		Secure:   IsProd,
	}
	gothic.Store = store

	// Initialize Google provider
	goth.UseProviders(
		google.New(googleClientID, googleClientSecret, "http://localhost:8080/v1/auth/google/callback"),
	)
	log.Println("Google provider initialized.")
}

func AuthHandler(c *gin.Context) {
	provider := c.Param("provider")
	log.Println("AuthHandler triggered with provider:", provider)
	if provider == "" {
		c.String(http.StatusBadRequest, "You must select a provider")
		log.Println("Error: No provider specified.")
		return
	}

	// Set provider as query parameter for Gothic
	c.Request.URL.RawQuery = "provider=" + provider

	// Attempt to start the authentication session
	session, err := gothic.Store.Get(c.Request, "gothic-session")
	if err != nil {
		log.Println("Error creating session in AuthHandler:", err)
	} else {
		log.Println("Session created in AuthHandler, session ID:", session.ID)
	}

	// Begin authentication process with Gothic
	gothic.BeginAuthHandler(c.Writer, c.Request)
}


func AuthCallback(c *gin.Context) {
    provider := c.Param("provider")

    // Add provider to context
    type contextKey string
    const providerKey contextKey = "provider"
    ctx := context.WithValue(c.Request.Context(), providerKey, provider)
    c.Request = c.Request.WithContext(ctx)

    // Attempt to retrieve session
    session, err := gothic.Store.Get(c.Request, "gothic-session")
    if err != nil {
        log.Println("Error retrieving session in AuthCallback:", err)
    } else {
        log.Println("Session retrieved in AuthCallback, session ID:", session.ID)
    }

    // Complete authentication and get user
    user, err := gothic.CompleteUserAuth(c.Writer, c.Request)
    if err != nil {
        c.String(http.StatusBadRequest, fmt.Sprint(err))
        return
    }

    // Check if the user already exists in the database
    var existingUser models.User
    err = db.DB.QueryRow("SELECT id, username, password, email FROM users WHERE email = ?", user.Email).Scan(&existingUser.ID, &existingUser.Username, &existingUser.Password, &existingUser.Email)
    if err == nil {
        log.Println("User already exists:", existingUser.Email)
        // Generate tokens for existing user
        accessToken, err := utils.GenerateAccessToken(existingUser.Username)
        if err != nil {
            log.Println("Error generating access token:", err)
            c.String(http.StatusInternalServerError, fmt.Sprintf("Error generating access token: %v", err))
            return
        }

        refreshToken, err := utils.GenerateRefreshToken(existingUser.Username)
        if err != nil {
            log.Println("Error generating refresh token:", err)
            c.String(http.StatusInternalServerError, fmt.Sprintf("Error generating refresh token: %v", err))
            return
        }

        // Redirect to frontend with user information and tokens
        c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("http://localhost:5173/auth/google/callback?email=%s&username=%s&access_token=%s&refresh_token=%s&id=%d", existingUser.Email, existingUser.Username, accessToken, refreshToken, existingUser.ID))
        return
    }

    // Generate a secure password
    password := generatePassword()

    // Extract and clean username from email
    newUsername := cleanUsername(user.Email)

    // Create a User model instance
    newUser := models.User{
        Username: newUsername,
        Password: password,
        Email:    user.Email,
    }

    // Hash the password
    err = newUser.HashPassword()
    if err != nil {
        c.String(http.StatusInternalServerError, fmt.Sprintf("Password hashing error: %v", err))
        return
    }

    // Insert user into the database
    result, err := db.DB.Exec("INSERT INTO users (username, password, email) VALUES (?, ?, ?)", newUser.Username, newUser.Password, newUser.Email)
    if err != nil {
        log.Println("Database error:", err)
        c.String(http.StatusInternalServerError, fmt.Sprintf("Database error: %v", err))
        return
    }
    
    userID, _ := result.LastInsertId()

    accessToken, err := utils.GenerateAccessToken(newUser.Username)
    if err != nil {
        c.String(http.StatusInternalServerError, fmt.Sprintf("Error generating access token: %v", err))
        return
    }

    refreshToken, err := utils.GenerateRefreshToken(newUser.Username)
    if err != nil {
        c.String(http.StatusInternalServerError, fmt.Sprintf("Error generating refresh token: %v", err))
        return
    }

    c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("http://localhost:5173/auth/google/callback?email=%s&username=%s&access_token=%s&refresh_token=%s&id=%d", newUser.Email, newUser.Username, accessToken, refreshToken, userID))
}


func cleanUsername(email string) string {
    username := email[:strings.Index(email, "@")]
    cleanedUsername := ""
    for _, char := range username {
        if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') {
            cleanedUsername += string(char)
        }
    }
    return cleanedUsername
}

func generatePassword() string {
    const passwordLength = 12
    chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()-_=+"

    password := make([]byte, passwordLength)
    for i := range password {
        index, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
        if err != nil {
            fmt.Println("Error generating random number:", err)
            return ""
        }
        password[i] = chars[index.Int64()]
    }

    return string(password)
}
