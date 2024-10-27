package models

type Notification struct {
    ID        int64  `json:"id"`
    UserID    int64  `json:"user_id"`
    Message   string `json:"message"`
    IsRead    bool   `json:"is_read"`
}
