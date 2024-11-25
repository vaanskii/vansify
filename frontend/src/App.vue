<template>
  <navigation />
  <router-view />
</template>

<script setup>
import Navigation from './components/Navigation.vue';
import { onMounted, onUnmounted, watch } from 'vue';
import { userStore } from '@/stores/user';
import { useNotificationStore } from '@/stores/appNotifications';
import { useActiveUsersStore } from '@/stores/activeUsers';

const store = userStore();
const notificationStore = useNotificationStore();
const activeUsersStore = useActiveUsersStore();

const closeWebSockets = () => {
  if (notificationStore.ws.value) {
    notificationStore.ws.value.close();
  }
  if (activeUsersStore.ws.value) {
    activeUsersStore.ws.value.close();
  }
};

onMounted(() => {
  store.initStore();
  if (store.user.isAuthenticated) {
    notificationStore.connectWebSocket();
    activeUsersStore.connectWebSocket();
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
    } else {
      closeWebSockets();
    }
  }
);
</script>
