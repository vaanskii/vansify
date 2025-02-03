<template>
  <div>
    <h1>Your Chats</h1>
    <h2>Active Users</h2>
    <ul v-if="chatStore.activeUsers.length > 0">
      <li v-for="user in chatStore.activeUsers" :key="user.username">
        <img :src="user.profile_picture" alt="Profile Picture" width="30" height="30" />
        {{ user.username }}
      </li>
    </ul>
    <div v-else>No active users found</div>

    <h2>Your Chats</h2>
    <ul v-if="chatStore.chats.length > 0" class="chat-list">
  <li 
    v-for="chat in sortedChats" 
    :key="chat.chat_id" 
    :class="['chat-item', { active: isActiveChat(chat.chat_id) }]"
  >
    <router-link 
      :to="{ name: 'chat', params: { chatID: chat.chat_id }, query: { user: chat.user } }" 
      @click.native="markChatAsRead(chat.chat_id)" 
      class="chat-link"
    >
      <img :src="chat.profile_picture" alt="Profile Picture" class="profile-picture" />
      <div class="chat-details">
        <span class="chat-user">{{ chat.user }}</span>
        <span v-if="chat.unread_count > 0" class="unread-count">({{ chat.unread_count }})</span>
        <br />
        <span class="last-message-time">{{ formatTime(chat.last_message_time) }}</span> - 
        <span class="last-message">{{ chat.last_message }}</span>
        <span v-if="isUserActive(chat.user)" class="active-indicator">●</span>
      </div>
    </router-link>
    <div class="chat-actions">
      <button @click="deleteChat(chat.chat_id)" class="delete-button">Delete</button>
      <button @click="deleteChatforUser(chat.chat_id)" class="delete-button">Delete for Me</button>
    </div>
  </li>
</ul>

    <div v-else>No chats found</div>
    <div v-if="chatStore.error" class="error">{{ chatStore.error }}</div>
  </div>
</template>

<script setup>
import { onMounted, onUnmounted, computed } from 'vue';
import { useChatStore } from '@/stores/chatStore';
import { useActiveUsersStore } from '@/stores/activeUsers';
import { useChatNotificationStore } from '@/stores/chatNotification';
import { userStore } from '@/stores/user';
import emitter from '@/eventBus';
import { useRoute } from 'vue-router'; 

const route = useRoute(); 

const store = userStore();
const chatStore = useChatStore();
const activeUsersStore = useActiveUsersStore();
const chatNotificationStore = useChatNotificationStore()

const sortedChats = computed(() => chatStore.sortedChats);

const isActiveChat = (chatID) => {
  return route.params.chatID === chatID;
};

onMounted(() => {
  if (store.user.isAuthenticated) {
    if (chatStore.chats.length === 0) {
      chatStore.fetchChats();
    }
    
    emitter.on('chat-ws-open', chatStore.handleWebSocketOpen);
    emitter.on('chat-ws-message', chatStore.handleWebSocketMessage);
    emitter.on('chat-ws-error', chatStore.handleWebSocketError);
    emitter.on('chat-ws-close', chatStore.handleWebSocketClose);
    activeUsersStore.connectWebSocket();
    chatNotificationStore.connectWebSocket();

    emitter.on('active-users-ws-open', chatStore.handleWebSocketOpen);
    emitter.on('active-users-ws-error', chatStore.handleWebSocketError);
    emitter.on('active-users-ws-close', chatStore.handleWebSocketClose);
    emitter.on('active-users-fetched', chatStore.handleActiveUsersFetched);

    chatStore.fetchActiveUsers();

    setInterval(chatStore.updateMessageTimes, 60000);
  }
});

onUnmounted(() => {
  emitter.off('ws-open', chatStore.handleWebSocketOpen);
  emitter.off('ws-message', chatStore.handleWebSocketMessage);
  emitter.off('ws-error', chatStore.handleWebSocketError);
  emitter.off('ws-close', chatStore.handleWebSocketClose);

  if (chatStore.wsConnected) {
    chatStore.handleWebSocketClose();
  }
  emitter.off('active-users-ws-open', chatStore.handleWebSocketOpen);
  emitter.off('active-users-ws-error', chatStore.handleWebSocketError);
  emitter.off('active-users-ws-close', chatStore.handleWebSocketClose);
  emitter.off('active-users-fetched', chatStore.handleActiveUsersFetched);
});

const deleteChat = (chatID) => {
  chatStore.deleteChat(chatID);
};

const deleteChatforUser = (chatID) => {
  chatStore.deleteMessagesForUser(chatID)
}

const markChatAsRead = (chatID) => {
  chatStore.markChatAsRead(chatID);
};

const formatTime = (timestamp) => {
  return chatStore.formatTimeAgo(timestamp);
};

const isUserActive = (username) => {
  return chatStore.activeUsers.some(user => user.username === username);
};
</script>

<style>
.chat-list {
  list-style: none;
  margin: 0;
  padding: 0;
}

.chat-item {
  display: flex;
  align-items: center;
  padding: 10px;
  border-bottom: 1px solid #ddd;
  transition: background-color 0.3s ease, color 0.3s ease;
}

.chat-item.active {
  background-color: #e6f7ff; /* Light blue for active chats */
  color: #1890ff;
  font-weight: bold;
}

.chat-link {
  display: flex;
  flex: 1;
  align-items: center;
  text-decoration: none;
  color: inherit;
}

.profile-picture {
  border-radius: 50%;
  margin-right: 10px;
}

.chat-details {
  flex-grow: 1;
}

.chat-user {
  font-size: 16px;
  font-weight: bold;
}

.unread-count {
  color: red;
}

.last-message-time {
  font-size: 12px;
  color: gray;
}

.last-message {
  font-size: 14px;
}

.active-indicator {
  color: green;
  font-size: 14px;
  margin-left: 5px;
}

.chat-actions {
  display: flex;
  gap: 10px;
}

.delete-button {
  background-color: #ff4d4f;
  color: white;
  border: none;
  padding: 5px 10px;
  border-radius: 4px;
  cursor: pointer;
  transition: background-color 0.2s ease;
}

.delete-button:hover {
  background-color: #ff7875;
}

</style>
