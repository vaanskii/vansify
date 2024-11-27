package user

import "sync"

type ChatManager struct {
	activeChats map[string]map[string]bool
	mux         sync.Mutex
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

func (cm *ChatManager) RemoveUserFromChat(chatID string, username string) {
	cm.mux.Lock()
	defer cm.mux.Unlock()
	if cm.activeChats[chatID] != nil {
		delete(cm.activeChats[chatID], username)
	}
}

func (cm *ChatManager) IsUserInChat(chatID string, username string) bool {
	cm.mux.Lock()
	defer cm.mux.Unlock()
	if cm.activeChats[chatID] != nil {
		return cm.activeChats[chatID][username]
	}
	return false
}
