package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
	"github.com/lpernett/godotenv"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
	"github.com/vaanskii/vansify/db"
	"github.com/vaanskii/vansify/models"
	activeUsers "github.com/vaanskii/vansify/services/user"
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

func init() {
    gob.Register(time.Time{})
}

func InitGoogleAuth() {
    if err := godotenv.Load(); err != nil {
        log.Printf("Error loading .env file init google auth: %v", err)
    }

    googleClientID := os.Getenv("GOOGLE_CLIENT_ID")
    googleClientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
    backendUrl := os.Getenv("BACKEND_URL")

    if googleClientID == "" || googleClientSecret == "" || backendUrl == "" {
        log.Fatal("Critical environment variables are missing")
    }

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

    // Initialize Google provider with updated scopes and access type
    googleProvider := google.New(googleClientID, googleClientSecret, backendUrl+"/v1/auth/google/callback", googleScopes...)
    googleProvider.SetAccessType("offline")
    goth.UseProviders(googleProvider)
}

func AuthHandler(c *gin.Context) {
    provider := c.Param("provider")
    allowedProviders := map[string]bool{
        "google": true,
    }

    if !allowedProviders[provider] {
        c.String(http.StatusBadRequest, "You must select a valid provider")
        return
    }

    c.Request.URL.RawQuery = "provider=" + provider
    session, err := gothic.Store.Get(c.Request, "gothic-session")
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
        return
    }

    session.Values["provider"] = provider
    if err := session.Save(c.Request, c.Writer); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session"})
        return
    }

    url, err := gothic.GetAuthURL(c.Writer, c.Request)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get auth URL"})
        return
    }

    url += "&prompt=select_account"
    http.Redirect(c.Writer, c.Request, url, http.StatusTemporaryRedirect)
}

// Create a thread-safe in-memory store for temporary tokens
var tokenStore = struct {
    sync.Mutex
    tokens map[string]string // Maps token -> email
}{tokens: make(map[string]string)}

// GenerateShortToken creates a random 10-character token
func GenerateShortToken() string {
    b := make([]byte, 8) // Generate 8 random bytes
    rand.Read(b)
    return base64.RawURLEncoding.EncodeToString(b)[:10] // Encode as string and trim to 10 characters
}

func AuthCallback(c *gin.Context) {
    provider := c.Param("provider")
    type contextKey string
    const providerKey contextKey = "provider"
    ctx := context.WithValue(c.Request.Context(), providerKey, provider)
    c.Request = c.Request.WithContext(ctx)

    _, err := gothic.Store.Get(c.Request, "gothic-session")
    if err != nil {
        log.Println("Error retrieving session in AuthCallback:", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve session"})
        return
    }

    user, err := gothic.CompleteUserAuth(c.Writer, c.Request)
    if err != nil {
        log.Println("Error completing user auth:", err)
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
    err = db.DB.QueryRow("SELECT id, username, password, email, oauth_user, active FROM users WHERE email = ?", user.Email).Scan(
        &existingUser.ID, &existingUser.Username, &existingUser.Password, &existingUser.Email, &existingUser.OauthUser, &existingUser.Active,
    )
    if err == nil {
        log.Println("User already exists:", existingUser.Email)

        // Generate tokens for existing user
        accessToken, err := utils.GenerateAccessToken(existingUser.Username, existingUser.Email)
        if err != nil {
            log.Println("Error generating access token:", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error generating access token: %v", err)})
            return
        }

        refreshToken, err := utils.GenerateRefreshToken(existingUser.Username, existingUser.Email)
        if err != nil {
            log.Println("Error generating refresh token:", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error generating refresh token: %v", err)})
            return
        }

        frontendUrl := os.Getenv("FRONTEND_URL")
        redirectURL := fmt.Sprintf("%s/auth/google/callback?username=%s&access_token=%s&refresh_token=%s&id=%d&oauth_user=%t&active=%t",
            frontendUrl, url.QueryEscape(existingUser.Username), url.QueryEscape(accessToken), url.QueryEscape(refreshToken), existingUser.ID, existingUser.OauthUser, existingUser.Active)

        // Send the response first
        c.Redirect(http.StatusTemporaryRedirect, redirectURL)
        return
    }

    // User does not exist: Generate a one-time token
    shortToken := GenerateShortToken()

    // Store the token and associated email in memory (thread-safe)
    tokenStore.Lock()
    tokenStore.tokens[shortToken] = user.Email // Link token to email securely
    tokenStore.Unlock()

    // Clean up the token after 10 minutes (time-bound single-use token)
    go func(token string) {
        time.Sleep(10 * time.Minute)
        tokenStore.Lock()
        delete(tokenStore.tokens, token) // Remove token after expiration
        tokenStore.Unlock()
    }(shortToken)

    // Redirect with the token
    frontendUrl := os.Getenv("FRONTEND_URL")
    redirectURL := fmt.Sprintf("%s/setauth?token=%s", frontendUrl, url.QueryEscape(shortToken))
    c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

func ValidateOauthToken(c *gin.Context) {
    var request struct {
        Token string `json:"token" binding:"required"`
    }
    if err := c.BindJSON(&request); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
        return
    }

    tokenStore.Lock()
    email, exists := tokenStore.tokens[request.Token]
    tokenStore.Unlock()

    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"email": email}) // Return the associated email
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
        Active:         true,
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

    accessToken, err := utils.GenerateAccessToken(newUser.Username, newUser.Email)
    if err != nil {
        c.String(http.StatusInternalServerError, fmt.Sprintf("Error generating access token: %v", err))
        return
    }

    refreshToken, err := utils.GenerateRefreshToken(newUser.Username, newUser.Email)
    if err != nil {
        c.String(http.StatusInternalServerError, fmt.Sprintf("Error generating refresh token: %v", err))
        return
    }

    // Broadcast the active users
    activeUsers.FetchActiveUsersAndBroadcast(db.DB)

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


func saveToken(path string, token *oauth2.Token) {
    fmt.Printf("Saving credential file to: %s\n", path)
    f, err := os.Create(path)
    if err != nil {
        log.Fatalf("Unable to save oauth token: %v", err)
    }
    defer f.Close()

    json.NewEncoder(f).Encode(token)
}