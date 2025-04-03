<template>
  <div class="chat-container">
    <div class="chat-box-container">
      <div v-if="chatID" class="py-4 border-b-[1px] border-[#D4D4D4] flex flex-row items-center justify-between">
        <h2 class="ml-4 hidden md:block cursor-pointer font-bold text-[20px]" @click="messageUtilsStore.goToProfile(chatUser)">{{ chatUser }}</h2>
        <div class="flex flex-row items-center ml-4 block md:hidden gap-4">
          <i class="fa-solid fa-angle-left fa-xl" @click="router.push('/inbox')"></i>
          <h2 class="font-bold" @click="messageUtilsStore.goToProfile(chatUser)">{{ chatUser }}</h2>
        </div>

      <!-- When sidebar is closed -->
      <svg v-if="!isSidebarOpen" 
          xmlns="http://www.w3.org/2000/svg" 
          fill="none" 
          viewBox="0 0 24 24" 
          stroke-width="1.5" 
          stroke="currentColor" 
          class="size-7 cursor-pointer mr-4"
          @click="toggleSidebar">
        <path stroke-linecap="round" stroke-linejoin="round" d="m11.25 11.25.041-.02a.75.75 0 0 1 1.063.852l-.708 2.836a.75.75 0 0 0 1.063.853l.041-.021M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Zm-9-3.75h.008v.008H12V8.25Z" />
      </svg>

      <!-- When sidebar is open -->
      <svg v-else 
          xmlns="http://www.w3.org/2000/svg" 
          viewBox="0 0 24 24" 
          fill="currentColor" 
          class="size-7 cursor-pointer mr-4"
          @click="toggleSidebar">
        <path fill-rule="evenodd" d="M2.25 12c0-5.385 4.365-9.75 9.75-9.75s9.75 4.365 9.75 9.75-4.365 9.75-9.75 9.75S2.25 17.385 2.25 12Zm8.706-1.442c1.146-.573 2.437.463 2.126 1.706l-.709 2.836.042-.02a.75.75 0 0 1 .67 1.34l-.04.022c-1.147.573-2.438-.463-2.127-1.706l.71-2.836-.042.02a.75.75 0 1 1-.671-1.34l.041-.022ZM12 9a.75.75 0 1 0 0-1.5.75.75 0 0 0 0 1.5Z" clip-rule="evenodd" />
      </svg>

      </div>
      <div v-if="!chatID" class="flex flex-col items-center justify-center h-screen">
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" class="size-20">
          <path fill-rule="evenodd" d="M4.848 2.771A49.144 49.144 0 0 1 12 2.25c2.43 0 4.817.178 7.152.52 1.978.292 3.348 2.024 3.348 3.97v6.02c0 1.946-1.37 3.678-3.348 3.97a48.901 48.901 0 0 1-3.476.383.39.39 0 0 0-.297.17l-2.755 4.133a.75.75 0 0 1-1.248 0l-2.755-4.133a.39.39 0 0 0-.297-.17 48.9 48.9 0 0 1-3.476-.384c-1.978-.29-3.348-2.024-3.348-3.97V6.741c0-1.946 1.37-3.68 3.348-3.97ZM6.75 8.25a.75.75 0 0 1 .75-.75h9a.75.75 0 0 1 0 1.5h-9a.75.75 0 0 1-.75-.75Zm.75 2.25a.75.75 0 0 0 0 1.5H12a.75.75 0 0 0 0-1.5H7.5Z" clip-rule="evenodd" />
        </svg>
        <h1 class="text-2xl font-bold uppercase">Your messages</h1>
        <h1 class="no-messages">choose chat to start messaging</h1>
      </div>
      <h1 v-if="formattedMessages.length === 0 && !isLoading" class="no-messages"></h1>
      <div class="flex flex-row overflow-y-auto flex-grow-1">
      <div v-if="!isLoading" class="messages-container" ref="messagesContainer">
        <!-- Loader -->
        <div v-if="loadingOlderMessages" class="flex items-center justify-center">
          <div class="loader border-t-2 rounded-full border-gray-500 bg-gray-300 animate-spin
          aspect-square w-8 flex justify-center items-center text-yellow-700">
        </div>
      </div>
      
      <div v-for="(message, index) in formattedMessages" :key="message.id" 
      class="relative"
      :class="{
        'message': true, 
        'sent mr-1 md:mr-4': message.isOwnMessage && !message.file_url, 
        'received': !message.isOwnMessage && !message.file_url, 
        'sent-image mr-1 md:mr-4 mt-1': message.isOwnMessage && message.file_url, 
        'received-image mt-1': !message.isOwnMessage && message.file_url,
        'mt-5' : detectMessageGap(index)
      }">
          <div class="flex flex-row gap-2 relative group">
          <!-- Profile picture and username (conditionally shown for the last received message) -->
            <img 
            v-if="message.username && !message.isOwnMessage && isLastReceivedMessage(index)" 
            :src="messageUtilsStore.formatProfilePictureUrl(message.profile_picture)"
            alt="Profile Picture" 
            class="w-6 h-6 md:w-8 md:h-8 rounded-full absolute bottom-0 cursor-pointer" 
            @click="messageUtilsStore.goToProfile(message.username)"
            />
            <div 
              class="flex flex-col ml-8 md:ml-10 max-w-[450px]"
              :class="{

                'py-1 px-3': !message.file_url, 
                'py-0 px-0 bg-black': message.file_url, 

                '!bg-transparent': message.file_url,

                // Single message styling (fully rounded for one message)
                'rounded-full text-black bg-blue-200': getMessagePosition(index) === 'single' && message.isOwnMessage,
                'rounded-full bg-[#D4D4D4]': getMessagePosition(index) === 'single' && !message.isOwnMessage,

                // Sender-specific styling
                'rounded-l-3xl rounded-r-md rounded-tr-3xl bg-blue-200': getMessagePosition(index) === 'first' && message.isOwnMessage,
                'rounded-l-3xl rounded-r-md bg-blue-200': getMessagePosition(index) === 'middle' && message.isOwnMessage,
                'rounded-l-3xl rounded-r-md rounded-br-4xl bg-blue-200': getMessagePosition(index) === 'last' && message.isOwnMessage,

                // Receiver-specific styling
                'rounded-r-3xl rounded-l-md rounded-tl-3xl bg-[#D4D4D4]': getMessagePosition(index) === 'first' && !message.isOwnMessage,
                'rounded-r-3xl rounded-l-md bg-[#D4D4D4]': getMessagePosition(index) === 'middle' && !message.isOwnMessage,
                'rounded-r-3xl rounded-l-md rounded-bl-3xl bg-[#D4D4D4]': getMessagePosition(index) === 'last' && !message.isOwnMessage,
              }"
            >
              <!-- Message content -->
              <div v-if="message.file_url" class="w-[175px] h-[230px] md:w-[220px] md:h-[300px] bg-black rounded-3xl">
                <img 
                  :src="message.file_url" 
                  alt="Uploaded Image" 
                  :class="{
                    'w-[175px] h-[230px] md:w-[220px] md:h-[300px] cursor-pointer transition duration-300 hover:brightness-50': true,
                    'rounded-tl-3xl rounded-tr-xl rounded-bl-xl': message.isOwnMessage,
                    'rounded-tr-3xl rounded-tl-xl rounded-br-xl': !message.isOwnMessage
                  }"
                />
              </div>
              <div v-else>
                <p>{{ message.message }}</p>
              </div>

              <span 
                class="opacity-0 group-hover:opacity-100 cursor-pointer transition-opacity duration-200 absolute top-1/2 -translate-y-1/2"
                :class="{
                  'right-full mr-[-20px] pl-0 md:pl-80': message.isOwnMessage,
                  'left-full !ml-[20px] pr-0 md:pr-80': !message.isOwnMessage
                }"
                @click="messageUtilsStore.toggleMessageOptions(message.id, $event)"
              >
                <i class="fa-solid fa-ellipsis-vertical"></i>
              </span>

              <div 
                v-if="messageUtilsStore.openMessageId === message.id" 
                class="absolute bg-white shadow-md rounded-lg p-2 z-10 w-32 flex flex-col space-y-2"
                :class="{
                  'right-full bottom-0': message.isOwnMessage,
                  'left-full ml-10': !message.isOwnMessage,
                }"
                @click.stop
              >
                <div class="flex flex-row items-center space-x-2 rtl:space-x-reverse">
                  <span class="text-sm font-normal text-gray-500 dark:text-gray-900">
                    {{ formatTimeAgo(message.created_at) }}
                  </span>
                </div>

                <!-- Copy Button (Only shown if message has no file) -->
                <button 
                  v-if="!message.file_url" 
                  @click="messageUtilsStore.copyMessage(message.message)" 
                  class="py-2 px-4 bg-gray-100 hover:bg-gray-300 rounded transition cursor-pointer"
                >
                  Copy <i class="fa-solid fa-copy"></i> 
                </button>

                <!-- Delete Button (Only for own messages) -->
                <button 
                  v-if="message.isOwnMessage" 
                  @click="deleteMessage(message.id)" 
                  class="py-2 px-4 bg-red-500 hover:bg-red-700 text-white rounded transition cursor-pointer"
                >
                  Unsent <i class="fa-solid fa-trash"></i>
                </button>
                <button 
                  v-if="!message.isOwnMessage" 
                  class="py-2 px-4 bg-yellow-300 hover:bg-yellow-500 text-white rounded transition cursor-pointer"
                >
                  Report <i class="fa-solid fa-trash"></i>
                </button>
              </div>
            </div>
          </div>

          <!-- Message footer (preserved exactly as you had it) -->
          <div class="message-footer">
            <span v-if="message.isOwnMessage">
              <!-- Show "sending" or "sent/delivered" only on the latest sent message -->
              <span v-if="message.id === latestSentMessageId" class="absolute right-1">
                <span v-if="message.status === 'sending'">Sending...</span>
                <span v-if="message.status === 'sent'">
                   <span>sent</span>
                </span>
                <span v-if="message.status === 'delivered'">
                  <span>delivered</span>
                </span>
              </span>

              <!-- Show "read" status only on the specific read message -->
              <span v-if="message.id === lastReadMessageId " class="absolute right-1">
                <span v-if="message.status === 'read'">
                  <span>seen</span>
                </span>
              </span>
            </span>
          </div>
        </div>
      </div>
      <!-- py-4 border-b-[1px] border-[#D4D4D4] flex flex-row items-center justify-between -->
      <div class="sidebar fixed inset-0 bg-[#ECECEC] w-full h-full transition-all duration-300 border-l-[1px] z-20 md:relative md:w-1/3 md:h-auto" v-if="isSidebarOpen">
        <div class="flex flex-col">
          <div class=" flex flex-row items-center border-b-[1px]">
            <i class="fa-solid ml-4 fa-angle-left fa-xl" @click="toggleSidebar"></i>
            <h1 class="ml-4 uppercase font-bold py-4 text-[16px]">details</h1>
          </div>
          <p class="font-bold mt-4 ml-4">Members</p>
          <div @click="messageUtilsStore.goToProfile(otherUser.username)" class="flex w-full items-center gap-4 flex-row mt-2 hover:bg-[#D4D4D4] p-2 cursor-pointer">
            <img 
              :src="messageUtilsStore.formatProfilePictureUrl(otherUser.profile_picture)" 
              alt="Profile Picture" 
              class="w-12 h-12 rounded-full border"
            />
            <h3 class="text-lg font-bold">{{ otherUser.username }}</h3>
          </div>
          <!-- this should be at the bottom -->
          <div class="border-t-[1px] absolute bottom-20 w-full">
            <button @click="deleteChatforUser(chatID)" class="delete-button text-red-600 px-2 py-4">Delete chat</button>
          </div>
        </div>
      </div>
    </div>

    <div class="form-container">
      <form @submit.prevent="sendMessage" v-if="!isLoading && chatID" class="message-form relative">
         <!-- Image Preview -->
         <div v-if="imagePreview" class="absolute top-2 left-2 w-14 h-14">
          <img :src="imagePreview" alt="Selected Image" class="w-full h-full object-cover rounded-lg shadow-md"/>
          
          <!-- Remove Image Button -->
          <button @click="removeSelectedImage" class="absolute top-0 right-0 bg-red-500 text-white rounded-full w-5 h-5 flex items-center justify-center">
            ✕
          </button>
        </div>
        <textarea 
          :class="{'h-[46px]': !imagePreview, '!pt-18': imagePreview}"
          v-model="newMessage" placeholder="Type a message" class="message-input" rows="1" @input="adjustTextareaHeight" @keydown="handleKeyDown"
        >
        </textarea>

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
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, computed, watch, nextTick } from 'vue';
import axios from 'axios';
import { useRoute, useRouter } from 'vue-router';
import { userStore } from '@/stores/user';
import { useMessageUtilsStore } from '@/stores/messageUtilsStore';
import { useChatStore } from '@/stores/chatStore';
import emitter from '@/eventBus';
import notify from '@/utils/notify';
import "@/assets/chat.css"
import { parseISO, formatDistanceToNow, differenceInMinutes, differenceInHours, differenceInDays } from 'date-fns';
import EXIF from "exif-js";

