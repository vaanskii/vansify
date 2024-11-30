package main

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/vaanskii/vansify/db"
	notifications "github.com/vaanskii/vansify/notifications"
	"github.com/vaanskii/vansify/notifications/chat_notifications"
	auth "github.com/vaanskii/vansify/services/auth"
	"github.com/vaanskii/vansify/services/aws"
	"github.com/vaanskii/vansify/services/chat"
	follow "github.com/vaanskii/vansify/services/follow"
	user "github.com/vaanskii/vansify/services/user"
	"github.com/vaanskii/vansify/utils"
)

func main() {
    db.ConnectToDatabase()
    auth.InitGoogleAuth()

    aws.InitAWSSession()

    r := gin.Default()

    r.Use(func(c *gin.Context){
        c.Writer.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' https://trusted.cdn.com; style-src 'self' https://trusted.cdn.com; img-src 'self' data: https://trusted.cdn.com; connect-src 'self' https://api.trusted.com")
        c.Next()
    })

    r.HandleMethodNotAllowed = true

    r.NoRoute(func(c *gin.Context) {
        c.JSON(http.StatusNotFound, gin.H{"error": "Page Not Found"})
    })
    r.NoMethod(func(c *gin.Context) {
        c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "Method Not Allowed"})
    })

    r.Use(cors.New(cors.Config{
        AllowOrigins:     []string{"https://vansify.vercel.app", "http://localhost:5173"},
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept"},
        ExposeHeaders:    []string{"Content-Length", "ETag", "x-amz-server-side-encryption", "x-amz-access-control-allow-origin"},
        AllowCredentials: false,
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
        v1.POST("/logout", auth.AuthMiddleware(), auth.LogoutUser)

        //google auth
        v1.GET("/auth/:provider", auth.AuthHandler) 
        v1.GET("/auth/:provider/callback", auth.AuthCallback)
        v1.POST("/create-user", auth.CreateUserWithUsername)
        
        // refresh token
        v1.POST("/refresh-token", utils.RefreshToken)

        // aws s3
        v1.POST("/upload/chat/:chatid", aws.UploadFile)
        v1.POST("/upload/profile/:username", aws.UploadFile)

        // Follow/Unfollow system Routes
        v1.POST("/follow/:username", auth.AuthMiddleware(), follow.FollowUser)
        v1.DELETE("/unfollow/:username", auth.AuthMiddleware(), follow.UnfollowUser)
        v1.GET("/is-following/:follower/:following", auth.AuthMiddleware(), follow.CheckFollowStatus)
        v1.GET("/followers/:username", auth.AuthMiddleware(), follow.GetFollowers)
        v1.GET("/following/:username", auth.AuthMiddleware(), follow.GetFollowing)
        v1.GET("/notifications/ws", auth.AuthMiddleware(), notifications.NotificationWsHandler)

        // Chat routes
        v1.POST("/create-chat", auth.AuthMiddleware(), chat.CreateChat)
        v1.GET("/chat/:chatID/ws", auth.AuthMiddleware(), chat.ChatWsHandler)
        v1.GET("/chat/:chatID/history", auth.AuthMiddleware(), chat.GetChatHistory)
        v1.GET("/check-chat/:user1/:user2", auth.AuthMiddleware(), chat.CheckChatExists)
        v1.GET("/notifications/chat/unread", auth.AuthMiddleware(), chat_notifications.GetUnreadChatNotifications)
        v1.POST("/notifications/chat/mark-read/:chatID", auth.AuthMiddleware(), chat.MarkChatNotificationsAsRead)
        v1.DELETE("/chat/:chatID", auth.AuthMiddleware(), chat.DeleteChat)
        v1.DELETE("/chat/:chatID/delete-messages", auth.AuthMiddleware(), chat.DeleteUserMessages)
        v1.DELETE("/message/:messageID", auth.AuthMiddleware(), chat.DeleteMessage)
        v1.GET("/chat-notifications/ws", auth.AuthMiddleware(), chat_notifications.ChatNotificationWsHandler)

        // User Profile Retrieval
        v1.GET("/me/chats", auth.AuthMiddleware(), user.GetUserChats)
        v1.GET("/user/:username", user.GetUserByUsername)
        v1.GET("/active-users", auth.AuthMiddleware(), user.GetActiveUsersHandler)
        v1.GET("/active-users/ws", user.HandleConnections)

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

    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    r.Run(":" + port)
}
