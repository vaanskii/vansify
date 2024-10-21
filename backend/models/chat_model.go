package models

type Message struct {
	ID 		  int 	   `json:"id"`
	ChatID    string   `json:"chat_id"`
	Message   string   `json:"message"`
	Username  string   `json:"username"`
}

type Chat struct {
	ID       int        `json:"id"`
	ChatID   string     `json:"chat_id"`
	User1    string 	`json:"user1"`
    User2    string     `json:"user2"`
}

