<template>
    <div>
      <h1>Your Chats</h1>
      <ul>
        <li v-for="chat in chats" :key="chat.chat_id">
          <router-link :to="{ name: 'chat', params: { chatID: chat.chat_id }, query: { user: chat.user } }">{{ chat.user }}</router-link>
        </li>
      </ul>
      <div v-if="error" class="error">{{ error }}</div>
    </div>
  </template>
  
  <script setup>
  import { ref, onMounted } from 'vue';
  import axios from 'axios';
  
  const chats = ref([]);
  const error = ref('');
  
  onMounted(async () => {
    try {
      const response = await axios.get('/v1/me/chats');
      chats.value = response.data.chats;
    } catch (err) {
      error.value = err.response ? err.response.data.error : 'An error occurred';
      console.error("Error fetching chats:", err);
    }
  });
  </script>
  