<template>
  <nav class="navbar">
    <div v-if="store.user.isAuthenticated">
      <router-link class="nav-link" to="/">Home</router-link>
      <router-link v-if="store.user.isAuthenticated" to="/inbox" class="nav-link">
        My Chats
        <span v-if="unreadCount > 0" class="badge">{{ unreadCount }}</span>
      </router-link>
      <router-link v-if="store.user.isAuthenticated" to="/notifications" class="nav-link">
        Notifications
        <span v-if="unreadNotificationCount > 0" class="badge">{{ unreadNotificationCount }}</span>
      </router-link>
      <router-link v-if="store.user.isAuthenticated" :to="`/${store.user.username}`" class="nav-link">
        My Profile
      </router-link>
      <button v-if="store.user.isAuthenticated" @click="logout" class="nav-button">Logout</button>
    </div>
    <div v-else>
      <router-link to="/login" class="nav-link">Login</router-link>
      <router-link to="/register" class="nav-link">Register</router-link>
    </div>
  </nav>
</template>

<script setup>
import { ref, onMounted, onUnmounted, watch } from 'vue';
import { useRouter } from 'vue-router';
import { userStore } from '@/stores/user';
import axios from 'axios';
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
    }
    if (data.unread_notification_count !== undefined) {
      unreadNotificationCount.value = data.unread_notification_count;
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

onMounted(() => {
  emitter.on('ws-open', handleWebSocketOpen);
  emitter.on('ws-message', handleWebSocketMessage);
  emitter.on('ws-error', handleWebSocketError);
  emitter.on('ws-close', handleWebSocketClose);
  emitter.on('notification-updated', fetchUnreadNotificationCount);
  emitter.on('chat-updated', fetchChatUnreadCount);
  emitter.on('chat-read', fetchChatUnreadCount);

  if (store.user.isAuthenticated) {
    fetchChatUnreadCount();
    fetchUnreadNotificationCount();
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
  emitter.off('chat-read', fetchChatUnreadCount);
});
</script>


<style scoped>
.navbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background-color: #333;
  padding: 10px;
}

.nav-link {
  color: white;
  text-decoration: none;
  margin: 0 10px;
}

.nav-link:hover {
  text-decoration: underline;
}

.nav-button {
  background-color: #ff4d4d;
  border: none;
  color: white;
  padding: 10px 20px;
  cursor: pointer;
  border-radius: 5px;
}

.nav-button:hover {
  background-color: #ff1a1a;
}

.badge {
  background-color: #ff4d4d;
  border-radius: 12px;
  padding: 2px 8px;
  color: white;
  font-size: 0.8em;
  margin-left: 5px;
}
</style>
