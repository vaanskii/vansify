package chat

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/vaanskii/vansify/db"
	"github.com/vaanskii/vansify/models"
	"github.com/vaanskii/vansify/utils"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize: 1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	hub = NewHub()
)

func generateChatID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
        return "", err
    }
    return hex.EncodeToString(bytes), nil
}

func CreateChat(c *gin.Context) {
	var chat models.Chat
	if err := c.ShouldBindJSON(&chat); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
	}

	// Check if chat already exists
	var existingChat string
	err := db.DB.QueryRow("SELECT chat_id FROM chats WHERE (user1 = ? or user2 = ?) OR (user1 = ? user2 = ?)",
		chat.User1, chat.User2, chat.User2, chat.User1).Scan(&existingChat)
	if err == nil {
		c.JSON(http.StatusOK, gin.H{"chat_id": existingChat})
		return
	} else if err != sql.ErrNoRows {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking existing chat"})  // Database Error
        return
	}

	 // if chat does not exist, create a new one
	chatID, err := generateChatID()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating chat ID"})
		return
	}

	chat.ChatID = chatID
	_, execErr := db.DB.Exec("INSERT INTO chats (chat_id, user1, user2) VALUES (?, ?, ?)", chat.ChatID, chat.User1, chat.User2)
	if execErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error saving chat to database"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"chat_id": chat.ChatID})
}

func WsHandler(c *gin.Context) {
	chatID := c.Param("chatID")
	token := c.Query("token")

	log.Printf("Attempting to upgrade connection for chatID: %s with token: %s", chatID, token)

	// Validate token
	token = strings.TrimSpace(token)
	log.Printf("Token after trimming: %s", token)

	parsedToken, err := utils.VerifyJWT(token)
	if err != nil {
        log.Println("Invalid token:", err)
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }

	claims, ok := parsedToken.Claims.(*utils.CustomClaims)
	if !ok || !parsedToken.Valid {
		log.Println("Invalid claims or token not valid")
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
	}

	log.Printf("Token valid for user: %s", claims.Username)

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WebSocket Upgrade error:", err)
		return
	}
	log.Println("WebSocket connection established for chatID:", chatID)
    defer conn.Close()

	// Add the new connection to the hub
	hub.AddConnection(conn)
	defer hub.RemoveConnection(conn)

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println("WebSocket ReadMessage error:", err)
            return
		}

		var incomingMessage models.Message
		if err := json.Unmarshal(p, &incomingMessage); err != nil {
            log.Println("Error decoding incoming message:", err)
            continue
        }

		incomingMessage.ChatID = chatID
		incomingMessage.Username = claims.Username

		// Save message to database
		_, execErr := db.DB.Exec("INSERT INTO messages (chat_id, message, username) VALUES (?, ?, ?)",
			incomingMessage.ChatID, incomingMessage.Message, incomingMessage.Username)
		if execErr != nil {
			log.Println("DB Exec error:", execErr)
			return
		}

		// Broadcast the message to all connected clients
		broadcastMessage, _ := json.Marshal(incomingMessage)
		hub.BroadcastMessage(conn, messageType, broadcastMessage)
	}
}

func GetChatHistory(c *gin.Context) {
	chatID := c.Param("chatID")
	rows, err := db.DB.Query("SELECT id, chat_id, message, username FROM messages where chat_id = ?", chatID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching chat history"})
		return
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var message models.Message
		if err := rows.Scan(&message.ID, &message.ChatID, &message.Message, &message.Username); 
		err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error scanning message"})
			return
		}
		messages = append(messages, message)
	}
	c.JSON(http.StatusOK, messages)
}