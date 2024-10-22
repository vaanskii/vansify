package utils

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/lpernett/godotenv"
)

type CustomClaims struct {
	Username   string   `json:"username"`
	jwt.RegisteredClaims
}

var jwtSecret []byte


// LoadEnv loads environment variables from the .env file
func LoadEnv() error {
	err := godotenv.Load()
	if err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}

	// Get the JWT Secret from .env
	jwtSecret = []byte(os.Getenv("JWT_SECRET"))
	return nil
}

// GenerateAccessToken generates a short-lived JWT access token for a user
func GenerateAccessToken(username string) (string, error) {
    expirationTime := time.Now().Add(15 * time.Minute)
    claims := &CustomClaims{
        Username: username,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(expirationTime),
            Subject:   username,
        },
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(jwtSecret)
}

// GenerateRefreshToken generates a long-lived JWT refresh token for a user
func GenerateRefreshToken(username string) (string, error) {
    expirationTime := time.Now().Add(7 * 24 * time.Hour)
    claims := &CustomClaims{
        Username: username,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(expirationTime),
            Subject:   username,
        },
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(jwtSecret)
}
// ValidateToken validates the token and returns the claims
func ValidateToken(tokenString string) (*jwt.RegisteredClaims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(t *jwt.Token) (interface{}, error) {
        return jwtSecret, nil
    })
    if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && token.Valid {
        return claims, nil
    }
    return nil, err
}


func VerifyJWT(tokenString string) (*jwt.Token, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
        return nil, err
    }
    if !token.Valid {
        return nil, fmt.Errorf("invalid token")
    }
    return token, nil
}


// RefreshToken handles refreshing the access token
func RefreshToken(c *gin.Context) {
    var request struct {
        RefreshToken string `json:"refresh_token"`
    }
    if err := c.ShouldBindJSON(&request); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
        return
    }
    claims, err := ValidateToken(request.RefreshToken)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token"})
        return
    }
    accessToken, err := GenerateAccessToken(claims.Subject)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating access token"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"access_token": accessToken})
}