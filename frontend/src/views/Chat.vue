<template>
  <div>
    <h2>Chat with {{ chatUser }}</h2>
    <div v-for="message in formattedMessages" :key="message.id">
      <strong v-if="message.username">{{ message.username }}:</strong>
      {{ message.message }}
    </div>
    <form @submit.prevent="sendMessage">
      <input v-model="newMessage" placeholder="Type a message" required />
      <button type="submit">Send</button>
    </form>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, computed } from 'vue';
import axios from 'axios';
import { useRoute } from 'vue-router';
import { userStore } from '@/stores/user';

const apiUrl = import.meta.env.VITE_WS_URL;
const messages = ref([]);
const newMessage = ref('');
const route = useRoute();
const store = userStore();
const username = store.user.username;
let ws;

// Define chatUser
const chatUser = ref('');

const formattedMessages = computed(() => {
  return messages.value.map(message => {
    if (message.username === username) {
      return { ...message, username: '' }; // Set username to an empty string for your own messages
    }
    return message;
  });
});

onMounted(async () => {
  const token = store.user.access;
  const chatID = route.params.chatID;

  if (chatID && token) {
    try {
      const response = await axios.get(`/v1/chat/${chatID}/history`)
      if (response.data) {
        messages.value = response.data.map(message => {
          if (message.username === username) {
            message.username = '';
          }
          return message;
        });
      }
    } catch (error) {
      console.error('Error fetching chat history:', error);
    }

    chatUser.value = route.query.user || 'Unknown';

    const wsURL = `ws://${apiUrl}/v1/chat/${chatID}?token=${encodeURIComponent(token)}`;
    ws = new WebSocket(wsURL);

    ws.onopen = () => {
      console.log('WebSocket connection established');
    };

    ws.onerror = (error) => {
      console.error('WebSocket error:', error);
    };

    ws.onclose = (event) => {
      console.log('WebSocket connection closed:', event);
    };

    ws.onmessage = (event) => {
      try {
        const message = JSON.parse(event.data);
        if (message.username === username) {
          message.username = '';
        }
        messages.value.push(message);
      } catch (e) {
        console.error("Error parsing message:", e);
      }
    };
  } else {
    console.error('No chat ID or token provided');
  }
});

onUnmounted(() => {
  if (ws) ws.close();
});

const sendMessage = () => {
  if (ws && newMessage.value) {
    const message = {
      username,
      message: newMessage.value
    };
    ws.send(JSON.stringify(message));
    messages.value.push({ ...message, username: '' }); // Set username to an empty string for your own messages in real-time
    newMessage.value = '';
  }
};
</script>
