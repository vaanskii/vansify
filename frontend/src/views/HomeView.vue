<template>
  <div>
    <button v-if="store.user.isAuthenticated" @click="logout">Logout</button>
    <button v-if="store.user.isAuthenticated" @click="goToChats">
      My Chats
      <span v-if="unreadCount > 0">({{ unreadCount }})</span> <!-- Show total unread count -->
    </button>
    <button v-if="store.user.isAuthenticated" @click="goToProfile">My Profile</button>
    <div v-else>
      <button @click="goToLogin">Login</button>
      <button @click="goToRegister">Register</button>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue';
import { useRouter } from 'vue-router';
import { userStore } from '@/stores/user';
import axios from 'axios';

const apiUrl = import.meta.env.VITE_WS_URL;
const store = userStore();
const router = useRouter();
const unreadCount = ref(0);
const wsUrl = `ws://${apiUrl}/v1/notifications/ws?token=${encodeURIComponent(store.user.access)}`;
let ws;

const connectNotificationWebSocket = () => {
  ws = new WebSocket(wsUrl);
  ws.onopen = () => {
    console.log("Notification WebSocket connection established");
  };
  ws.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data);
      if (data.unread_count !== undefined) {
        unreadCount.value = data.unread_count;
      }
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

const logout = () => {
  store.removeToken();
  router.push('/login');
};

const goToChats = () => {
  router.push('/inbox');
};

const goToProfile = () => {
  router.push(`/${store.user.username}`);
};

const goToLogin = () => {
  router.push('/login');
};

const goToRegister = () => {
  router.push('/register');
};

onMounted(() => {
  if (store.user.isAuthenticated) {
    fetchUnreadCount();
    connectNotificationWebSocket();
  }
});

onUnmounted(() => {
  if (ws) ws.close();
});

const fetchUnreadCount = async () => {
  try {
    const token = store.user.access;
    const response = await axios.get('/v1/notifications/unread', {
      headers: { Authorization: `Bearer ${token}` }
    });
    unreadCount.value = response.data.unread_count;
  } catch (error) {
    console.error('Error fetching unread message count:', error);
  }
};
</script>
