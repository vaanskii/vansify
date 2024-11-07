package models

type NotificationType string

const (
    FollowNotificationType NotificationType = "FOLLOW"
)
type Notification struct {
    ID        int64           `json:"id"`
    UserID    int64           `json:"user_id"`
    Message   string          `json:"message"`
    IsRead    bool            `json:"is_read"`
    Type      NotificationType `json:"type"`
    CreatedAt string          `json:"created_at"`
}

type ChatNotification struct {
    ID        int64  `json:"id"`
    UserID    int64  `json:"user_id"`
    Message   string `json:"message"`
    IsRead    bool   `json:"is_read"`
}
