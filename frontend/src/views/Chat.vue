<template>
  <div>
    <h2>Chat with {{ chatUser }}</h2>
    <div v-if="formattedMessages.length === 0 && !isLoading">No messages yet</div>
    <div v-if="!isLoading">
      <div v-for="message in formattedMessages" :key="message.id"> <!-- Key bound to message.id -->
        <strong v-if="message.username && !message.isOwnMessage">
          <img :src="message.profile_picture" alt="Profile Picture" width="30" height="30" />
          {{ message.username }}
        </strong>
        {{ message.message }} - {{ formatTime(message.created_at) }}
        <span v-if="message.isOwnMessage">
          <span v-if="message.status === true">(Sent)</span>
          <span v-if="message.status === false">(Not Sent)</span>
          <button @click="deleteMessage(message.id)">Delete</button>
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

const formatTime = (timestamp) => {
  const timeDiff = Math.floor((Date.now() - new Date(timestamp)) / 1000);
  if (timeDiff < 60) return 'Just now';
  const minutes = Math.floor(timeDiff / 60);
  if (minutes < 60) return `${minutes} min ago`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours} ${hours === 1 ? 'hour' : 'hours'} ago`;
  const days = Math.floor(hours / 24);
  return `${days} ${days === 1 ? 'day' : 'days'} ago`;
};

const updateMessageTimes = () => {
  messages.value = messages.value.map(message => ({
    ...message,
    time: formatTime(message.created_at) 
  }));
};

const deleteMessage = async (messageID) => {
  try {
    console.log(`Attempting to delete message with ID: ${messageID}`);
    await axios.delete(`/v1/message/${messageID}`);
    messages.value = messages.value.filter(message => message.id !== messageID);
    console.log(`Message deleted with ID: ${messageID}`);
  } catch (err) {
    console.error("Error deleting message:", err);
  }
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
    }, Math.min(1000 * Math.pow(2, retryAttempt), 30000));
  } else {
    console.error('Max reconnection attempts reached.');
  }
};
ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  console.log('Received message:', message);

  if (message.type === 'MESSAGE_DELETED') {
    console.log('Message deleted:', message.message_id);

    // Ensure message is removed from the array
    const index = messages.value.findIndex(msg => msg.id == message.message_id);
    console.log('Index of message to be deleted:', index);

    if (index !== -1) {
      messages.value.splice(index, 1);
      console.log('Message deleted from array. Current messages:', messages.value);
    } else {
      console.error('Message ID not found in the array');
    }
  } else {
    if (message.id && message.username !== username) {
      if (!messages.value.some(msg => msg.id == message.id)) { 
        console.log('Adding new message to array:', message);
        messages.value.push({
          ...message,
          isOwnMessage: message.username === username,
          profile_picture: `/${message.profile_picture}`
        });
        console.log('Current messages array after addition:', messages.value);
      }
    }
  }
  if (route.params.chatID === chatID) {
    markChatNotificationsAsRead(chatID);
  }
};
};

// Fetch chat history
const fetchChatHistory = async (chatID) => {
  try {
    const response = await axios.get(`/v1/chat/${chatID}/history?user=${route.query.user}`);
    if (response.data) {
      const newMessages = response.data.map(message => {
        console.log('Profile Picture:', message.profile_picture); 
        return {
          ...message,
          isOwnMessage: message.username === username,
          status: true,
          time: message.created_at,
          profile_picture: `/${message.profile_picture}`
        };
      });
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


const markChatNotificationsAsRead = async (chatID) => {
  try {
    const token = store.user.access;
    await axios.post(`/v1/notifications/mark-read/${chatID}`, null, {
      headers: { Authorization: `Bearer ${token}` }
    });
    console.log(`Notifications for chat ${chatID} marked as read`);
  } catch (error) {
    console.error('Error marking notifications as read:', error);
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
    await markChatNotificationsAsRead(chatID);

    setInterval(updateMessageTimes, 60000);
  } else {
    isLoading.value = false;
  }
});

onUnmounted(() => {
  if (ws) ws.close();
  saveOfflineMessages();
});

// Send a message function
const sendMessage = async () => {
  const message = {
    username,
    message: newMessage.value,
    created_at: new Date().toISOString()
  };

  if (ws && isConnected.value) {
    ws.send(JSON.stringify(message));
    console.log('Message sent:', message);

    const receiveResponse = new Promise((resolve, reject) => {
      const responseHandler = (event) => {
        const response = JSON.parse(event.data);
        if (response.id) {
          ws.removeEventListener('message', responseHandler);
          resolve(response.id);
        } else {
          reject("Message ID not received");
        }
      };
      ws.addEventListener('message', responseHandler);
    });
    try {
      const messageID = await receiveResponse;
      console.log("Received message ID:", messageID);
      // Add the message with the real ID
      messages.value.push({ ...message, id: messageID, isOwnMessage: true, status: true });
    } catch (error) {
      console.error("Error receiving message ID:", error);
    }
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
