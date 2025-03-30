import { defineStore } from 'pinia';
import { ref, computed} from 'vue';
import axios from 'axios';
import emitter from '@/eventBus';
import { userStore } from './user';
import { useRouter } from 'vue-router';
import { parseISO, formatDistanceToNow, differenceInMinutes, differenceInHours, differenceInDays } from 'date-fns';

export const useChatStore = defineStore('chatStore', () => {
  const chats = ref([]);
  const error = ref('');
  const wsConnected = ref(false);
  const loader = ref(true);
  const store = userStore();
  const activeUsers = ref([]);
  const router = useRouter();
  const fetchChats = async () => {
    try {
        const response = await axios.get('/v1/me/chats', {
            headers: {
                Authorization: `Bearer ${store.user.access}`
            }
        });

        if (response.data && response.data.chats) {
            console.log("chats", response.data.chats);

            const newChats = response.data.chats
                .filter(chat => 
                    (!chat.deleted_for || !chat.deleted_for.includes(store.user.username)) && chat.last_message)
                .map(chat => ({
                    chat_id: chat.chat_id,
                    user: chat.user,
                    unread_count: chat.unread_count,
                    last_message_time: chat.last_message_time || "",
                    profile_picture: chat.profile_picture,
                    last_message: chat.last_message
                }));

            chats.value = newChats.length > 0 ? newChats : [];
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
      await axios.delete(`/v1/chat/${chatID}`, {
        headers: { 
          Authorization: `Bearer ${store.user.access}` 
        }
      });
      chats.value = chats.value.filter(chat => chat.chat_id !== chatID);
      router.push("/inbox")
    } catch (err) {
      error.value = err.response ? err.response.data.error : 'An error occurred';
    }
  };

  const deleteMessagesForUser = async (chatID) => {
    console.log(`Attempting to delete messages for chatID: ${chatID}`);
    try {
      await axios.delete(`/v1/chat/${chatID}/delete-messages`, {
        headers: { 
          Authorization: `Bearer ${store.user.access}` 
        }
      });
      console.log("Messages deleted successfully on the backend");
  
      // Remove the chat from the list immediately
      chats.value = chats.value.filter(chat => chat.chat_id !== chatID);
      router.push("/inbox")
      console.log("Updated chats:", chats.value);
    } catch (err) {
      console.error('Error:', err.response ? err.response.data.error : 'An error occurred');
    }
  };
  

  const sortedChats = computed(() => {
    return chats.value.slice().sort((a, b) => new Date(b.last_message_time) - new Date(a.last_message_time));
  });


  function formatTimeAgo(utcTime) {
    const localTime = parseISO(utcTime);
  
    const minutesDiff = differenceInMinutes(new Date(), localTime);
    if (minutesDiff < 1) {
      return '1m';
    }
  
    if (minutesDiff < 60) {
      return `${minutesDiff}m`;
    }

    const hoursDiff = differenceInHours(new Date(), localTime);
    if (hoursDiff < 24) {
      return `${hoursDiff}h`;
    }
  
    const daysDiff = differenceInDays(new Date(), localTime);
    if (daysDiff >= 1) {
      return `${daysDiff}d`;
    }
  
    return formatDistanceToNow(localTime, { addSuffix: true });
  }

  const updateMessageTimes = () => {
    chats.value = chats.value.map(chat => ({
      ...chat,
      time: formatTimeAgo(chat.last_message_time) 
    }));
  };

  const handleActiveUsersFetched = (activeUsersData) => {
    activeUsers.value = activeUsersData; 
  };

  const handleWebSocketOpen = () => {
    console.log("WebSocket connection opened");
    wsConnected.value = true;
    loader.value = false;
  };
  

  const handleWebSocketMessage = (data) => {
    try {
      console.log("WebSocket message received:", data);
  
      const chatIndex = chats.value.findIndex(chat => chat.chat_id === data.chat_id);
      console.log("Chat index found:", chatIndex);
  
      if (chatIndex !== -1) {
        console.log("Updating existing chat:", chats.value[chatIndex]);
  
        // Determine whether the current user is the sender or the recipient
        const isSender = data.sender === store.user.username;
  
        const updatedChat = {
          ...chats.value[chatIndex],
          last_message_time: data.last_message_time,
          last_message: data.last_message,
          user: isSender ? data.receiver : data.sender, // Update username based on role
          profile_picture: isSender ? data.receiver_profile_picture : data.sender_profile_picture, // Update profile picture based on role
        };
  
        if (!isSender) {
          // If the message is from the recipient, update unread count
          updatedChat.unread_count = data.unread_count || chats.value[chatIndex].unread_count;
        }
  
        chats.value[chatIndex] = updatedChat;
        console.log("Updated chat:", chats.value[chatIndex]);
      } else {
        console.log("Adding new chat:", data);
  
        // Determine whether the current user is the sender or the recipient
        const isSender = data.sender === store.user.username;
  
        chats.value.push({
          chat_id: data.chat_id,
          user: isSender ? data.recipient : data.sender, // Add the appropriate username
          unread_count: data.unread_count || 0,
          last_message_time: data.last_message_time,
          last_message: data.last_message,
          profile_picture: isSender ? data.receiver_profile_picture : data.sender_profile_picture, // Add the appropriate profile picture
        });
        console.log("New chat added:", data);
      }
  
      chats.value = chats.value.slice().sort((a, b) => new Date(b.last_message_time) - new Date(a.last_message_time));
      console.log("Chats after sorting:", chats.value);
    } catch (e) {
      console.error("Error processing WebSocket message:", e);
    }
  };
  
  const fetchActiveUsers = async () => {
    try {
      const response = await axios.get('/v1/active-users', {
        headers: {
          Authorization: `Bearer ${store.user.access}`
        }
      });
      if (response.data && response.data.active_users) {
        activeUsers.value = response.data.active_users;
      } else {
        activeUsers.value = [];
      }
    } catch (err) {
      error.value = err.response ? err.response.data.error : 'An error occurred';
      console.error("Error fetching active users:", err);
    }
  };

  const markChatAsRead = async (chatID) => {
    try {
      await axios.post(`/v1/notifications/chat/mark-read/${chatID}`, {}, {
        headers: {
          Authorization: `Bearer ${store.user.access}`
        }
      });
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
    activeUsers,
    fetchChats,
    deleteChat,
    sortedChats,
    fetchActiveUsers,
    updateMessageTimes,
    handleWebSocketOpen,
    handleWebSocketMessage,
    markChatAsRead,
    handleWebSocketError,
    handleWebSocketClose,
    handleActiveUsersFetched,
    deleteMessagesForUser,
    formatTimeAgo
  };
});
