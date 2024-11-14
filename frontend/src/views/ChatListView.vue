<template>
  <h1 v-if="chatStore.loader">Loading</h1>
  <div v-else>
    <h1>Your Chats</h1>
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
        </router-link>
        <button @click="deleteChat(chat.chat_id)">Delete</button>
      </li>
    </ul>
    <div v-else>No chats found</div>
    <div v-if="chatStore.error" class="error">{{ chatStore.error }}</div>
  </div>
</template>


<script setup>
import { onMounted, onUnmounted, computed } from 'vue';
import { useChatStore } from '@/stores/chatStore';
import { userStore } from '@/stores/user';
import emitter from '@/eventBus';

const store = userStore();
const chatStore = useChatStore();

const sortedChats = computed(() => chatStore.sortedChats);

onMounted(() => {
  if (store.user.isAuthenticated) {
    if (chatStore.chats.length === 0) {
      chatStore.fetchChats();
    }
    emitter.on('ws-open', chatStore.handleWebSocketOpen);
    emitter.on('ws-message', chatStore.handleWebSocketMessage);
    emitter.on('ws-error', chatStore.handleWebSocketError);
    emitter.on('ws-close', chatStore.handleWebSocketClose);
    setInterval(chatStore.updateMessageTimes, 60000);
  }
});

onUnmounted(() => {
  emitter.off('ws-open', chatStore.handleWebSocketOpen);
  emitter.off('ws-message', chatStore.handleWebSocketMessage);
  emitter.off('ws-error', chatStore.handleWebSocketError);
  emitter.off('ws-close', chatStore.handleWebSocketClose);
});

const deleteChat = (chatID) => {
  chatStore.deleteChat(chatID);
};

const markChatAsRead = (chatID) => {
  chatStore.markChatAsRead(chatID);
};

const formatTime = (timestamp) => {
  return chatStore.formatTime(timestamp);
};
</script>
