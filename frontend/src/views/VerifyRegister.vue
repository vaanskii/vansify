<template>
    <div class="verify-email-container">
      <h1>Email Verification</h1>
      <p v-if="message">{{ message }}</p>
      <div v-if="loading" class="loader">Loading...</div>
      <div v-if="error" class="error">{{ error }}</div>
    </div> 
  </template>
  
  <script setup>
  import { ref, onMounted } from 'vue';
  import axios from 'axios';
  import { useRoute, useRouter } from 'vue-router';
  
  const message = ref('');
  const loading = ref(true);
  const error = ref('');
  
  const route = useRoute();
  const router = useRouter();
  
  const verifyEmail = async (token) => {
    try {
      const response = await axios.get(`/v1/verify?token=${token}`);
      message.value = response.data.message;
      loading.value = false;
      setTimeout(() => {
        router.push('/login'); 
      }, 5000);
    } catch (err) {
      error.value = err.response.data.error || 'Verification failed. Please try again.';
      loading.value = false;
    }
  };
  
  onMounted(() => {
    const token = route.query.token;
    if (token) {
      verifyEmail(token);
    } else {
      error.value = 'Invalid or missing token';
      loading.value = false;
    }
  });
  </script>
  
  <style scoped>
  .verify-email-container {
    max-width: 400px;
    margin: 0 auto;
    padding: 20px;
    text-align: center;
  }
  
  .loader {
    font-size: 16px;
    color: #007bff;
  }
  
  .error {
    color: red;
  }
  </style>
  