package main

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/vaanskii/vansify/db"
	services "github.com/vaanskii/vansify/services/auth"
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

	// Routes
	r.POST("/register", services.RegisterUser)
	r.POST("/login", services.LoginUser)
	r.GET("/verify", services.VerifyEmail)
	
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "CORS-enabled route!",
		})
	})

	r.Run(":8080")
}