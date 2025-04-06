package search

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func SearchUsers(db *sql.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        query := c.Query("q")
        if query == "" {
            log.Println("[ERROR] Search query is empty")
            c.JSON(http.StatusBadRequest, gin.H{"error": "Search query cannot be empty"})
            return
        }

        log.Printf("[INFO] Received search request for query: %s\n", query)

        // ✅ Debugging log: Check SQL query execution
        rows, err := db.Query(`
            SELECT id, username, profile_picture, 
                MATCH(username) AGAINST (? IN NATURAL LANGUAGE MODE) AS relevance
            FROM users 
            WHERE MATCH(username) AGAINST (? IN NATURAL LANGUAGE MODE)
            OR username LIKE ? 
            ORDER BY relevance DESC, CHAR_LENGTH(username) ASC
        `, query, query, "%"+query+"%")

        if err != nil {
            log.Printf("[ERROR] Database query execution failed: %v\n", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
            return
        }
        defer rows.Close()

        var users []map[string]interface{}
        log.Println("[INFO] Processing search results...")

        for rows.Next() {
            var id int64
            var username string
            var profilePicture string
            var relevance float64

            if err := rows.Scan(&id, &username, &profilePicture, &relevance); err != nil {
                log.Printf("[ERROR] Failed to scan row data: %v\n", err)
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan results"})
                return
            }

            log.Printf("[DEBUG] Found user: ID=%d, Username=%s, Relevance=%f\n", id, username, relevance)

            users = append(users, gin.H{
                "id":             id,
                "username":       username,
                "profile_picture": profilePicture,
                "relevance":      relevance,
            })
        }

        if len(users) == 0 {
            log.Println("[WARN] No search results found")
        } else {
            log.Printf("[INFO] Search complete. %d results found.\n", len(users))
        }

        c.JSON(http.StatusOK, gin.H{"query": query, "results": users})
    }
}
