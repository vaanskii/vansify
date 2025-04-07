<template>
  <div v-if="store.user.isAuthenticated" class="inbox-container">
    <ChatListView class="chat-list" />
    <Chat class="chat-view" />
  </div>
  <div v-else>
    <p>Please make authorization first.</p>
  </div>
</template>
  
<script setup>
import Chat from './Chat.vue';
import ChatListView from './ChatListView.vue';
import { userStore } from '@/stores/user';

const store = userStore();
</script>
  
<style scoped>
.inbox-container {
  display: flex;
  height: 100dvh;
}

.chat-list {
  flex: 0 1 400px; 
  border-right: 1px solid #ccc;
  overflow-y: auto;
}

.chat-view {
  flex: 1;
  overflow-y: auto;
}

@media (width <= 1024px ) {
    .chat-list {
    flex: 0 1 70px; 
  }
}

@media (width <= 48rem) {
  .chat-view {
    display: none;
  }
}
</style>
  