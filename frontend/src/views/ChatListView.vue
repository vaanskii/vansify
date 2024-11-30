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
    <ul v-if="chatStore.chats.length > 0">
      <li v-for="chat in sortedChats" :key="chat.chat_id">
        <router-link 
          :to="{ name: 'chat', params: { chatID: chat.chat_id }, query: { user: chat.user } }" 
          @click.native="markChatAsRead(chat.chat_id)"
        >
          <img :src="chat.profile_picture" alt="Profile Picture" width="30" height="30" />
          {{ chat.user }}
          <span v-if="chat.unread_count > 0">({{ chat.unread_count }})</span>
          <br>
          <span>{{ formatTime(chat.last_message_time) }}</span> - 
          <span>{{ chat.last_message }}</span>
          <span v-if="isUserActive(chat.user)" class="active-indicator">●</span>
        </router-link>
        <button @click="deleteChat(chat.chat_id)">Delete</button>
        <button @click="deleteChatforUser(chat.chat_id)">delete for me</button>
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

const store = userStore();
const chatStore = useChatStore();
const activeUsersStore = useActiveUsersStore();
const chatNotificationStore = useChatNotificationStore()

const sortedChats = computed(() => chatStore.sortedChats);

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
  return chatStore.formatTime(timestamp);
};

const isUserActive = (username) => {
  return chatStore.activeUsers.some(user => user.username === username);
};
</script>

<style>
.active-indicator {
  color: green;
  margin-left: 5px;
}
</style>
