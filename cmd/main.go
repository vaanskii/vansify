package main

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/vaanskii/vansify/db"
	services_auth "github.com/vaanskii/vansify/services/auth"
	services_follow "github.com/vaanskii/vansify/services/follow"
)

func main() {
	db.ConnectToDatabase()
	// db.CreateTable()  if i want to make migrations automatically. now am using makefile migrations for it.

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Authorization Routes
	r.POST("/register", services_auth.RegisterUser)
	r.POST("/login", services_auth.LoginUser)
	r.GET("/verify", services_auth.VerifyEmail)
	r.DELETE("/delete-account", services_auth.DeleteUser)


	// Follow/Unfollow system Routers
	r.POST("/follow/:username", services_follow.FollowUser)      
	r.DELETE("/unfollow/:username", services_follow.UnfollowUser)

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "CORS-enabled route!",
		})
	})

	r.Run(":8080")
}