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
import emitter from '@/eventBus';

const store = userStore();
const router = useRouter();
const unreadCount = ref(0);
const unreadNotificationCount = ref(0);
const wsConnected = ref(false);
const loader = ref(true);

const handleWebSocketOpen = () => {
  console.log("WebSocket connection opened for general notifications");
  wsConnected.value = true;
  loader.value = false;
  fetchUnreadCount();
  fetchUnreadNotificationCount();
};

const handleWebSocketMessage = (data) => {
  if (data.unread_count !== undefined) {
    unreadCount.value = data.unread_count;
    addData('chats', { chat_id: 'messageCounter', unread_count: data.unread_count });
  }
  if (data.unread_notification_count !== undefined) {
    unreadNotificationCount.value = data.unread_notification_count;
    addData('notifications', { id: 'notificationCounter', unread_count: data.unread_notification_count });
  }
};

const handleWebSocketError = (error) => {
  console.error("Notification WebSocket error: ", error);
};

const handleWebSocketClose = () => {
  console.log("WebSocket connection closed for general notifications.");
  wsConnected.value = false;
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

const notifications = () => {
  router.push('/notifications');
};

const goToLogin = () => {
  router.push('/login');
};

const goToRegister = () => {
  router.push('/register');
};

onMounted(async () => {
  if (store.user.isAuthenticated) {
    emitter.on('ws-open', handleWebSocketOpen);
    emitter.on('ws-message', handleWebSocketMessage);
    emitter.on('ws-error', handleWebSocketError);
    emitter.on('ws-close', handleWebSocketClose);

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
  emitter.off('ws-open', handleWebSocketOpen);
  emitter.off('ws-message', handleWebSocketMessage);
  emitter.off('ws-error', handleWebSocketError);
  emitter.off('ws-close', handleWebSocketClose);
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
