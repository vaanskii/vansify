<template>
  <div>
    <div class="chatlist-header" v-if="chatStore.activeUsers.length > 0">
      <h2 class="uppercase ml-3 font-bold text-[#757575] mt-2">Active Users</h2>
      <ul class="flex flex-row gap-6 ml-2 mt-2 mb-7">
        <li class="text-center cursor-pointer" v-for="user in chatStore.activeUsers" :key="user.username">
          <img class="w-16 rounded-full" :src="user.profile_picture" alt="Profile Picture" width="30" height="30" />
          <span>{{ user.username }}</span>
        </li>
      </ul>
    </div>

    <h2 class="uppercase mt-0 ml-3 text-[#757575] font-bold msg">Messages</h2>
    <ul v-if="chatStore.chats.length > 0" class="chat-list">
    <li 
      v-for="chat in sortedChats" 
      :key="chat.chat_id" 
      :class="['chat-item', { active: isActiveChat(chat.chat_id) }]"
    >
      <router-link 
        :to="{ name: 'chat', params: { chatID: chat.chat_id }, query: { user: chat.user } }" 
        @click.native="markChatAsRead(chat.chat_id)" 
        class="chat-link"
      >
        <div class="relative">
          <img :src="chat.profile_picture" alt="Profile Picture" class="w-12 rounded-full" />
          <span v-if="isUserActive(chat.user)" 
            class="active-indicator absolute left-8 top-8 z-10 bg-white rounded-l-full rounded-full w-4 h-4 flex items-center justify-center"
            :class="['active-indicator'], {active: isActiveChat(chat.chat_id)}"
            >●
          </span>
        </div>
        <div class="chat-details ml-4 relative">
          <span 
            class="text-[16px] text-black"
            :class="[
              'text-[16px] text-black',
              chat.unread_count > 0 ? 'font-bold' : ''
            ]"
            >{{ chat.user }}
          </span>
          <span v-if="chat.unread_count > 0" class="absolute top-4 right-3 text-[10px] text-[#757575]">●</span>
          <br />
          <div class="flex items-center">
            <div class="w-auto max-w-[240px] overflow-hidden"> 
              <span
                :class="[
                'truncate block w-auto text-[13px] text-black mr-1',
                chat.unread_count > 0 ? 'font-bold' : ''
                ]"
                >{{ chat.last_message }}
              </span>
            </div>
            <span class="text-[14px] text-[#757575]">• {{ formatTime(chat.last_message_time) }} </span>
          </div>
        </div>
      </router-link>
    </li>
    </ul>
    <div v-else>No chats found</div>
    <div v-if="chatStore.error" class="error">{{ chatStore.error }}</div>
  </div>
</template>

<script setup>
import { onMounted, onUnmounted, computed } from 'vue';
import { useChatStore } from '@/stores/chatStore';
import { useActiveUsersStore } from '@/stores/activeUsers';
import { useChatNotificationStore } from '@/stores/chatNotification';
import { userStore } from '@/stores/user';
import emitter from '@/eventBus';
import { useRoute } from 'vue-router'; 

const route = useRoute(); 

const store = userStore();
const chatStore = useChatStore();
const activeUsersStore = useActiveUsersStore();
const chatNotificationStore = useChatNotificationStore()

const sortedChats = computed(() => chatStore.sortedChats);

const isActiveChat = (chatID) => {
  return route.params.chatID === chatID;
};

onMounted(() => {
  if (store.user.isAuthenticated) {
    if (chatStore.chats.length === 0) {
      chatStore.fetchChats();
    }
    
    emitter.on('chat-ws-open', chatStore.handleWebSocketOpen);
    emitter.on('chat-ws-message', chatStore.handleWebSocketMessage);
    emitter.on('chat-ws-error', chatStore.handleWebSocketError);
    emitter.on('chat-ws-close', chatStore.handleWebSocketClose);
    activeUsersStore.connectWebSocket();
    chatNotificationStore.connectWebSocket();

    emitter.on('active-users-ws-open', chatStore.handleWebSocketOpen);
    emitter.on('active-users-ws-error', chatStore.handleWebSocketError);
    emitter.on('active-users-ws-close', chatStore.handleWebSocketClose);
    emitter.on('active-users-fetched', chatStore.handleActiveUsersFetched);

    chatStore.fetchActiveUsers();

    setInterval(chatStore.updateMessageTimes, 60000);
  }
});

onUnmounted(() => {
  emitter.off('ws-open', chatStore.handleWebSocketOpen);
  emitter.off('ws-message', chatStore.handleWebSocketMessage);
  emitter.off('ws-error', chatStore.handleWebSocketError);
  emitter.off('ws-close', chatStore.handleWebSocketClose);

  if (chatStore.wsConnected) {
    chatStore.handleWebSocketClose();
  }
  emitter.off('active-users-ws-open', chatStore.handleWebSocketOpen);
  emitter.off('active-users-ws-error', chatStore.handleWebSocketError);
  emitter.off('active-users-ws-close', chatStore.handleWebSocketClose);
  emitter.off('active-users-fetched', chatStore.handleActiveUsersFetched);
});

const deleteChat = (chatID) => {
  chatStore.deleteChat(chatID);
};

const deleteChatforUser = (chatID) => {
  chatStore.deleteMessagesForUser(chatID)
}

const markChatAsRead = (chatID) => {
  chatStore.markChatAsRead(chatID);
};

const formatTime = (timestamp) => {
  return chatStore.formatTimeAgo(timestamp);
};

const isUserActive = (username) => {
  return chatStore.activeUsers.some(user => user.username === username);
};
</script>

<style scoped>
.chat-list {
  list-style: none;
  margin: 0;
  padding: 0;
}

.chat-item {
  display: flex;
  align-items: center;
  padding: 10px;
  border-bottom: 1px solid #ddd;
  transition: background-color 0.3s ease, color 0.3s ease;
}

.chat-item.active {
  background-color: #D4D4D4;
}

.chat-link {
  display: flex;
  flex: 1;
  align-items: center;
  text-decoration: none;
  color: black;
}

.profile-picture {
  border-radius: 50%;
  margin-right: 10px;
}

.chat-details {
  flex-grow: 1;
}

.chat-user {
  font-size: 16px;
  font-weight: bold;
}

.unread-count {
  color: red;
}

.last-message-time {
  font-size: 12px;
  color: gray;
}

.active-indicator {
  color: green;
  font-size: 14px;
  /* margin-left: 5px; */
}

.active-indicator.active {
  background-color: #D4D4D4;
}

.chat-actions {
  display: flex;
  gap: 10px;
}

.delete-button {
  background-color: #ff4d4f;
  color: white;
  border: none;
  padding: 5px 10px;
  border-radius: 4px;
  cursor: pointer;
  transition: background-color 0.2s ease;
}

.delete-button:hover {
  background-color: #ff7875;
}

@media (width <= 1024px) and (width > 769px) {
  .chat-details {
    display: none;
  }
  .chatlist-header {
    display: none;
  }
  .msg{
    display: none;
  }
}
</style>
