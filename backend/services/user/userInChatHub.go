package user

import (
	"sync"
)

type ChatManager struct {
    activeChats map[string]map[string]bool
    mux         sync.RWMutex
}

var ChatManagerInstance = &ChatManager{
    activeChats: make(map[string]map[string]bool),
}

// AddUserToChat marks a user as active in a chat
func (cm *ChatManager) AddUserToChat(chatID string, username string) {
    cm.mux.Lock()
    defer cm.mux.Unlock()
    if cm.activeChats[chatID] == nil {
        cm.activeChats[chatID] = make(map[string]bool)
    }
    cm.activeChats[chatID][username] = true
}

// RemoveUserFromChat removes a user from a chat
func (cm *ChatManager) RemoveUserFromChat(chatID string, username string) {
    cm.mux.Lock()
    defer cm.mux.Unlock()
    if chat, exists := cm.activeChats[chatID]; exists {
        delete(chat, username)
        if len(chat) == 0 {
            delete(cm.activeChats, chatID)
        }
    }
}

// IsUserInChat checks if a user is active in a chat
func (cm *ChatManager) IsUserInChat(chatID string, username string) bool {
    cm.mux.RLock()
    defer cm.mux.RUnlock()
    if chat, exists := cm.activeChats[chatID]; exists {
        return chat[username]
    }
    return false
}
