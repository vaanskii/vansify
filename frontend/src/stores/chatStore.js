import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import axios from 'axios';
import emitter from '@/eventBus';

export const useChatStore = defineStore('chatStore', () => {
  const chats = ref([]);
  const error = ref('');
  const wsConnected = ref(false);
  const loader = ref(true);

  const fetchChats = async () => {
    try {
      const response = await axios.get('/v1/me/chats');
      if (response.data && response.data.chats) {
        const newChats = response.data.chats.map(chat => ({
          chat_id: chat.chat_id,
          user: chat.user,
          unread_count: chat.unread_count,
          last_message_time: chat.last_message_time,
          profile_picture: chat.profile_picture
        }));
        chats.value = newChats;
        loader.value = false;
      } else {
        chats.value = [];
      }
    } catch (err) {
      error.value = err.response ? err.response.data.error : 'An error occurred';
      console.error("Error fetching chats:", err);
    }
  };

  const deleteChat = async (chatID) => {
    try {
      await axios.delete(`/v1/chat/${chatID}`);
      chats.value = chats.value.filter(chat => chat.chat_id !== chatID);
    } catch (err) {
      error.value = err.response ? err.response.data.error : 'An error occurred';
    }
  };

  const sortedChats = computed(() => {
    return chats.value.slice().sort((a, b) => new Date(b.last_message_time) - new Date(a.last_message_time));
  });

  const formatTime = (timestamp) => {
    const timeDiff = Math.floor((Date.now() - new Date(timestamp)) / 1000);
    if (timeDiff < 60) return 'Just now';
    const minutes = Math.floor(timeDiff / 60);
    if (minutes < 60) return `${minutes} min ago`;
    const hours = Math.floor(minutes / 60);
    if (hours < 24) return `${hours} ${hours === 1 ? 'hour' : 'hours'} ago`;
    const days = Math.floor(hours / 24);
    return `${days} ${days === 1 ? 'day' : 'days'} ago`;
  };

  const updateMessageTimes = () => {
    chats.value = chats.value.map(chat => ({
      ...chat,
      time: formatTime(chat.last_message_time) 
    }));
  };

  const handleWebSocketOpen = () => {
    console.log("WebSocket connection opened in ChatListView");
    wsConnected.value = true;
    loader.value = false;
  };

  const handleWebSocketMessage = (data) => {
    try {
      const chatIndex = chats.value.findIndex(chat => chat.chat_id === data.chat_id);
      if (chatIndex !== -1) {
        chats.value[chatIndex] = {
          ...chats.value[chatIndex],
          unread_count: data.unread_count,
          last_message_time: data.last_message_time || new Date().toISOString(),
          message: data.message,
          user: data.user,
          profile_picture: data.profile_picture
        };
      } else {
        chats.value.push({
          chat_id: data.chat_id,
          user: data.user,
          unread_count: data.unread_count,
          last_message_time: data.last_message_time || new Date().toISOString(),
          message: data.message,
          profile_picture: data.profile_picture
        });
      }
      chats.value = chats.value.slice().sort((a, b) => new Date(b.last_message_time) - new Date(a.last_message_time));
    } catch (e) {
      console.error("Error processing WebSocket message:", e);
    }
  };

  const markChatAsRead = async (chatID) => {
    try {
      await axios.post(`/v1/notifications/chat/mark-read/${chatID}`);
      chats.value = chats.value.map(chat =>
        chat.chat_id === chatID ? { ...chat, unread_count: 0 } : chat
      );
      emitter.emit('chat-updated');
    } catch (error) {
      console.error('Error marking chat as read:', error);
    }
  };

  const handleWebSocketError = (error) => {
    console.error("Notification WebSocket error: ", error);
  };

  const handleWebSocketClose = () => {
    console.log("WebSocket connection closed");
    wsConnected.value = false;
  };

  return {
    chats,
    error,
    wsConnected,
    loader,
    fetchChats,
    deleteChat,
    sortedChats,
    formatTime,
    updateMessageTimes,
    handleWebSocketOpen,
    handleWebSocketMessage,
    markChatAsRead,
    handleWebSocketError,
    handleWebSocketClose
  };
});
