<template>
  <div class="chat-container">
    <h2>Chat with {{ chatUser }}</h2>
    <div v-if="formattedMessages.length === 0 && !isLoading" class="no-messages">No messages yet</div>
    <div v-if="!isLoading" class="messages-container" ref="messagesContainer">
      <div v-if="loadingOlderMessages" class="loader">Loading...</div>
      <div v-for="message in formattedMessages" :key="message.id" :class="{'message': true, 'sent': message.isOwnMessage, 'received': !message.isOwnMessage}">
        <div class="message-header" v-if="message.username && !message.isOwnMessage" @click="goToProfile(message.username)">
          <img :src="message.profile_picture" alt="Profile Picture" class="profile-picture" />
          <strong>{{ message.username }}</strong>
        </div>
        <div class="message-body">
          <div v-if="message.file_url">
            <img :src="message.file_url" alt="Uploaded Image" class="uploaded-image"/>
          </div>
          <div v-else>
            {{ message.message }}
          </div>
        </div>
        <div class="message-footer">
          <span v-if="message.isOwnMessage">
            <span v-if="message.status === true">(Sent)</span> 
            <span v-if="message.status === false">(Sending...)</span>
          </span>
          <span>{{ formatTime(message.created_at) }}</span>
          <button v-if="message.isOwnMessage" @click="deleteMessage(message.id)" class="delete-button">Delete</button>
        </div>
      </div>
    </div>
    <form @submit.prevent="sendMessage" v-if="!isLoading" class="message-form">
      <input v-model="newMessage" placeholder="Type a message" class="message-input" />
      <input type="file" ref="fileInput" @change="onFileSelected" accept="image/*" />
      <button type="submit" class="send-button">Send</button>
    </form>
  </div>
</template>


<script setup>
import { ref, onMounted, onUnmounted, computed, watch, nextTick } from 'vue';
import axios from 'axios';
import { useRoute, useRouter } from 'vue-router';
import { userStore } from '@/stores/user';
import emitter from '@/eventBus';

const apiUrl = import.meta.env.VITE_WS_URL;
const messages = ref([]);
const newMessage = ref('');
const isLoading = ref(true);
const isConnected = ref(false);
const route = useRoute();
const router = useRouter();
const store = userStore();
const username = store.user.username;
let ws;
let retryAttempt = 0;
const maxRetries = 10;
const offlineStorageKey = `offline-messages-${route.params.chatID}`;
const chatUser = ref('');
const messagesContainer = ref(null);
const fileInput = ref(null);
const selectedFile = ref(null);
const loadingOlderMessages = ref(false);
const hasMoreMessages = ref(true);

const goToProfile = (username) => {
  router.push({ name: 'userprofile', params: { username }})
};

const scrollToBottom = () => {
  const container = messagesContainer.value;
  if (container) {
    container.scrollTop = container.scrollHeight;
  }
};

const formattedMessages = computed(() => {
  return messages.value.map(message => {
    const isOwnMessage = message.username === username;
    return { ...message, isOwnMessage };
  });
});

const loadOfflineMessages = () => {
  const savedMessages = localStorage.getItem(offlineStorageKey);
  if (savedMessages) {
    const parsedMessages = JSON.parse(savedMessages);
    parsedMessages.forEach(msg => {
      messages.value.push({ ...msg, isOwnMessage: true, status: false });
    });
  }
};

const saveOfflineMessages = () => {
  const unsentMessages = messages.value.filter(msg => msg.status === false && msg.isOwnMessage);
  localStorage.setItem(offlineStorageKey, JSON.stringify(unsentMessages));
};

const removeOfflineMessages = () => {
  localStorage.removeItem(offlineStorageKey);
};

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
  messages.value = messages.value.map(message => ({
    ...message,
    time: formatTime(message.created_at) 
  }));
};

const deleteMessage = async (messageID) => {
  try {
    await axios.delete(`/v1/message/${messageID}`, {
      headers: {
        Authorization: `Bearer ${store.user.access}`
      }
    });
    messages.value = messages.value.filter(message => message.id !== messageID);
  } catch (err) {
    console.error("Error deleting message:", err);
  }
};

