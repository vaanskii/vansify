import { defineStore } from 'pinia';
import { ref } from 'vue';
import emitter from '@/eventBus';
import { userStore } from '@/stores/user';
import notify from '@/utils/notify';

export const useActiveUsersStore = defineStore('activeUsers', () => {
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
      const wsUrl = `${wsConst}//${apiUrl}/v1/active-users/ws?username=${encodeURIComponent(store.user.username)}`;
      ws.value = new WebSocket(wsUrl);

      ws.value.onopen = () => {
        reconnectAttempts = 0;
        reconnecting = false;
        if (import.meta.env.MODE === 'development') {
          console.log('WebSocket connection established for active users');
        }
        if (!initialConnection) {
          notify('Connected.', 'success', 3000);
        }
        initialConnection = false;
        emitter.emit('active-users-ws-open');
      };

      ws.value.onmessage = (event) => {
        const data = JSON.parse(event.data);
        emitter.emit('active-users-fetched', data);
      };

      ws.value.onerror = (error) => {
        if (import.meta.env.MODE === 'development') {
          console.error('WebSocket error for active users:', error);
        }
        emitter.emit('active-users-ws-error', error);
        if (!reconnecting) {
          notify('Something went wrong, please try again later...', 'error', 3000);
          reconnecting = true;
        }
      };

      ws.value.onclose = () => {
        if (import.meta.env.MODE === 'development') {
          console.log('WebSocket connection closed for active users');
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
        emitter.emit('active-users-ws-close');
      };
    }
  };

  return {
    connectWebSocket,
    ws,
  };
});
