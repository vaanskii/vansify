<template>
  <router-view />
</template>

<script setup>
import { onMounted, onUnmounted, ref } from 'vue';
import { userStore } from '@/stores/user';
import emitter from '@/eventBus';

const store = userStore();
const ws = ref(null);
const apiUrl = import.meta.env.VITE_WS_URL;

onMounted(() => {
  store.initStore();

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
    emitter.emit('ws-close');
  };
});

onUnmounted(() => {
  if (ws.value) {
    ws.value.close();
  }
});
</script>
