package auth

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
	"github.com/lpernett/godotenv"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
	"github.com/vaanskii/vansify/db"
	"github.com/vaanskii/vansify/models"
	"github.com/vaanskii/vansify/utils"
	"golang.org/x/oauth2"
)

const (
	key    = "RandomString"
	MaxAge = 86400 * 30
	IsProd = false
)

type UserRequest struct {
    Username string `json:"username" binding:"required"`
    Email    string `json:"email" binding:"required"`
}

type ContextKey string

func InitGoogleAuth() {
    // Load environment variables
    err := godotenv.Load()
    if err != nil {
        log.Fatalf("Error loading .env file: %v", err)
    }

    googleClientID := os.Getenv("GOOGLE_CLIENT_ID")
    googleClientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
    googleScopes := []string{
        "openid",                        
        "profile",                       
        "email",                         
        "https://www.googleapis.com/auth/drive.file",
        "https://www.googleapis.com/auth/drive.readonly",
        "https://www.googleapis.com/auth/drive",
    }

    store := sessions.NewCookieStore([]byte(key))
    store.Options = &sessions.Options{
        Path:     "/",
        MaxAge:   MaxAge,
        HttpOnly: true,
        Secure:   IsProd,
    }
    gothic.Store = store

    log.Println("Initializing Google provider with Client ID:", googleClientID)
    log.Println("Using Scopes:", googleScopes)

    // Initialize Google provider with updated scopes and access type
    googleProvider := google.New(googleClientID, googleClientSecret, "http://localhost:8080/v1/auth/google/callback", googleScopes...)
    googleProvider.SetAccessType("offline")
    goth.UseProviders(googleProvider)

    log.Println("Google provider initialized with scopes and offline access.")
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
        // Set provider in session
        session.Values["provider"] = provider
        err = session.Save(c.Request, c.Writer)
        if err != nil {
            log.Println("Error saving session in AuthHandler:", err)
        } else {
            log.Println("Session created in AuthHandler, session ID:", session.ID)
        }
    }

    gothic.BeginAuthHandler(c.Writer, c.Request)
}



func AuthCallback(c *gin.Context) {
    provider := c.Param("provider")
    type contextKey string
    const providerKey contextKey = "provider"
    ctx := context.WithValue(c.Request.Context(), providerKey, provider)
    c.Request = c.Request.WithContext(ctx)

    session, err := gothic.Store.Get(c.Request, "gothic-session")
    if err != nil {
        log.Println("Error retrieving session in AuthCallback:", err)
    } else {
        log.Println("Session retrieved in AuthCallback, session ID:", session.ID)
    }

    user, err := gothic.CompleteUserAuth(c.Writer, c.Request)
    if err != nil {
        c.String(http.StatusBadRequest, fmt.Sprint(err))
        return
    }

    token := &oauth2.Token{
        AccessToken:  user.AccessToken,
        RefreshToken: user.RefreshToken,
        Expiry:       user.ExpiresAt,
    }

    saveToken("token.json", token)

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

        c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("http://localhost:5173/auth/google/callback?email=%s&username=%s&access_token=%s&refresh_token=%s&id=%d", existingUser.Email, existingUser.Username, accessToken, refreshToken, existingUser.ID))
        return
    }

    c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("http://localhost:5173/choose-username?email=%s", user.Email))
}


func CreateUserWithUsername(c *gin.Context) {
    var userReq UserRequest
    if err := c.BindJSON(&userReq); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
        return
    }

    var existingUsername string
    err := db.DB.QueryRow("SELECT username FROM users WHERE username = ?", userReq.Username).Scan(&existingUsername)
    if err == nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Username already taken"})
        return
    }

    password := generatePassword()

    newUser := models.User{
        Username:       userReq.Username,
        Password:       password,
        Email:          userReq.Email,
        Verified:        true,
        OauthUser:      true,
    }

    err = newUser.HashPassword()
    if err != nil {
        c.String(http.StatusInternalServerError, fmt.Sprintf("Password hashing error: %v", err))
        return
    }

    result, err := db.DB.Exec("INSERT INTO users (username, password, email, verified, oauth_user) VALUES (?, ?, ?, ?, ?)", newUser.Username, newUser.Password, newUser.Email, newUser.Verified, newUser.OauthUser)
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

    c.JSON(http.StatusOK, gin.H{
        "access_token":  accessToken,
        "refresh_token": refreshToken,
        "username":      newUser.Username,
        "email":         newUser.Email,
        "id":            userID,
        "oauth_user":    newUser.OauthUser,
    })
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
