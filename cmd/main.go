package main

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/vaanskii/vansify/db"
	auth "github.com/vaanskii/vansify/services/auth"
	follow "github.com/vaanskii/vansify/services/follow"
	user "github.com/vaanskii/vansify/services/user"
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

	v1 := r.Group("/v1")
	{
		// Authorization Routes
		v1.POST("/register", auth.RegisterUser)
		v1.POST("/login", auth.LoginUser)
		v1.GET("/verify", auth.VerifyEmail)
		v1.DELETE("/delete-account", auth.DeleteUser)

		// Follow/Unfollow system Routes
		v1.POST("/follow/:username", follow.FollowUser)
		v1.DELETE("/unfollow/:username", follow.UnfollowUser)

		// User Profile Retrieval
		v1.GET("/user/:username", user.GetUserByUsername)
	}

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "CORS-enabled route!",
		})
	})

	r.Run(":8080")
}