const connectWebSocket = (chatID, token) => {
  const wsURL = `ws://${apiUrl}/v1/chat/${chatID}/ws?token=${encodeURIComponent(token)}`;
  ws = new WebSocket(wsURL);

  ws.onopen = () => {
    console.log('WebSocket connection established');
    isConnected.value = true;
    retryAttempt = 0;
    isLoading.value = false;

    // Resend unsent messages
    const unsentMessages = messages.value.filter(msg => msg.status === false && msg.isOwnMessage);
    unsentMessages.forEach(message => {
      ws.send(JSON.stringify({ message: message.message, username }));
      message.status = true;
      console.log('Resent message:', message);
    });
    removeOfflineMessages();
  };

  ws.onerror = (error) => {
    console.error('WebSocket error:', error);
    isConnected.value = false;
  };

  ws.onclose = () => {
    console.log('WebSocket connection closed');
    isConnected.value = false;
    if (retryAttempt < maxRetries) {
      setTimeout(() => {
        retryAttempt++;
        console.log(`Reconnecting... Attempt ${retryAttempt}`);
        connectWebSocket(chatID, token);
      }, Math.min(1000 * Math.pow(2, retryAttempt), 30000));
    } else {
      console.error('Max reconnection attempts reached.');
    }
  };

  ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  if (message.type === 'MESSAGE_DELETED') {
    const index = messages.value.findIndex(msg => msg.id == message.message_id);
    if (index !== -1) {
      messages.value.splice(index, 1);
    }
  } else {
    if (message.chat_id === route.params.chatID) {
      if (message.id && message.username !== username) {
        if (!messages.value.some(msg => msg.id == message.id)) {
          messages.value.push({
            ...message,
            isOwnMessage: message.username === username,
            profile_picture: `/${message.profile_picture}`,
            last_message: message.last_message,
            file_url: message.file_url
          });
        }
      } else {
        const index = messages.value.findIndex(msg => msg.id == message.tempID);
        if (index !== -1) {
          messages.value[index] = { ...messages.value[index], id: message.id, status: true };
        }
      }
    }
  }
  if (route.params.chatID === chatID) {
    markChatNotificationsAsRead(chatID);
  }
};
};

const fetchChatHistory = async (chatID, limit = 20, offset = 0) => {
  try {
    const response = await axios.get(`/v1/chat/${chatID}/history`, {
      headers: {
        Authorization: `Bearer ${store.user.access}`
      },
      params: {
        user: route.query.user, 
        limit,
        offset
      }
    });
    if (response.data) {
      const newMessages = response.data.map(message => ({
        ...message,
        isOwnMessage: message.username === username,
        status: true,
        time: new Date(message.created_at), 
        profile_picture: `/${message.profile_picture}`
      }));

      const existingMessageIds = new Set(messages.value.map(msg => msg.id));
      newMessages.forEach(newMessage => {
        if (!existingMessageIds.has(newMessage.id)) {
          messages.value.unshift(newMessage);
        }
      });

      messages.value.sort((a, b) => new Date(a.time) - new Date(b.time));

      if (newMessages.length < limit) {
        hasMoreMessages.value = false; 
      }
      if (offset === 0) loadOfflineMessages();
    }
  } catch (error) {
    console.error('Error fetching chat history:', error);
  } finally {
    if (offset === 0) isLoading.value = false;
  }
};

const markChatNotificationsAsRead = async (chatID) => {
  try {
    await axios.post(`/v1/notifications/chat/mark-read/${chatID}`, {}, {
      headers: {
        Authorization: `Bearer ${store.user.access}`
      }
    });
    emitter.emit('chat-read', chatID);
  } catch (error) {
    console.error('Error marking notifications as read:', error);
  }
};

const saveScrollPosition = () => {
  const container = messagesContainer.value;
  if (container) {
    return container.scrollHeight - container.scrollTop;
  }
  return 0;
};

const restoreScrollPosition = (scrollPosition) => {
  const container = messagesContainer.value;
  if (container) {
    container.scrollTop = container.scrollHeight - scrollPosition;
  }
};

const loadMoreMessages = async () => {
  if (loadingOlderMessages.value || !hasMoreMessages.value) return;
  loadingOlderMessages.value = true;
  const currentMessageCount = messages.value.length;
  const scrollPosition = saveScrollPosition(); 
  await fetchChatHistory(route.params.chatID, 20, currentMessageCount);
  nextTick(() => {
    restoreScrollPosition(scrollPosition);
    loadingOlderMessages.value = false;
  });
};

const onScroll = () => {
  const container = messagesContainer.value;
  if (container && container.scrollTop === 0) {
    loadMoreMessages();
  }
};

onMounted(async () => {
  const token = store.user.access;
  const chatID = route.params.chatID;
  if (chatID && token) {
    loadOfflineMessages();
    await fetchChatHistory(chatID);
    chatUser.value = route.query.user || 'Unknown';
    connectWebSocket(chatID, token);
    await markChatNotificationsAsRead(chatID);

    setInterval(updateMessageTimes, 60000);
    nextTick(scrollToBottom);
    const container = messagesContainer.value;
    if (container) {
      container.addEventListener('scroll', onScroll);
    }
  } else {
    isLoading.value = false;
  }
});

onUnmounted(() => {
  if (ws) ws.close();
  saveOfflineMessages();

  const container = messagesContainer.value;
  if (container) {
    container.removeEventListener('scroll', onScroll);
  }
});



const onFileSelected = (event) => {
  selectedFile.value = event.target.files[0];
};

