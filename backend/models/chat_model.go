package models

import "time"

type Message struct {
    ID          int       `json:"id"`
    ChatID      string    `json:"chat_id"`
    Message     string    `json:"message"`
    Username    string    `json:"username"`
    FileURL     string    `json:"file_url,omitempty"`
    CreatedAt   time.Time `json:"created_at"`
    DeletedFor  string    `json:"deleted_for"`
    Status      string    `json:"status"`
}

type Chat struct {
    ID       int    `json:"id"`
    ChatID   string `json:"chat_id"`
    User1    string `json:"user1"`
    User2    string `json:"user2"`
}