const apiUrl = import.meta.env.VITE_WS_URL;
const messages = ref([]);
const newMessage = ref('');
const isLoading = ref(false);
const isConnected = ref(false);
const route = useRoute();
const router = useRouter();
const store = userStore();
const messageUtilsStore = useMessageUtilsStore();
const username = store.user.username;
let ws;
let retryAttempt = 0;
const maxRetries = 10;
const chatUser = ref('');
const messagesContainer = ref(null);
const fileInput = ref(null);
const selectedFile = ref(null);
const loadingOlderMessages = ref(false);
const hasMoreMessages = ref(true);
const isSidebarOpen = ref(false);
const chatID = computed(() => route.params.chatID);
const chatStore = useChatStore()

const otherUser = ref({ username: "", profile_picture: "" });

const toggleSidebar = () => {
  isSidebarOpen.value = !isSidebarOpen.value;
};

const deleteChatforUser = (chatID) => {
  chatStore.deleteMessagesForUser(chatID)
}

const wsConst = import.meta.env.VITE_WS;
const token = store.user.access;
let notificationsRead = false;

const closeMenu = (event) => {
  messageUtilsStore.closeMenu(event);
};

const scrollToBottom = () => {
  const container = messagesContainer.value;
  if (container) {
    container.scrollTop = container.scrollHeight;
  }
};

