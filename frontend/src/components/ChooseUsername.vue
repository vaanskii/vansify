<template>
    <div class="username-selection">
      <h1>Choose Your Username</h1>
      <form @submit.prevent="submitUsername">
        <label for="username">Username:</label>
        <input type="text" id="username" v-model="username" required />
        <button type="submit">Submit</button>
      </form>
      <p v-if="error">{{ error }}</p>
    </div>
  </template>
  
  <script setup>
import { ref } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import axios from 'axios';
import { userStore } from '@/stores/user';

const router = useRouter();
const route = useRoute();
const store = userStore();
const username = ref('');
const error = ref('');

const submitUsername = async () => {
  try {
    const email = route.query.email;
    const response = await axios.post('/v1/create-user', {
      username: username.value,
      email: email,
      active: true
    });

    // Extract tokens and additional user data from the response
    const { access_token, refresh_token, id, oauth_user } = response.data;

    // Store tokens and user information
    store.setToken({
      access: access_token,
      refresh: refresh_token,
      id: id,
      username: username.value,
      email: email,
      oauth_user: oauth_user,
    });

    // Redirect to the home page
    router.push('/');
  } catch (err) {
    error.value = err.response ? err.response.data.error : 'An error occurred';
  }
};
</script>

  
  <style scoped>
  .username-selection {
    max-width: 400px;
    margin: 0 auto;
    padding: 20px;
    border: 1px solid #ccc;
    border-radius: 5px;
  }
  label {
    display: block;
    margin-bottom: 8px;
  }
  input {
    width: 100%;
    padding: 8px;
    margin-bottom: 16px;
  }
  button {
    padding: 10px 20px;
  }
  p {
    color: red;
  }
  </style>
  