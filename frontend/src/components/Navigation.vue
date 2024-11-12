<template>
  <div>
    <div v-if="store.user.isAuthenticated">
      <router-link v-if="store.user.isAuthenticated" to="/inbox">My Chats
        <span v-if="unreadCount > 0">({{ unreadCount }})</span>
      </router-link> | 
      <router-link v-if="store.user.isAuthenticated" to="/notifications"> Notifications 
        <span v-if="unreadNotificationCount > 0">({{ unreadNotificationCount }})</span>
      </router-link> |
      <router-link v-if="store.user.isAuthenticated" :to="`/${store.user.username}`">My Profile</router-link> |
      <button v-if="store.user.isAuthenticated" @click="logout">Logout</button>
    </div>
    <div v-else>
      <router-link to="/login">Login</router-link>
      <router-link to="/register">Register</router-link>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, watch } from 'vue';
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

const fetchChatUnreadCount = async () => {
  if (store.user.isAuthenticated) {
    try {
      const response = await axios.get('/v1/notifications/chat/unread', {
        headers: {
          Authorization: `Bearer ${store.user.access}`
        }
      });
      unreadCount.value = response.data.unread_count;
      addData('chats', { chat_id: 'messageCounter', unread_count: response.data.unread_count });
    } catch (error) {
      console.error('Error fetching unread message count:', error);
    } finally {
      loader.value = false;
    }
  }
};

const fetchUnreadNotificationCount = async () => {
  if (store.user.isAuthenticated) {
    try {
      const response = await axios.get('/v1/notifications/count', {
        headers: {
          Authorization: `Bearer ${store.user.access}`
        }
      });
      unreadNotificationCount.value = response.data.unread_count;
      addData('notifications', { id: 'notificationCounter', unread_count: response.data.unread_count });
    } catch (error) {
      console.error('Error fetching unread notifications count:', error);
    } finally {
      loader.value = false;
    }
  }
};

const handleWebSocketOpen = () => {
  wsConnected.value = true;
  loader.value = false;
  fetchChatUnreadCount();
  fetchUnreadNotificationCount();
};

const handleWebSocketMessage = (data) => {
  if (store.user.isAuthenticated) {
    if (data.sender === store.user.username) return;

    if (data.total_unread_count !== undefined) {
      unreadCount.value = data.total_unread_count;
      addData('chats', { chat_id: 'messageCounter', total_unread_count: data.total_unread_count });
    }
    if (data.unread_notification_count !== undefined) {
      unreadNotificationCount.value = data.unread_notification_count;
      addData('notifications', { id: 'notificationCounter', total_unread_count: data.unread_notification_count });
    }
  }
};

const handleWebSocketError = (error) => {
  console.error("WebSocket error in navigation.vue:", error);
};

const handleWebSocketClose = () => {
  console.log("WebSocket connection closed in navigation.vue");
  wsConnected.value = false;
};

const logout = () => {
  store.removeToken();
  router.push('/login');
};

onMounted(async () => {
  emitter.on('ws-open', handleWebSocketOpen);
  emitter.on('ws-message', handleWebSocketMessage);
  emitter.on('ws-error', handleWebSocketError);
  emitter.on('ws-close', handleWebSocketClose);
  emitter.on('notification-updated', fetchUnreadNotificationCount);
  emitter.on('chat-updated', fetchChatUnreadCount);
  emitter.on('chat-read', fetchChatUnreadCount);
  
  if (store.user.isAuthenticated) {
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

// Watch for authentication changes and fetch data
watch(
  () => store.user.isAuthenticated,
  async (newVal) => {
    if (newVal) {
      await fetchChatUnreadCount();
      await fetchUnreadNotificationCount();
    }
  }
);

onUnmounted(() => {
  emitter.off('ws-open', handleWebSocketOpen);
  emitter.off('ws-message', handleWebSocketMessage);
  emitter.off('ws-error', handleWebSocketError);
  emitter.off('ws-close', handleWebSocketClose);
  emitter.off('notification-updated', fetchUnreadNotificationCount);
  emitter.off('chat-updated', fetchChatUnreadCount);
  emitter.on('chat-read', fetchChatUnreadCount);
});
</script>