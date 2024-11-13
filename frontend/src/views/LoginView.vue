<template>
  <div>
    <form @submit.prevent="login">
      <!-- Existing login form fields -->
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

const username = ref('');
const password = ref('');
const rememberMe = ref(false);
const error = ref('');
const message = ref('');
const router = useRouter();
const store = userStore();

const login = async () => {
  try {
    const response = await axios.post('/v1/login', {
      username: username.value,
      password: password.value,
      remember_me: rememberMe.value,
    });
    message.value = response.data.message;

    // Handle success - save tokens and user info
    const accessToken = response.data.access_token;
    const refreshToken = response.data.refresh_token;
    store.setToken({
      access: accessToken,
      refresh: refreshToken,
      id: response.data.id,
      username: response.data.username,
      email: response.data.email,
    });
    
    router.push('/');
  } catch (err) {
    error.value = err.response ? err.response.data.error : 'An error occurred';
  }
};

const loginWithGoogle = () => {
  window.location.href = '/v1/auth/google';
};

</script>
