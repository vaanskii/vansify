//chatNotification.js
import { defineStore } from 'pinia';
import { ref } from 'vue';
import emitter from '@/eventBus';
import notify from '@/utils/notify';
import { userStore } from '@/stores/user';

export const useChatNotificationStore = defineStore('chatNotifications', () => {
  const ws = ref(null);
  const apiUrl = import.meta.env.VITE_WS_URL;
  const wsConst = import.meta.env.VITE_WS;
  const store = userStore();
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

  const connectWebSocket = () => {
    if (store.user.isAuthenticated && !ws.value) {
      const wsUrl = `${wsConst}//${apiUrl}/v1/chat-notifications/ws?token=${encodeURIComponent(store.user.access)}`;
      ws.value = new WebSocket(wsUrl);

      ws.value.onopen = () => {
        reconnectAttempts = 0;
        reconnecting = false;
        if (import.meta.env.MODE === 'development') {
          console.log('WebSocket connection established in chat notifications store');
        }
        if (!initialConnection) {
          notify('Connected.', 'success', 3000);
        }
        initialConnection = false;
        emitter.emit('chat-ws-open');
      };

      ws.value.onmessage = (event) => {
        const data = JSON.parse(event.data);
        console.log("Received chat notification data", event.data);
        emitter.emit('chat-ws-message', data);
      };

      ws.value.onerror = (error) => {
        if (import.meta.env.MODE === 'development') {
          console.error('WebSocket error in chat notifications store:', error);
        }
        emitter.emit('chat-ws-error', error);
        if (!reconnecting && store.user.isAuthenticated) {
          notify('Something went wrong, please try again later...', 'error', 3000);
          reconnecting = true;
        }
      };

      ws.value.onclose = () => {
        if (import.meta.env.MODE === 'development') {
          console.log('WebSocket connection closed in chat notifications store');
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
        emitter.emit('chat-ws-close');
      };
    }
  };

  return {
    connectWebSocket,
    ws,
  };
});
