<template>
  <div>
    <h2>Chat with {{ chatUser }}</h2>
    <div v-if="formattedMessages.length === 0 && !isLoading">No messages yet</div>
    <div v-if="!isLoading">
      <div v-for="message in formattedMessages" :key="message.id">
        <strong v-if="message.username && !message.isOwnMessage">{{ message.username }}</strong>
        {{ message.message }}
        <span v-if="message.isOwnMessage">
          <span v-if="message.status === true">(Sent)</span>
          <span v-if="message.status === false">(Not Sent)</span>
        </span>
      </div>
    </div>
    <form @submit.prevent="sendMessage" v-if="!isLoading">
      <input v-model="newMessage" placeholder="Type a message" required />
      <button type="submit">Send</button>
    </form>
  </div>
</template>


<script setup>
import { ref, onMounted, onUnmounted, computed, watch } from 'vue';
import axios from 'axios';
import { useRoute } from 'vue-router';
import { userStore } from '@/stores/user';

const apiUrl = import.meta.env.VITE_WS_URL;
const messages = ref([]);
const newMessage = ref('');
const isLoading = ref(true);
const isConnected = ref(false);
const route = useRoute();
const store = userStore();
const username = store.user.username;
let ws;
let retryAttempt = 0;
const maxRetries = 10;
const offlineStorageKey = `offline-messages-${route.params.chatID}`;
const chatUser = ref('');

// Compute formatted messages
const formattedMessages = computed(() => {
  return messages.value.map(message => {
    const isOwnMessage = message.username === username;
    return { ...message, isOwnMessage };
  });
});

// Load messages from localStorage
const loadOfflineMessages = () => {
  const savedMessages = localStorage.getItem(offlineStorageKey);
  if (savedMessages) {
    const parsedMessages = JSON.parse(savedMessages);
    parsedMessages.forEach(msg => {
      messages.value.push({ ...msg, isOwnMessage: true, status: false });
    });
  }
};

// Save unsent messages to localStorage
const saveOfflineMessages = () => {
  const unsentMessages = messages.value.filter(msg => msg.status === false && msg.isOwnMessage);
  localStorage.setItem(offlineStorageKey, JSON.stringify(unsentMessages));
};

// Remove offline messages from localStorage once sent
const removeOfflineMessages = () => {
  localStorage.removeItem(offlineStorageKey);
};

// WebSocket connection logic
const connectWebSocket = (chatID, token) => {
  const wsURL = `ws://${apiUrl}/v1/chat/${chatID}?token=${encodeURIComponent(token)}`;
  ws = new WebSocket(wsURL);
  ws.onopen = () => {
    console.log('WebSocket connection established');
    isConnected.value = true;
    retryAttempt = 0;
    isLoading.value = false;
    // Resend unsent messages
    const unsentMessages = messages.value.filter(msg => msg.status === false && msg.isOwnMessage);
    unsentMessages.forEach(message => {
      ws.send(JSON.stringify({ message: message.message, username }));
      message.status = true;
      console.log('Resent message:', message);
    });
    removeOfflineMessages();
  };
  ws.onmessage = (event) => {
    const message = JSON.parse(event.data);
    messages.value.push(message);
    console.log('Received message:', message);
  };
  ws.onerror = (error) => {
    console.error('WebSocket error:', error);
    isConnected.value = false;
  };
  ws.onclose = () => {
    console.log('WebSocket connection closed');
    isConnected.value = false;
    if (retryAttempt < maxRetries) {
      setTimeout(() => {
        retryAttempt++;
        console.log(`Reconnecting... Attempt ${retryAttempt}`);
        connectWebSocket(chatID, token);
      }, Math.min(1000 * Math.pow(2, retryAttempt), 30000)); // Exponential backoff
    } else {
      console.error('Max reconnection attempts reached.');
    }
  };
};

// Fetch chat history
const fetchChatHistory = async (chatID) => {
  try {
    const response = await axios.get(`/v1/chat/${chatID}/history`);
    if (response.data) {
      const newMessages = response.data.map(message => ({
        ...message,
        isOwnMessage: message.username === username,
        status: true,
      }));
      const existingMessageIds = new Set(messages.value.map(msg => msg.id));
      newMessages.forEach(newMessage => {
        if (!existingMessageIds.has(newMessage.id)) {
          messages.value.push(newMessage);
        }
      });
      // Ensure offline messages are included in the main messages array
      loadOfflineMessages();
    }
  } catch (error) {
    console.error('Error fetching chat history:', error);
  } finally {
    isLoading.value = false;
  }
};

// Lifecycle hooks
onMounted(async () => {
  const token = store.user.access;
  const chatID = route.params.chatID;
  if (chatID && token) {
    loadOfflineMessages();
    await fetchChatHistory(chatID);
    chatUser.value = route.query.user || 'Unknown';
    connectWebSocket(chatID, token);
  } else {
    isLoading.value = false;
  }
});

onUnmounted(() => {
  if (ws) ws.close();
  saveOfflineMessages();
});

// Send a message function
const sendMessage = () => {
  const message = {
    username,
    message: newMessage.value,
    timestamp: Date.now()
  };
  if (ws && isConnected.value) {
    ws.send(JSON.stringify(message));
    messages.value.push({ ...message, isOwnMessage: true, status: true });
    console.log('Message sent:', message);
  } else {
    messages.value.push({ ...message, isOwnMessage: true, status: false });
    saveOfflineMessages();
    console.log('Message saved as offline:', message);
  }
  newMessage.value = '';
};

// Watch for changes in connection status to update offline messages
watch(isConnected, (newVal) => {
  if (newVal) {
    const unsentMessages = messages.value.filter(msg => msg.status === false && msg.isOwnMessage);
    unsentMessages.forEach(message => {
      ws.send(JSON.stringify({ message: message.message, username }));
      message.status = true;
      console.log('Resent unsent message:', message);
    });
    removeOfflineMessages();
  }
});
</script>
