package notifications

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var notificationUpgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin:     func(r *http.Request) bool { return true },
}

func NotificationWsHandler(c *gin.Context) {
    conn, err := notificationUpgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        log.Println("WebSocket Upgrade error:", err)
        return
    }
    defer conn.Close()

    GlobalNotificationHub.AddConnection(conn)
    defer GlobalNotificationHub.RemoveConnection(conn)

    for {
        messageType, p, err := conn.ReadMessage()
        if err != nil {
            log.Println("WebSocket ReadMessage error:", err)
            return
        }
        log.Printf("Received: %s", p)
        if err := conn.WriteMessage(messageType, p); err != nil {
            log.Println("WebSocket WriteMessage error:", err)
            return
        }
    }
}
