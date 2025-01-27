<template>
  <div class="layout">
    <navigation class="navigation" />
    <router-view />
  </div>
</template>

<script setup>
import Navigation from './components/Navigation.vue';
import { onMounted, onUnmounted, watch } from 'vue';
import { userStore } from '@/stores/user';
import { useNotificationStore } from '@/stores/appNotifications';
import { useChatNotificationStore } from './stores/chatNotification';
import { useActiveUsersStore } from '@/stores/activeUsers';

const store = userStore();
const notificationStore = useNotificationStore();
const activeUsersStore = useActiveUsersStore();
const chatNotificationStore = useChatNotificationStore();

const closeWebSockets = () => {
  if (notificationStore.ws.value) {
    notificationStore.ws.value.close();
  }
  if (activeUsersStore.ws.value) {
    activeUsersStore.ws.value.close();
  }
  if (chatNotificationStore.ws.value) {
    chatNotificationStore.ws.value.close();
  }
};

onMounted(() => {
  store.initStore();
  if (store.user.isAuthenticated) {
    notificationStore.connectWebSocket();
    activeUsersStore.connectWebSocket();
    chatNotificationStore.connectWebSocket();
  }
});

onUnmounted(() => {
  closeWebSockets();
});

watch(
  () => store.user.isAuthenticated,
  (newVal) => {
    if (import.meta.env.MODE === 'development') {
      console.log('Authentication state changed:', newVal);
    }
    if (newVal) {
      notificationStore.connectWebSocket();
      activeUsersStore.connectWebSocket();
      chatNotificationStore.connectWebSocket();
    } else {
      closeWebSockets();
    }
  }
);
</script>

<style>
.layout {
  display: flex;
  flex-direction: column;
  height: 100dvh;
}

.navigation {
  width: 100%;
}

@media (min-width: 769px) {
  .navigation {
    position: relative;
    flex-shrink: 0;
  }
}

@media (max-width: 768px) {
  .navigation {
    position: fixed;
    bottom: 0;
    width: 100%;
    z-index: 10;
  }
}

.router-view {
  flex-grow: 1;
  overflow-y: auto;
  padding-bottom: 60px;
}
</style>