const detectMessageGap = (index) => {
  if (index === 0) return false;

  const currentMessage = formattedMessages.value[index];
  const prevMessage = formattedMessages.value[index - 1];

  const getTimeDifference = (messageA, messageB) => {
    if (!messageA || !messageB) return Infinity;
    return Math.abs(new Date(messageB.created_at) - new Date(messageA.created_at)) / 60000;
  };

  const latestSeenMessageIndex = formattedMessages.value
    .map((msg, i) => (msg.status === "read" && msg.isOwnMessage ? i : null))
    .filter(i => i !== null)
    .pop(); 

  if (prevMessage && index - 1 === latestSeenMessageIndex) {
    return true;
  }

  return getTimeDifference(prevMessage, currentMessage) > 1;
};


const formattedMessages = computed(() => {
  return messages.value.map(message => {
    const isOwnMessage = message.username === username;
    return { ...message, isOwnMessage };
  });
});

const latestSentMessageId = computed(() => {
  const sentMessages = formattedMessages.value.filter(msg => msg.isOwnMessage);
  return sentMessages.length ? sentMessages[sentMessages.length - 1].id : null;
});

const lastReadMessageId = computed(() => {
  // Find the last read message by the second user
  const readMessages = formattedMessages.value.filter(msg => msg.status === 'read' && msg.isOwnMessage);
  return readMessages.length ? readMessages[readMessages.length - 1].id : null;
});