const uploadFile = async (file, tempMessageID) => {
  const formData = new FormData();
  formData.append('file', file);
  formData.append('metadata', new Blob([JSON.stringify({
    name: file.name,
    parents: [route.params.chatID]
  })], { type: 'application/json' }));

  try {
    const response = await axios.post(`/v1/upload/chat/${route.params.chatID}`, formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
        Authorization: `Bearer ${store.user.access}`
      }
    });
    if (response.data) {
      const index = messages.value.findIndex(msg => msg.id === tempMessageID);
      if (index !== -1) {
        messages.value[index] = {
          ...messages.value[index],
          file_url: response.data.fileURL,
          message: "Sent picture",
          status: true,
        };
        ws.send(JSON.stringify({ ...messages.value[index], tempID: tempMessageID }));
        saveOfflineMessages();
      }
    }
  } catch (error) {
    console.error("Error uploading file:", error);
    const index = messages.value.findIndex(msg => msg.id === tempMessageID);
    if (index !== -1) {
      messages.value[index].status = false;
      saveOfflineMessages();
    }
  }
};

const sendMessage = async () => {
  if (!newMessage.value && !selectedFile.value) {
    return;
  }

  if (selectedFile.value) {
    const file = selectedFile.value;
    const tempFileMessage = {
      id: Date.now(),
      username,
      file_url: URL.createObjectURL(file),
      created_at: new Date().toISOString(),
      isOwnMessage: true,
      status: false,
    };
    messages.value.push(tempFileMessage);
    const tempMessageID = tempFileMessage.id;
    saveOfflineMessages();

    await uploadFile(file, tempMessageID);
    selectedFile.value = null;
    fileInput.value.value = "";
  } else {
    const tempTextMessage = {
      id: Date.now(),
      username,
      message: newMessage.value,
      created_at: new Date().toISOString(),
      isOwnMessage: true,
      status: false,
    };
    messages.value.push(tempTextMessage);
    const tempMessageID = tempTextMessage.id;

    if (ws && isConnected.value) {
      ws.send(JSON.stringify(tempTextMessage));
      const receiveResponse = new Promise((resolve, reject) => {
        const responseHandler = (event) => {
          const response = JSON.parse(event.data);
          if (response.id) {
            ws.removeEventListener('message', responseHandler);
            resolve(response.id);
          } else {
            reject("Message ID not received");
          }
        };
        ws.addEventListener('message', responseHandler);
      });
      try {
        const messageID = await receiveResponse;
        const index = messages.value.findIndex(msg => msg.id === tempMessageID);
        if (index !== -1) {
          messages.value[index] = { ...tempTextMessage, id: messageID, status: true };
        }
      } catch (error) {
        console.error("Error receiving message ID:", error);
      }
    } else {
      saveOfflineMessages();
    }
    newMessage.value = '';
    nextTick(scrollToBottom);
  }
};

watch(isConnected, (newVal) => {
  if (newVal) {
    const unsentMessages = messages.value.filter(msg => msg.status === false && msg.isOwnMessage);
    unsentMessages.forEach(message => {
      ws.send(JSON.stringify({ message: message.message, username }));
      message.status = true;
    });
    removeOfflineMessages();
  }
});

watch(messages, () => {
  nextTick(scrollToBottom)
});
</script>

<style scoped>
.chat-container {
  display: flex;
  flex-direction: column;
  max-width: 600px;
  margin: 0 auto;
  padding: 20px;
  border: 1px solid #ccc;
  border-radius: 10px;
  background-color: #f9f9f9;
  margin-top: 200px;
}

.loader {
  text-align: center;
  padding: 10px;
  font-size: 16px;
  color: #666;
}


.no-messages {
  text-align: center;
  color: #888;
}

.messages-container {
  display: flex;
  flex-direction: column;
  gap: 10px;
  height: 400px;
  overflow-y: auto;
  margin-bottom: 10px;
}

.message {
  display: flex;
  flex-direction: column;
  padding: 10px;
  border-radius: 10px;
  word-wrap: break-word;
  max-width: 60%;
}

.sent {
  align-self: flex-end;
  background-color: #daf8cb;
}

.uploaded-image{
  width: 150px;
}

.received {
  align-self: flex-start;
  background-color: #e4e6eb;
}

.message-header {
  display: flex;
  align-items: center;
  margin-bottom: 5px;
}

.profile-picture {
  width: 30px;
  height: 30px;
  border-radius: 50%;
  margin-right: 10px;
}

.message-body {
  font-size: 14px;
}

.message-footer {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  color: #888;
}

.message-form {
  display: flex;
}

.message-input {
  flex-grow: 1;
  padding: 10px;
  border: 1px solid #ccc;
  border-radius: 10px;
  margin-right: 10px;
}

.send-button {
  background-color: #007bff;
  color: white;
  border: none;
  padding: 10px;
  border-radius: 10px;
  cursor: pointer;
}

.send-button:hover {
  background-color: #0056b3;
}

.delete-button {
  background-color: #dc3545;
  color: white;
  border: none;
  padding: 5px;
  border-radius: 5px;
  cursor: pointer;
}

.delete-button:hover {
  background-color: #c82333;
}
</style>