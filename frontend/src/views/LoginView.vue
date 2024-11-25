<template>
  <div>
    <form @submit.prevent="login">
      <div>
        <label for="username">Username:</label>
        <input type="text" v-model="username" id="username" required>
      </div>
      <div>
        <label for="password">Password:</label>
        <input type="password" v-model="password" id="password" required>
      </div>
      <div>
        <router-link to="/forgot-password">Forgot Password?</router-link>
      </div>
      <div>
        <label>
          <input type="checkbox" v-model="rememberMe">
          Remember Me
        </label>
      </div>
      <button type="submit">Login</button>
    </form>
    <button @click="loginWithGoogle">Login with Google</button>
    <div v-if="error" class="error">{{ error }}</div>
    <div v-if="message" class="message">{{ message }}</div>
  </div>
</template>

<script setup>
import { ref } from 'vue';
import axios from 'axios';
import { useRouter } from 'vue-router';
import { userStore } from '@/stores/user';
import { useActiveUsersStore } from '@/stores/activeUsers';

const username = ref('');
const password = ref('');
const rememberMe = ref(false);
const error = ref('');
const message = ref('');
const router = useRouter();
const store = userStore();
const activeUsersStore = useActiveUsersStore();

const login = async () => {
  try {
    console.log('Login function called');
    const response = await axios.post('/v1/login', {
      username: username.value,
      password: password.value,
      remember_me: rememberMe.value,
    });
    console.log('Response received:', response.data);

    message.value = response.data.message;

    const accessToken = response.data.access_token;
    const refreshToken = response.data.refresh_token;
    store.setToken({
      access: accessToken,
      refresh: refreshToken,
      id: response.data.id,
      username: response.data.username,
      email: response.data.email,
      oauth_user: response.data.oauth_user,
    });

    activeUsersStore.connectWebSocket();

    router.push('/');
  } catch (err) {
    console.error('Login failed:', err);
    error.value = err.response ? err.response.data.error : 'An error occurred';
  }
};

const loginWithGoogle = () => {
  console.log('Login with Google called');
  const apiUrl = import.meta.env.VITE_API_URL;
  window.location.href = `${apiUrl}/v1/auth/google`;
};
</script>
