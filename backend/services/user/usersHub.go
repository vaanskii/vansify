package user

import (
	"database/sql"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/vaanskii/vansify/db"
)

var (
    clients     = make(map[*websocket.Conn]string)
    broadcast   = make(chan []struct {
        Username       string `json:"username"`
        ProfilePicture string `json:"profile_picture"`
    })
    upgrader = websocket.Upgrader{
        CheckOrigin: func(r *http.Request) bool {
            return true
        },
    }
    clientMutex = &sync.Mutex{}
)

// Handle WebSocket connections
func HandleConnections(c *gin.Context) {
    ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upgrade to WebSocket"})
        return
    }
    username := c.Query("username")

    clientMutex.Lock()
    clients[ws] = username
    clientMutex.Unlock()

    defer func() {
        ws.Close()

        clientMutex.Lock()
        delete(clients, ws)
        clientMutex.Unlock()

        // Schedule the inactive status update after 5 minutes
        go func(username string) {
            time.Sleep(5 * time.Minute)

            clientMutex.Lock()
            defer clientMutex.Unlock()

            // Check if the user is still disconnected
            for _, connectedUsername := range clients {
                if connectedUsername == username {
                    // User reconnected, reset becoming_inactive status
                    db.DB.Exec("UPDATE users SET becoming_inactive = FALSE WHERE username = ?", username)
                    return
                }
            }            

            _, err := db.DB.Exec("UPDATE users SET active = false, becoming_inactive = FALSE, last_active = NOW() WHERE username = ?", username)
            if err == nil {
                FetchActiveUsersAndBroadcast(db.DB)
            }
        }(username)
    }()

    _, err = db.DB.Exec("UPDATE users SET active = true, last_active = NULL, becoming_inactive = FALSE WHERE username = ?", username)
    if err == nil {
        FetchActiveUsersAndBroadcast(db.DB)
    }

    for {
        var msg []string
        err := ws.ReadJSON(&msg)
        if err != nil {
            clientMutex.Lock()
            delete(clients, ws)
            clientMutex.Unlock()

            _, err := db.DB.Exec("UPDATE users SET becoming_inactive = TRUE WHERE username = ?", username)
            if err == nil {
                break
            }
        }
    }
}

func HandleMessages() {
    for {
        msg := <-broadcast
        clientMutex.Lock()
        for client := range clients {
            err := client.WriteJSON(msg)
            if err != nil {
                client.Close()
                delete(clients, client)
            }
        }
        clientMutex.Unlock()
    }
}

func FetchActiveUsersAndBroadcast(db *sql.DB) {
    var activeUsers []struct {
        Username       string `json:"username"`
        ProfilePicture string `json:"profile_picture"`
    }

    rows, err := db.Query("SELECT username, profile_picture FROM users WHERE active = true")
    if err != nil {
        return
    }
    defer rows.Close()

    for rows.Next() {
        var user struct {
            Username       string `json:"username"`
            ProfilePicture string `json:"profile_picture"`
        }
        if err := rows.Scan(&user.Username, &user.ProfilePicture); err != nil {
            return
        }
        activeUsers = append(activeUsers, user)
    }
    if err := rows.Err(); err != nil {
        return
    }

    clientMutex.Lock()
    for client, username := range clients {
        // Exclude the broadcasting user
        usersToSend := []struct {
            Username       string `json:"username"`
            ProfilePicture string `json:"profile_picture"`
        }{}
        for _, user := range activeUsers {
            if user.Username != username {
                usersToSend = append(usersToSend, user)
            }
        }
        err := client.WriteJSON(usersToSend)
        if err != nil {
            client.Close()
            delete(clients, client)
        }
    }
    clientMutex.Unlock()
}
