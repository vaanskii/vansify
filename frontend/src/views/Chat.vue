<template>
  <div class="chat-container">
    <div class="chat-box-container">
      <div v-if="formattedMessages.length === 0 && !isLoading" class="no-messages">No messages yet</div>
      <h2>Chat with {{ chatUser }}</h2>
    <div v-if="!isLoading" class="messages-container" ref="messagesContainer">
      <div v-if="loadingOlderMessages" class="loader">Loading...</div>
      <div v-for="message in formattedMessages" :key="message.id" 
           :class="{
             'message': true, 
             'sent': message.isOwnMessage && !message.file_url, 
             'received': !message.isOwnMessage && !message.file_url, 
             'sent-image': message.isOwnMessage && message.file_url, 
             'received-image': !message.isOwnMessage && message.file_url
           }">
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
            <span v-if="message.status">({{ message.status }})</span> 
          </span>
          <span>{{ formatTime(message.created_at) }}</span>
          <button v-if="message.isOwnMessage" @click="deleteMessage(message.id)" class="delete-button">Delete</button>
        </div>
      </div>
    </div>
      <form @submit.prevent="sendMessage" v-if="!isLoading" class="message-form">
        <textarea v-model="newMessage" placeholder="Type a message" class="message-input" rows="1" @input="adjustTextareaHeight" @keydown="handleKeyDown"></textarea>
        <div class="fileUpload" @click="triggerFileInput">
          <input type="file" class="upload" ref="fileInput" @change="onFileSelected" accept="image/*" style="display: none;"/>
          <i class="fa-solid fa-image fa-lg"></i>
        </div>
        <button type="submit" class="send-button">
                <i class="fa-solid fa-paper-plane fa-lg"></i>
        </button>
      </form>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, computed, watch, nextTick } from 'vue';
import axios from 'axios';
import { useRoute, useRouter } from 'vue-router';
import { userStore } from '@/stores/user';
import emitter from '@/eventBus';
import notify from '@/utils/notify';
import "@/assets/chat.css"

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
const wsConst = import.meta.env.VITE_WS;
let notificationsRead = false;

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

let intentionalClosure = false;

const connectWebSocket = (chatID, token) => {
  const wsURL = `${wsConst}//${apiUrl}/v1/chat/${chatID}/ws?token=${encodeURIComponent(token)}`;
  ws = new WebSocket(wsURL);

  ws.onopen = () => {
    isConnected.value = true;
    retryAttempt = 0;
    isLoading.value = false;
    console.log("websocket established in chat")
    const unsentMessages = messages.value.filter(msg => msg.status === false && msg.isOwnMessage);
    unsentMessages.forEach(message => {
      ws.send(JSON.stringify({ message: message.message, username }));
      message.status = true;
    });
    removeOfflineMessages();
  };

  ws.onerror = (error) => {
    console.error('WebSocket error:', error);
    isConnected.value = false;
  };

  ws.onclose = () => {
  isConnected.value = false;
  if (!intentionalClosure && retryAttempt < maxRetries) {
    setTimeout(() => {
      retryAttempt++;
      connectWebSocket(chatID, token);
    }, Math.min(1000 * Math.pow(2, retryAttempt), 30000));
  } else if (!intentionalClosure) {
    console.error('Max reconnection attempts reached.');
  } else {
    console.log("WebSocket intentionally closed, not attempting to reconnect.");
    intentionalClosure = false;
  }
};

// Watch for changes in route.params.chatID to handle WebSocket closure when chatID becomes invalid
watch(
  () => route.params.chatID,
  (newChatID, oldChatID) => {
    if (!newChatID) {
      intentionalClosure = true;
      if (ws && ws.readyState === WebSocket.OPEN) {
        ws.close();
        console.log("WebSocket closed because chatID is invalid");
      }
    } else if (newChatID !== oldChatID) {
      intentionalClosure = true; 
      if (ws && ws.readyState === WebSocket.OPEN) {
        ws.close();
        console.log("WebSocket closed to reconnect to a new chatID");
      }
      intentionalClosure = false;
      connectWebSocket(newChatID, token);
    }
  }
);


  ws.onmessage = (event) => {
    const message = JSON.parse(event.data);
    console.log("message", message)
    message.message_id = message.message_id || message.id;

    switch (message.type) {
      case 'MESSAGE_ID':
        const messageId = message.id;
        console.log(`Received message ID: ${messageId}`);
        break;

      case 'MESSAGE_DELETED':
        const deleteIndex = messages.value.findIndex(msg => msg.id == message.message_id);
        if (deleteIndex !== -1) {
          messages.value.splice(deleteIndex, 1);
        }
        break;

      case 'STATUS_UPDATE':
        if (message.chat_id === route.params.chatID && message.message_ids) {
          message.message_ids.forEach((msgID) => {
            const updateIndex = messages.value.findIndex(msg => msg.id == msgID);
            if (updateIndex !== -1) {
              messages.value[updateIndex].status = message.status;
              console.log(`Message status updated to ${message.status} for message ID ${msgID}`);
            } else {
              console.log(`Message ID ${msgID} not found`);
            }
          });
        } else {
          console.log(`Chat ID mismatch or no message_ids in STATUS_UPDATE for chat ${message.chat_id}`);
        }
        break;

      case 'STATUS_UPDATE_READ':
        if (message.chat_id === route.params.chatID && message.username !== store.user.username) {
          messages.value.forEach((msg) => {
            if (msg.status !== 'read') {
              msg.status = 'read';
              console.log(`Message status updated to read for chat ${message.chat_id}`);
            }
          });
        } else if (message.chat_id !== route.params.chatID) {
          console.log(`Chat ID mismatch in STATUS_UPDATE_READ for chat ${message.chat_id}`);
        } else {
          console.log(`STATUS_UPDATE_READ received for sender's own message, no update needed`);
        }
        break;

      default:
        if (message.chat_id === route.params.chatID) {
          if (message.username === store.user.username) {
            const ownIndex = messages.value.findIndex(msg => msg.id == message.message_id);
            if (ownIndex !== -1) {
              messages.value[ownIndex].status = message.status;
            } else {
              console.log(`Message ID ${message.message_id} not found`);
            }
          } else {
            if (!messages.value.some(msg => msg.id == message.message_id)) {
              messages.value.push({
                ...message,
                isOwnMessage: message.username === store.user.username,
                profile_picture: `/${message.profile_picture}`,
                last_message: message.last_message,
                file_url: message.file_url,
                receiver: message.receiver,
              });
            }
          }
        }
        break;
    }
      if (route.params.chatID === message.chat_id && message.username !== store.user.username && !notificationsRead) {
        markChatNotificationsAsRead(message.chat_id);
        notificationsRead = true; 
      }
  };

};