const isLastReceivedMessage = (index) => {
  const currentMessage = formattedMessages.value[index];
  const nextMessage = formattedMessages.value[index + 1];

  if (!nextMessage || nextMessage.isOwnMessage) {
    return true;
  }

  return currentMessage.username !== nextMessage.username;
};

const getMessagePosition = (index) => {
  const prevMessage = formattedMessages.value[index - 1];
  const nextMessage = formattedMessages.value[index + 1];
  const currentMessage = formattedMessages.value[index];

  const getTimeDifference = (messageA, messageB) => {
    if (!messageA || !messageB) return Infinity;
    return Math.abs(new Date(messageB.created_at) - new Date(messageA.created_at)) / 60000;
  };

  const isSingleMessage =
    (!prevMessage || prevMessage.username !== currentMessage.username) &&
    (!nextMessage || nextMessage.username !== currentMessage.username);

  if (isSingleMessage) {
    return 'single';
  }

  const isNewSingleMessage =
    prevMessage &&
    prevMessage.username === currentMessage.username &&
    getTimeDifference(prevMessage, currentMessage) > 1 &&
    (!nextMessage || nextMessage.username !== currentMessage.username);

  if (isNewSingleMessage) {
    return 'single';
  }

  const isFirstMessage =
    (!prevMessage || prevMessage.username !== currentMessage.username || getTimeDifference(prevMessage, currentMessage) > 1);

  const isLastMessage =
    !nextMessage || nextMessage.username !== currentMessage.username;

  if (isFirstMessage) return 'first';
  if (isLastMessage) return 'last';
  return 'middle';
};

