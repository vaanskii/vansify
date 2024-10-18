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

	// Authorization Routes
	r.POST("/register", auth.RegisterUser)
	r.POST("/login", auth.LoginUser)
	r.GET("/verify", auth.VerifyEmail)
	r.DELETE("/delete-account", auth.DeleteUser)


	// Follow/Unfollow system Routers
	r.POST("/follow/:username", follow.FollowUser)      
	r.DELETE("/unfollow/:username", follow.UnfollowUser)

	
	r.GET("/user/:username", user.GetUserProfile)

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "CORS-enabled route!",
		})
	})

	r.Run(":8080")
}