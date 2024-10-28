<template>
  <div>
    <h1>Your Chats</h1>
    <ul>
      <li v-for="chat in sortedChats" :key="chat.chat_id">
        <router-link :to="{ name: 'chat', params: { chatID: chat.chat_id }, query: { user: chat.user } }">
          {{ chat.user }}
          <span v-if="chat.unread_count > 0">({{ chat.unread_count }})</span> <!-- Show unread count per chat -->
        </router-link>
      </li>
    </ul>
    <div v-if="error" class="error">{{ error }}</div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, computed } from 'vue';
import axios from 'axios';
import { userStore } from '@/stores/user';

const chats = ref([]);
const error = ref('');
const apiUrl = import.meta.env.VITE_WS_URL;
const store = userStore();
const wsUrl = `ws://${apiUrl}/v1/notifications/ws?token=${encodeURIComponent(store.user.access)}`;
let ws;

const fetchChats = async () => {
  try {
    const response = await axios.get('/v1/me/chats');
    chats.value = response.data.chats.map(chat => ({
      chat_id: chat.chat_id,
      user: chat.user,
      unread_count: chat.unread_count,
      last_message_time: chat.last_message_time
    }));
    console.log("Fetched chats with unread counts:", chats.value);
  } catch (err) {
    error.value = err.response ? err.response.data.error : 'An error occurred';
    console.error("Error fetching chats:", err);
  }
};

// Sort chats by the last received message time
const sortedChats = computed(() => {
  console.log("Sorting chats based on last message time", chats.value);  // Debug to ensure sorting
  return chats.value.slice().sort((a, b) => new Date(b.last_message_time) - new Date(a.last_message_time));
});

const connectNotificationWebSocket = () => {
  ws = new WebSocket(wsUrl);
  ws.onopen = () => {
    console.log("Notification WebSocket connection established");
  };
  ws.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data);
      console.log("WebSocket message received:", data);

      // Check if the chat already exists in the list
      const chatIndex = chats.value.findIndex(chat => chat.chat_id === data.chat_id);
      if (chatIndex !== -1) {
        // Update existing chat
        chats.value[chatIndex] = {
          ...chats.value[chatIndex],
          unread_count: data.unread_count,
          last_message_time: data.last_message_time || new Date().toISOString(),
          message: data.message,
          user: data.user
        };
      } else {
        // Add new chat
        chats.value.push({
          chat_id: data.chat_id,
          user: data.user,
          unread_count: data.unread_count,
          last_message_time: data.last_message_time || new Date().toISOString(),
          message: data.message
        });
      }

      // Trigger sort
      chats.value = chats.value.slice().sort((a, b) => new Date(b.last_message_time) - new Date(a.last_message_time));
      console.log("Updated sorted chat list:", chats.value);
    } catch (e) {
      console.error("Error processing WebSocket message:", e);
    }
  };
  ws.onerror = (error) => {
    console.error("Notification WebSocket error: ", error);
  };
  ws.onclose = () => {
    console.log("Notification WebSocket connection closed");
  };
};


onMounted(() => {
  if (store.user.isAuthenticated) {
    fetchChats();
    connectNotificationWebSocket();
  }
});

onUnmounted(() => {
  if (ws) ws.close();
});
</script>