function formatTimeAgo(utcTime) {
  const localTime = parseISO(utcTime);

  const minutesDiff = differenceInMinutes(new Date(), localTime);
  if (minutesDiff < 1) {
    return 'just now';
  }

  if (minutesDiff < 60) {
    return `${minutesDiff} min ago`;
  }
  const hoursDiff = differenceInHours(new Date(), localTime);
  if (hoursDiff < 24) {
    return `${hoursDiff} ${hoursDiff === 1 ? 'hour' : 'hours'} ago`;
  }

  const daysDiff = differenceInDays(new Date(), localTime);
  if (daysDiff >= 1) {
    return `${daysDiff} ${daysDiff === 1 ? 'day' : 'days'} ago`;
  }

  return formatDistanceToNow(localTime, { addSuffix: true });
}

const updateMessageTimes = () => {
  messages.value = messages.value.map(message => ({
    ...message,
    time: formatTimeAgo(message.created_at) 
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
  if (ws) {
    ws.onclose = () => {
      console.log("Previous WebSocket closed, establishing new connection");
      setupNewWebSocket(chatID, token);
    };
    ws.close();
  } else {
    setupNewWebSocket(chatID, token);
  }
};

const setupNewWebSocket = (chatID, token) => {
  const wsURL = `${wsConst}//${apiUrl}/v1/chat/${chatID}/ws?token=${encodeURIComponent(token)}`;
  ws = new WebSocket(wsURL);

  ws.onopen = () => {
    isConnected.value = true;
    retryAttempt = 0;
    isLoading.value = false;
    console.log("websocket established in chat", chatID);
    const unsentMessages = messages.value.filter(msg => msg.status === false && msg.isOwnMessage);
    unsentMessages.forEach(message => {
      ws.send(JSON.stringify({ message: message.message, username, chatID })); 
      message.status = true;
    });
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

  ws.onmessage = (event) => {
    const message = JSON.parse(event.data);
    console.log("message", message);
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

watch(
  () => route.params.chatID,
  async (newChatID, oldChatID) => {
    if (newChatID !== oldChatID) {
      intentionalClosure = true;
      connectWebSocket(newChatID, token);
      intentionalClosure = false;
      isSidebarOpen.value = false;
    }
  }
);


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

        if (message.username !== username && !otherUser.value.username) {
          otherUser.value.username = message.username;
          otherUser.value.profile_picture = `/${message.profile_picture}`;
          console.log("Other User Saved:", otherUser.value);
        }

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
          messages.value.push(newMessage); // Append to the bottom
        }
      });

      // Sort messages by created_at to maintain correct order
      messages.value = messages.value.slice().sort((a, b) => new Date(a.created_at) - new Date(b.created_at));
      if (newMessages.length < limit) {
        hasMoreMessages.value = false; 
      }
    }
  } catch (error) {
    console.error('Error fetching chat history:', error);
  } finally {
    if (offset === 0) isLoading.value = false;
  }
};

