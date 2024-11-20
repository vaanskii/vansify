<template>
  <navigation />
  <router-view />
</template>

<script setup>
import Navigation from './components/Navigation.vue';
import { onMounted, onUnmounted, ref, watch } from 'vue';
import { userStore } from '@/stores/user';
import emitter from '@/eventBus';
import notify from '@/utils/notify';

const store = userStore();
const ws = ref(null);
const apiUrl = import.meta.env.VITE_WS_URL;
const wsConst = import.meta.env.VITE_WS;
let reconnectAttempts = 0;
const maxReconnectAttempts = 10;
let initialConnection = true;
let reconnecting = false;

const reconnectWebSocket = () => {
  if (reconnectAttempts < maxReconnectAttempts) {
    reconnectAttempts++;
    if (!reconnecting) {
      notify('Reconnecting...', 'info', 3000);
      reconnecting = true;
    }
    setTimeout(connectWebSocket, Math.min(1000 * reconnectAttempts, 30000));
  } else {
    notify('Something went wrong, please try again later.', 'error', 5000);
  }
};

// Function to establish WebSocket connection
const connectWebSocket = () => {
  if (store.user.isAuthenticated && !ws.value) {
    const wsUrl = `${wsConst}//${apiUrl}/v1/notifications/ws?token=${encodeURIComponent(store.user.access)}`;
    ws.value = new WebSocket(wsUrl);

    // Handle WebSocket events
    ws.value.onopen = () => {
      reconnectAttempts = 0;
      reconnecting = false;
      if (import.meta.env.MODE === 'development') {
        console.log('WebSocket connection established in app.vue');
      }
      if (!initialConnection) {
        notify('Connected.', 'success', 3000);
      }
      initialConnection = false;
      emitter.emit('ws-open');
    };

    ws.value.onmessage = (event) => {
      const data = JSON.parse(event.data);
      emitter.emit('ws-message', data);
    };

    ws.value.onerror = (error) => {
      if (import.meta.env.MODE === 'development') {
        console.error('WebSocket error:', error);
      }
      emitter.emit('ws-error', error);
      if (!reconnecting && store.user.isAuthenticated) {
        notify('Something went wrong, Please try again later...', 'error', 3000);
        reconnecting = true;
      }
    };

    ws.value.onclose = () => {
      if (import.meta.env.MODE === 'development') {
        console.log('WebSocket connection closed in app.vue');
      }
      if (store.user.isAuthenticated) {
        ws.value = null;
        reconnectWebSocket();
        if (!reconnecting) {
          notify('WebSocket connection closed. Attempting to reconnect...', 'info', 3000);
          reconnecting = true;
        }
      } else {
        reconnecting = false;
      }
      emitter.emit('ws-close');
    };
  } else if (!store.user.isAuthenticated && ws.value) {
    ws.value.close();
    ws.value = null;
  }
};

onMounted(() => {
  store.initStore();
  connectWebSocket();
});

onUnmounted(() => {
  if (ws.value) {
    ws.value.close();
  }
});

// Watch for changes in authentication state
watch(
  () => store.user.isAuthenticated,
  (newVal) => {
    if (import.meta.env.MODE === 'development') {
      console.log('Authentication state changed:', newVal);
    }
    if (!newVal) {
      initialConnection = true;
    }
    connectWebSocket();
  }
);
</script>
