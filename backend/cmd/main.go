package main

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/vaanskii/vansify/db"
	notifications "github.com/vaanskii/vansify/notifications"
	auth "github.com/vaanskii/vansify/services/auth"
	"github.com/vaanskii/vansify/services/chat"
	follow "github.com/vaanskii/vansify/services/follow"
	user "github.com/vaanskii/vansify/services/user"
	"github.com/vaanskii/vansify/utils"
)

func main() {
    db.ConnectToDatabase()

    r := gin.Default()

    // Enable global handling of Method Not Allowed
    r.HandleMethodNotAllowed = true

    // Handle non-existent and incorrect methods routes
    r.NoRoute(func(c *gin.Context) {
        c.JSON(http.StatusNotFound, gin.H{"error": "Page Not Found"})
    })
    r.NoMethod(func(c *gin.Context) {
        c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "Method Not Allowed"})
    })

    r.Use(cors.New(cors.Config{
        AllowAllOrigins:  true,
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: true,
        MaxAge:           12 * time.Hour,
    }))

    r.Static("/assets", "./assets")

    v1 := r.Group("/v1")
    {
        // Authorization Routes
        v1.POST("/register", auth.RegisterUser)
        v1.POST("/login", auth.LoginUser)
        v1.GET("/verify", auth.VerifyEmail)
        v1.DELETE("/delete-account", auth.AuthMiddleware(), auth.DeleteUser)
        v1.POST("/forgot-password", auth.ForgotPassword)
        v1.POST("/reset-password", auth.ResetPassword)

        // refresh token
        v1.POST("/refresh-token", utils.RefreshToken)

        // Follow/Unfollow system Routes
        v1.POST("/follow/:username", auth.AuthMiddleware(), follow.FollowUser)
        v1.DELETE("/unfollow/:username", auth.AuthMiddleware(), follow.UnfollowUser)
        v1.GET("/is-following/:follower/:following", auth.AuthMiddleware(), follow.CheckFollowStatus)
        v1.GET("/followers/:username", auth.AuthMiddleware(), follow.GetFollowers)
        v1.GET("/following/:username", auth.AuthMiddleware(), follow.GetFollowing)
        v1.GET("/notifications/ws", auth.AuthMiddleware(), notifications.NotificationWsHandler)

        // Chat routes
        v1.POST("/create-chat", auth.AuthMiddleware(), chat.CreateChat)
        v1.GET("/chat/:chatID", auth.AuthMiddleware(), chat.WsHandler)
        v1.GET("/chat/:chatID/history", auth.AuthMiddleware(), chat.GetChatHistory)
        v1.GET("/check-chat/:user1/:user2", auth.AuthMiddleware(), chat.CheckChatExists)
        v1.GET("/notifications/chat/unread", auth.AuthMiddleware(), notifications.GetUnreadChatNotifications)
        v1.POST("/notifications/chat/mark-read/:chatID", auth.AuthMiddleware(), notifications.MarkChatNotificationsAsRead)
        v1.DELETE("/chat/:chatID", auth.AuthMiddleware(), chat.DeleteChat)
        v1.DELETE("/message/:messageID", auth.AuthMiddleware(), chat.DeleteMessage)

        // User Profile Retrieval
        v1.GET("/me/chats", auth.AuthMiddleware(), user.GetUserChats)
        v1.GET("/user/:username", user.GetUserByUsername)

        // General Notifications
        v1.GET("/notifications", auth.AuthMiddleware(), notifications.GetNotifications)
        v1.GET("/notifications/count", auth.AuthMiddleware(), notifications.GetUnreadNotificationCount) 
        v1.POST("/notifications/general/mark-read/:notificationID", auth.AuthMiddleware(), notifications.MarkNotificationAsRead)
        v1.DELETE("/notifications/delete/:notificationID", auth.AuthMiddleware(), notifications.DeleteNotification)
    }

    r.GET("/", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{
            "message": "CORS-enabled route!",
        })
    })

    // Use the PORT environment variable if available, otherwise default to 8080
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    r.Run(":" + port)
}