const loadChat = async () => {
  const chatID = route.params.chatID;
  await fetchChatHistory(chatID);
  chatUser.value = route.query.user;
  connectWebSocket(chatID, token)
}

watch(() => route.params.chatID, () => {
  messages.value = [];
  chatUser.value = '';
  otherUser.value = [];
  loadChat();
})

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
  document.addEventListener('click', closeMenu);
  const token = store.user.access;
  const chatID = route.params.chatID;
  if (chatID && token) {
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
  document.removeEventListener('click', closeMenu);
  if (ws) ws.close();

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

const imagePreview = ref("");

const triggerFileInput = () => {
  fileInput.value.click();
};

const removeSelectedImage = () => {
  imagePreview.value = "";
  selectedFile.value = null;
  fileInput.value.value = "";
};

const onFileSelected = (event) => {
  const file = event.target.files[0];
  selectedFile.value = file;

  if (file) {
    imagePreview.value = URL.createObjectURL(file);
  }

  const reader = new FileReader();
  reader.onload = (e) => {
    const image = new Image();
    image.src = e.target.result;

    EXIF.getData(image, function () {
      const orientation = EXIF.getTag(this, "Orientation");

      if (orientation && orientation !== 1) {
        rotateImage(image, orientation);
      }
    });
  };

  reader.readAsDataURL(file);
};

const rotateImage = (img, orientation, format) => {
  const canvas = document.createElement("canvas");
  const ctx = canvas.getContext("2d");

  const width = img.width;
  const height = img.height;

  canvas.width = orientation > 4 ? height : width;
  canvas.height = orientation > 4 ? width : height;

  ctx.translate(canvas.width / 2, canvas.height / 2);

  switch (orientation) {
    case 2: ctx.scale(-1, 1); break;
    case 3: ctx.rotate(Math.PI); break;
    case 4: ctx.scale(1, -1); break;
    case 5: ctx.rotate(Math.PI / 2); ctx.scale(1, -1); break;
    case 6: ctx.rotate(Math.PI / 2); break;
    case 7: ctx.rotate(-Math.PI / 2); ctx.scale(1, -1); break;
    case 8: ctx.rotate(-Math.PI / 2); break;
  }

  ctx.drawImage(img, -width / 2, -height / 2);
  
  canvas.toBlob((blob) => {
    selectedFile.value = blob;
  }, format, 1);
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
      imagePreview.value = "";
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
    message: trimmedMessage || "Sent a photo",
    created_at: new Date().toISOString(),
    isOwnMessage: true,
    status: 'sending',
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

  // Add the message immediately to the chat history
  messages.value.push(messageToSend);

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
      // Find the message in the chat history and update its status and id
      const messageIndex = messages.value.findIndex(msg => msg.created_at === messageToSend.created_at && msg.username === username);
      if (messageIndex !== -1) {
        messages.value[messageIndex].id = messageID;
        messages.value[messageIndex].status = 'sent'; // Update the status to 'sent'
      }
    } catch (error) {
      console.error("Error receiving message ID:", error);
    }
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
  }
});


watch(messages, () => {
  nextTick(scrollToBottom)
});
</script>

