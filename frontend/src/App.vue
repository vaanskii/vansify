<template>
  <navigation/>
  <router-view />
</template>

<script setup>
import Navigation from './components/Navigation.vue';
import { onMounted, onUnmounted, ref, watch } from 'vue';
import { userStore } from '@/stores/user';
import emitter from '@/eventBus';

const store = userStore();
const ws = ref(null);
const apiUrl = import.meta.env.VITE_WS_URL;

// Function to establish WebSocket connection
const connectWebSocket = () => {
  if (store.user.isAuthenticated && !ws.value) {
    const wsUrl = `ws://${apiUrl}/v1/notifications/ws?token=${encodeURIComponent(store.user.access)}`;
    ws.value = new WebSocket(wsUrl);

    // Handle WebSocket events
    ws.value.onopen = () => {
      console.log('WebSocket connection established in app.vue');
      emitter.emit('ws-open');
    };

    ws.value.onmessage = (event) => {
      const data = JSON.parse(event.data);
      emitter.emit('ws-message', data);
    };

    ws.value.onerror = (error) => {
      console.error('WebSocket error:', error);
      emitter.emit('ws-error', error);
    };

    ws.value.onclose = () => {
      console.log('WebSocket connection closed in app.vue');
      ws.value = null;
      emitter.emit('ws-close');
    };
  } else if (!store.user.isAuthenticated && ws.value) {
    // Close the WebSocket if the user logs out
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
watch(() => store.user.isAuthenticated, (newVal) => {
  console.log('Authentication state changed:', newVal);
  connectWebSocket();
});
</script>