const fetchChatHistory = async (chatID, limit = 20, offset = 0) => {
  try {
    console.log('Fetching chat history for chatID:', chatID, 'with limit:', limit, 'and offset:', offset);
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
    console.log('Response received:', response.data);

    if (response.data) {
      const newMessages = response.data.map(message => {
        const localTime = new Date(message.created_at).toLocaleString();
        console.log('Converted time:', localTime);
        
        return {
          ...message,
          isOwnMessage: message.username === username,
          status: message.status,
          time: localTime,
          profile_picture: `/${message.profile_picture}`
        };
      });

      const existingMessageIds = new Set(messages.value.map(msg => msg.id));
      newMessages.forEach(newMessage => {
        if (!existingMessageIds.has(newMessage.id)) {
          messages.value.unshift(newMessage);
        }
      });

      messages.value.sort((a, b) => new Date(a.time) - new Date(b.time));
      console.log('Sorted messages:', messages.value);

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

    messages.value.forEach(message => {
      if (message.chat_id === chatID && message.username !== username) {
        message.status = 'read';
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
    notificationsRead = false;

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

const handleKeyDown = (event) => {
  if (event.key === 'Enter' && !event.shiftKey) {
    event.preventDefault();
    sendMessage();
  } else {
    adjustTextareaHeight(event);
  }
};


const adjustTextareaHeight = (event) => {
  const textarea = event.target;
  textarea.style.height = 'auto';
  textarea.style.height = textarea.scrollHeight + 'px';

  const maxHeight = 150;
  if (textarea.scrollHeight > maxHeight) {
    textarea.style.height = maxHeight + 'px';
    textarea.style.overflowY = 'auto';
  }
};

const triggerFileInput = () => {
  fileInput.value.click();
};

const onFileSelected = (event) => {
  selectedFile.value = event.target.files[0];
};

const uploadFile = async (file) => {
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
      },
      withCredentials: false
    });

    if (response.data) {
      return response.data.fileURL;
    } else {
      throw new Error("File upload failed");
    }
  } catch (error) {
    console.error("Error uploading file:", error);
    throw error;
  }
};

const sendMessage = async () => {
  const trimmedMessage = newMessage.value.trim();

  // Only proceed if there's text or an image
  if (!trimmedMessage && !selectedFile.value) {
    return;
  }

  let messageToSend = {
    username,
    message: trimmedMessage || "Sent a file",
    created_at: new Date().toISOString(),
    isOwnMessage: true,
  };

  console.log("message to send", messageToSend);

  if (selectedFile.value) {
    notify("Sending image...", "info");
    try {
      const fileURL = await uploadFile(selectedFile.value);
      messageToSend.file_url = fileURL;
      selectedFile.value = null;
      fileInput.value.value = "";
      notify("Image sent!", "success");
    } catch (error) {
      notify("Failed to upload file.", "error");
      return;
    }
  }

  if (ws && isConnected.value) {
    ws.send(JSON.stringify(messageToSend));
    const receiveResponse = new Promise((resolve, reject) => {
      const responseHandler = (event) => {
        const response = JSON.parse(event.data);
        if (response.type === 'MESSAGE_ID' && response.id) {
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
      messageToSend.id = messageID;
      messageToSend.status = 'sent';
      messages.value.push(messageToSend);
    } catch (error) {
      console.error("Error receiving message ID:", error);
    }
  } else {
    saveOfflineMessages();
  }

  newMessage.value = '';
  nextTick(scrollToBottom);
};



watch(isConnected, (newVal) => {
  if (newVal) {
    const unsentMessages = messages.value.filter(msg => msg.status === 'sent' && msg.isOwnMessage);
    unsentMessages.forEach(message => {
      ws.send(JSON.stringify(message));
      message.status = 'sent';
    });
    removeOfflineMessages();
  }
});



watch(messages, () => {
  nextTick(scrollToBottom)
});
</script>