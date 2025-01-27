<template>
    <div class="inbox-container">
      <ChatListView class="chat-list" />
      <router-view class="chat-view" />
    </div>
  </template>
  
  <script setup>
  import ChatListView from './ChatListView.vue';
  </script>
  
  <style scoped>
  .inbox-container {
    display: flex;
    height: 100dvh;
  }
  
  .chat-list {
    flex: 0 1 25%; 
    border-right: 1px solid #ccc;
    overflow-y: auto;
  }
  
  .chat-view {
    flex: 1;
    overflow-y: auto;
  }
  </style>
  