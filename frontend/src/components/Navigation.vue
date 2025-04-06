<template>
  <div>
    <nav v-if="store.user.isAuthenticated && !isChatRoute" class="navbar">
      <!-- <router-link class="nav-link" to="/">
        <i class="fa-solid fa-house fa-lg"></i>
      </router-link> -->
      <router-link to="/inbox" class="nav-link">
        <i v-if="unreadCount > 0" class="fa-solid fa-comment fa-lg"></i>
        <i v-else class="fa-regular fa-comment fa-lg"></i>
        <span v-if="unreadCount > 0" class="badge">{{ unreadCount }}</span>
      </router-link>
      <router-link to="/notifications" class="nav-link">
        <i v-if="unreadNotificationCount > 0" class="fa-solid fa-bell fa-lg"></i>
        <i v-else class="fa-regular fa-bell fa-lg"></i>
        <span v-if="unreadNotificationCount > 0" class="badge">{{ unreadNotificationCount }}</span>
      </router-link>
      <router-link :to="`/${store.user.username}`" class="nav-link">
        <i class="fa-solid fa-user fa-lg"></i>
        <!-- {{ store.user.username }} -->
      </router-link>
      <div @click="focusInput" 
          class="relative cursor-pointer w-9 flex items-center border-transparent 
                  rounded-xl border-2 transition-all duration-500 focus-within:w-64 
                  focus-within:bg-white">
        <i
          class="fa-solid fa-magnifying-glass text-lg px-2 fa-lg transition-all duration-500"
          :class="{'text-white': !isFocused, 'text-black': isFocused}"></i>
        <input
          ref="searchInput"
          v-model="searchQuery" @keydown.enter="handleSearch"
          @focus="handleFocus"
          @blur="handleBlur"
          type="text" autocomplete="off"
          class="bg-transparent border-none outline-none w-full opacity-0 
                pointer-events-none transition-all duration-500 focus:opacity-100 
                focus:pointer-events-auto focus:pl-2 p-0.5">
      </div>
      <button @click="logout" class="nav-button ml-4">
        <i class="fa-solid fa-arrow-right-from-bracket"></i>
      </button>
    </nav>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { userStore } from '@/stores/user';
import axios from 'axios';
import emitter from '@/eventBus';

const store = userStore();
const router = useRouter();
const route = useRoute();
const unreadCount = ref(0);
const unreadNotificationCount = ref(0);
const wsConnected = ref(false);
const loader = ref(true);


const searchQuery = ref("");
const searchInput = ref(null);
const isFocused = ref(false);

const focusInput = () => {
  searchInput.value?.focus();
};

const handleFocus = () => {
  isFocused.value = true;
};

const handleBlur = () => {
  isFocused.value = false;
};

const handleSearch = () => {
  if (searchQuery.value.trim()) {
    router.push(`/search?query=${searchQuery.value}`);
  }
};

const isMobile = ref(window.innerWidth <= 768);
const isChatRoute = computed(() => {
  return route.path.startsWith('/inbox/') && isMobile.value;
});

const updateIsMobile = () => {
  isMobile.value = window.innerWidth <= 768;
};

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

const logout = async () => {
  try {
    await axios.post('/v1/logout', {},);
    console.log('User logged out successfully');
  } catch (error) {
    console.error('Error logging out:', error);
  }
  store.removeToken();
  router.push('/');
};


onMounted(() => {
  emitter.on('chat-ws-open', handleWebSocketOpen);
  emitter.on('chat-ws-message', handleWebSocketMessage);
  emitter.on('chat-ws-error', handleWebSocketError);
  emitter.on('chat-ws-close', handleWebSocketClose);

  emitter.on('global-ws-open', handleWebSocketOpen);
  emitter.on('global-ws-message', handleWebSocketMessage);
  emitter.on('global-ws-error', handleWebSocketError);
  emitter.on('global-ws-close', handleWebSocketClose);

  emitter.on('notification-updated', fetchUnreadNotificationCount);
  emitter.on('chat-updated', fetchChatUnreadCount);
  emitter.on('chat-read', fetchChatUnreadCount);

  window.addEventListener('resize', updateIsMobile);
  if (store.user.isAuthenticated) {
    fetchChatUnreadCount();
    fetchUnreadNotificationCount();
  }
  if(route.query.query) {
    searchQuery.value = route.query.query;
  }
});

watch(
  () => store.user.isAuthenticated,
  async (newVal) => {
    if (newVal) {
      await fetchChatUnreadCount();
      await fetchUnreadNotificationCount();
    }
  }
);

watch(() => route.query.query, (newQuery) => {
  searchQuery.value = newQuery || "";
});

watch(route, (newRoute) => {
  if (isMobile.value && newRoute.path.startsWith('/inbox/')) {
    // Hide navigation for chat view on mobile
    console.log('Navigated to chat view on mobile');
  } else {
    // Show navigation for other views
    console.log('Navigated away from chat view');
  }
});

onUnmounted(() => {
  emitter.off('chat-ws-open', handleWebSocketOpen);
  emitter.off('chat-ws-message', handleWebSocketMessage);
  emitter.off('chat-ws-error', handleWebSocketError);
  emitter.off('chat-ws-close', handleWebSocketClose);

  emitter.off('global-ws-open', handleWebSocketOpen);
  emitter.off('global-ws-message', handleWebSocketMessage);
  emitter.off('global-ws-error', handleWebSocketError);
  emitter.off('global-ws-close', handleWebSocketClose);

  emitter.off('notification-updated', fetchUnreadNotificationCount);
  emitter.off('chat-updated', fetchChatUnreadCount);
  emitter.off('chat-read', fetchChatUnreadCount);

  window.removeEventListener('resize', updateIsMobile);
});
</script>

<style scoped>
.navbar {
  display: flex;
  justify-content: center;
  align-items: center;
  background-color: #333;
  padding: 10px;
  position: -webkit-sticky;
  position: sticky;
  top: 0;
  z-index: 10;
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

.nav-link {
  position: relative;
  color: white;
  text-decoration: none;
  margin: 0 10px;
}

.nav-link.router-link-active i {
  color: red;
}

.nav-link:hover {
  text-decoration: underline;
}

.badge {
  position: absolute;
  top: -10px;
  right: -8px;
  background-color: #ff4d4d;
  border-radius: 12px;
  padding: 2px 6px;
  color: white;
  font-size: 0.5em;
}
</style>
