package user

import (
	"database/sql"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/vaanskii/vansify/db"
)

var (
    clients     = make(map[*websocket.Conn]string)  // Store username with each connection
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
        log.Fatal("Error upgrading to WebSocket:", err)
    }
    username := c.Query("username")
    log.Printf("Client connected: %s, Username: %s", ws.RemoteAddr(), username)

    clientMutex.Lock()
    clients[ws] = username  // Store the username in the clients map
    clientMutex.Unlock()

    defer func() {
        log.Printf("Client disconnected: %s, Username: %s", ws.RemoteAddr(), username)
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
                    // User reconnected, no need to update inactive status
                    return
                }
            }

            _, err := db.DB.Exec("UPDATE users SET active = false WHERE username = ?", username)
            if err != nil {
                log.Printf("Error updating user active status for username %s: %v", username, err)
            } else {
                log.Printf("Updated user active status to inactive for username %s", username)
            }

            // Broadcast updated active users
            go FetchActiveUsersAndBroadcast(db.DB)
        }(username)
    }()

    // Mark user as active immediately upon connection
    _, err = db.DB.Exec("UPDATE users SET active = true WHERE username = ?", username)
    if err != nil {
        log.Printf("Error updating user active status for username %s: %v", username, err)
    } else {
        log.Printf("Updated user active status to active for username %s", username)
    }

    go FetchActiveUsersAndBroadcast(db.DB)

    for {
        var msg []string
        err := ws.ReadJSON(&msg)
        if err != nil {
            log.Printf("Read error: %v, closing connection for client: %s, Username: %s", err, ws.RemoteAddr(), username)
            clientMutex.Lock()
            delete(clients, ws)
            clientMutex.Unlock()
            break
        }
        log.Printf("Received message from client: %s, Username: %s, Message: %v", ws.RemoteAddr(), username, msg)
    }
}

// Broadcast messages to all clients
func HandleMessages() {
    for {
        msg := <-broadcast
        clientMutex.Lock()
        for client := range clients {
            err := client.WriteJSON(msg)
            if err != nil {
                log.Printf("Error broadcasting to client: %v, Client: %s", err, client.RemoteAddr())
                client.Close()
                delete(clients, client)
            }
        }
        clientMutex.Unlock()
    }
}

func FetchActiveUsersAndBroadcast(db *sql.DB) {
    log.Println("Fetching active users")

    var activeUsers []struct {
        Username       string `json:"username"`
        ProfilePicture string `json:"profile_picture"`
    }

    rows, err := db.Query("SELECT username, profile_picture FROM users WHERE active = true")
    if err != nil {
        log.Printf("Error fetching active users: %v", err)
        return
    }
    defer rows.Close()

    for rows.Next() {
        var user struct {
            Username       string `json:"username"`
            ProfilePicture string `json:"profile_picture"`
        }
        if err := rows.Scan(&user.Username, &user.ProfilePicture); err != nil {
            log.Printf("Error scanning user: %v", err)
            return
        }
        activeUsers = append(activeUsers, user)
    }
    if err := rows.Err(); err != nil {
        log.Printf("Error iterating over rows: %v", err)
        return
    }

    log.Printf("Broadcasting %d active users", len(activeUsers))
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
        log.Printf("Broadcasting to client: %s, Username: %s", client.RemoteAddr(), username)
        err := client.WriteJSON(usersToSend)
        if err != nil {
            log.Printf("Error broadcasting to client: %v, Client: %s", err, client.RemoteAddr())
            client.Close()
            delete(clients, client)
        }
    }
    clientMutex.Unlock()
    log.Println("Broadcast complete")
}
