<template>
  <div class="login-container flex flex-col gap-3 max-w-[26rem] w-full mx-auto px-10 py-10">
    <form @submit.prevent="login">
      <div class="relative z-0">
          <input required type="text" id="username" v-model="username" autocomplete="username"  class="block py-2.5 px-0 w-full text-sm text-gray-900 bg-transparent border-0 border-b-2 border-gray-300 appearance-none dark:text-black dark:border-gray-600 dark:focus:border-blue-500 focus:outline-none focus:ring-0 focus:border-blue-600 peer" placeholder=" " />
          <label for="username" class="absolute text-sm text-gray-500 dark:text-gray-[#757575] duration-300 transform -translate-y-6 scale-75 top-3 -z-10 origin-[0] peer-focus:start-0 peer-focus:text-blue-600 peer-focus:dark:text-blue-500 peer-placeholder-shown:scale-100 peer-placeholder-shown:translate-y-0 peer-focus:scale-75 peer-focus:-translate-y-6 rtl:peer-focus:translate-x-1/4 rtl:peer-focus:left-auto">Username</label>
      </div>
      <div class="relative z-0 mt-5">
        <input required type="password" id="password" v-model="password" autocomplete="current-password" class="block py-2.5 px-0 w-full text-sm text-gray-900 bg-transparent border-0 border-b-2 border-gray-300 appearance-none dark:text-black dark:border-gray-600 dark:focus:border-blue-500 focus:outline-none focus:ring-0 focus:border-blue-600 peer" placeholder=" " />
        <label for="password" class="absolute text-sm text-gray-500 dark:text-gray-[#757575] duration-300 transform -translate-y-6 scale-75 top-3 -z-10 origin-[0] peer-focus:start-0 peer-focus:text-blue-600 peer-focus:dark:text-blue-500 peer-placeholder-shown:scale-100 peer-placeholder-shown:translate-y-0 peer-focus:scale-75 peer-focus:-translate-y-6 rtl:peer-focus:translate-x-1/4 rtl:peer-focus:left-auto">Password</label>
      </div>
      <div class="flex items-center justify-between mt-4 xs:flex-row gap-2 flex-col">
        <label class="flex items-center space-x-2 text-sm sm:text-base text-[#757575]">
          <input class="text-sm sm:text-base" type="checkbox" v-model="rememberMe">
          <span class="text-sm sm:text-base">Remember Me</span>
        </label>
        <router-link class="text-sm sm:text-base text-[#757575]" to="/forgot-password">Forgot Password?</router-link>
      </div>
      <button class="w-full mt-6 shadow shadow-hover-sm cursor-pointer h-11 rounded-[3px] text-[14px] font-medium text-[#757575]" type="submit">
        <span v-if="isLoading"> 
          <div class="animate-spin inline-block size-6 border-3 border-current border-t-transparent text-[#757575] rounded-full" role="status" aria-label="loading">
            <span class="sr-only">Loading...</span>
          </div>
        </span>
        <span v-else>Login</span>
      </button> 
    </form>
    <button @click="loginWithGoogle" type="button" class="login-with-google-btn" >Sign in with Google</button>
    <div v-if="router.currentRoute.value.path === '/'">
      <div class="flex items-center my-4">
      <div class="flex-grow border-t border-gray-300"></div>
      <span class="mx-4 text-sm text-gray-500">or</span>
      <div class="flex-grow border-t border-gray-300"></div>
    </div>

    <router-link to="/signup" class="w-full mt-6 shadow shadow-hover-md bg-[#757575] hover:bg-[#707070] cursor-pointer h-11 rounded-[3px] text-[14px] font-medium text-white flex items-center justify-center">
        Create New Account
    </router-link>
    </div>
    <div v-else-if="router.currentRoute.value.path === '/login'" class="text-center">
      <span class="text-[#757575]">• </span>
      <router-link to="/signup" class="text-[#757575] text-sm underline">Sign up for Vansify</router-link>
    </div>
    <div v-if="error" class="error">{{ error }}</div>
    <div v-if="message" class="message">{{ message }}</div>
  </div>
</template>

<script setup>
import "@/assets/buttons.css"
import "@/assets/main.css"
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
const isLoading = ref(false);
const router = useRouter();
const store = userStore();
const activeUsersStore = useActiveUsersStore();

const login = async () => {
  isLoading.value = true;
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
  } finally {
    isLoading.value = false;
  }
};

const loginWithGoogle = () => {
  console.log('Login with Google called');
  const apiUrl = import.meta.env.VITE_API_URL;
  window.location.href = `${apiUrl}/v1/auth/google`;
};
</script>

<style scoped>
.login-container {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  gap: 1;
}
</style>
