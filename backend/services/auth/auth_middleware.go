package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/vaanskii/vansify/utils"
)

func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        var token string
        authHeader := c.GetHeader("Authorization")
        if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
            token = strings.TrimPrefix(authHeader, "Bearer ")
        }

        if token == "" {
            token = c.Query("token")
        }

        if token == "" {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
            return
        }

        parsedToken, err := utils.VerifyJWT(token)
        if err != nil {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
            return
        }

        claims, ok := parsedToken.Claims.(*utils.CustomClaims)
        if !ok || !parsedToken.Valid {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
            return
        }

        c.Set("claims", claims)
        c.Next()
    }
}
