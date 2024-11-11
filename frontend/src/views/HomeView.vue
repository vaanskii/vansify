<template>
  <div>
    <button v-if="store.user.isAuthenticated" @click="logout">Logout</button>
    <button v-if="store.user.isAuthenticated" @click="goToChats">
      My Chats
      <span v-if="unreadCount > 0">({{ unreadCount }})</span>
    </button>
    <button v-if="store.user.isAuthenticated" @click="notifications">
      Notifications
      <span v-if="unreadNotificationCount > 0">({{ unreadNotificationCount }})</span>
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
import { addData, getData } from '@/utils/notifDB';

const apiUrl = import.meta.env.VITE_WS_URL;
const store = userStore();
const router = useRouter();
const unreadCount = ref(0);
const unreadNotificationCount = ref(0);
const wsUrl = `ws://${apiUrl}/v1/notifications/ws?token=${encodeURIComponent(store.user.access)}`;
let ws;
const wsConnected = ref(false);
const loader = ref(true);

const connectNotificationWebSocket = () => {
  ws = new WebSocket(wsUrl);
  ws.onopen = () => {
    if (import.meta.env.MODE === 'development') { 
      console.log("Notification WebSocket connection established");
    }
    wsConnected.value = true;
    loader.value = false;
    fetchUnreadCount();
    fetchUnreadNotificationCount();
  };
  ws.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data);
      if (data.unread_count !== undefined) {
        unreadCount.value = data.unread_count;
        addData('chats', { chat_id: 'messageCounter', unread_count: data.unread_count });
      }
      if (data.unread_notification_count !== undefined) {
        unreadNotificationCount.value = data.unread_notification_count;
        addData('notifications', { id: 'notificationCounter', unread_count: data.unread_notification_count });
      }
    } catch (e) {
      console.error("Error processing WebSocket message:", e);
    }
  };
  ws.onerror = (error) => {
    if (import.meta.env.MODE === 'development') { 
      console.error("Notification WebSocket error: ", error); 
    }
  };
  ws.onclose = () => {
    if (import.meta.env.MODE === 'development') { 
      console.log("Notification WebSocket connection closed");
    }
    wsConnected.value = false; 
  };
};

const logout = () => {
  store.removeToken();
  router.push('/login');
};

const goToChats = () => {
  if (wsConnected.value) {
    router.push('/inbox');
  } else {
    const interval = setInterval(() => {
      if (wsConnected.value) {
        clearInterval(interval);
        router.push('/inbox');
      }
    }, 100);
  }
};

const goToProfile = () => {
  if (wsConnected.value) {
    router.push(`/${store.user.username}`);
  } else {
    const interval = setInterval(() => {
      if (wsConnected.value) {
        clearInterval(interval);
        router.push(`/${store.user.username}`);
      }
    }, 100);
  }
};

const notifications = () => {
  if (wsConnected.value) {
    router.push('/notifications');
  } else {
    const interval = setInterval(() => {
      if (wsConnected.value) {
        clearInterval(interval);
        router.push('/notifications');
      }
    }, 100);
  }
};

const goToLogin = () => {
  router.push('/login');
};

const goToRegister = () => {
  router.push('/register');
};

onMounted(async () => {
  if (store.user.isAuthenticated) {
    connectNotificationWebSocket();
    const chatData = await getData('chats', 'messageCounter');
    if (chatData) {
      unreadCount.value = chatData.unread_count;
    }
    const notificationData = await getData('notifications', 'notificationCounter');
    if (notificationData) {
      unreadNotificationCount.value = notificationData.unread_count;
    }
  }
});

onUnmounted(() => {
  if (ws) ws.close();
});

const fetchUnreadCount = async () => {
  try {
    const response = await axios.get('/v1/notifications/chat/unread');
    unreadCount.value = response.data.unread_count;
    addData('chats', { chat_id: 'messageCounter', unread_count: response.data.unread_count });
  } catch (error) {
    console.error('Error fetching unread message count:', error);
  } finally {
    loader.value = false;
  }
};

const fetchUnreadNotificationCount = async () => {
  try {
    const response = await axios.get('/v1/notifications/count');
    unreadNotificationCount.value = response.data.unread_count;
    addData('notifications', { id: 'notificationCounter', unread_count: response.data.unread_count });
  } catch (error) {
    console.error('Error fetching unread notifications count:', error);
  } finally {
    loader.value = false;
  }
};
</script>
