<template>
  <div class="signup-container flex flex-col gap-3 max-w-[26rem] w-full mx-auto px-10 py-10">
    <form @submit.prevent="register">
      <div class="relative z-0">
          <input required type="text" id="username" v-model="username" autocomplete="username"  class="block py-2.5 px-0 w-full text-sm text-gray-900 bg-transparent border-0 border-b-2 border-gray-300 appearance-none dark:text-black dark:border-gray-600 dark:focus:border-blue-500 focus:outline-none focus:ring-0 focus:border-blue-600 peer" placeholder=" " />
          <label for="username" class="absolute text-sm text-gray-500 dark:text-gray-[#757575] duration-300 transform -translate-y-6 scale-75 top-3 -z-10 origin-[0] peer-focus:start-0 peer-focus:text-blue-600 peer-focus:dark:text-blue-500 peer-placeholder-shown:scale-100 peer-placeholder-shown:translate-y-0 peer-focus:scale-75 peer-focus:-translate-y-6 rtl:peer-focus:translate-x-1/4 rtl:peer-focus:left-auto">Username</label>
      </div>
      <div class="relative z-0 mt-5">
          <input required type="email" id="email" v-model="email" autocomplete="email" class="block py-2.5 px-0 w-full text-sm text-gray-900 bg-transparent border-0 border-b-2 border-gray-300 appearance-none dark:text-black dark:border-gray-600 dark:focus:border-blue-500 focus:outline-none focus:ring-0 focus:border-blue-600 peer" placeholder=" " />
          <label for="email" class="absolute text-sm text-gray-500 dark:text-gray-[#757575] duration-300 transform -translate-y-6 scale-75 top-3 -z-10 origin-[0] peer-focus:start-0 peer-focus:text-blue-600 peer-focus:dark:text-blue-500 peer-placeholder-shown:scale-100 peer-placeholder-shown:translate-y-0 peer-focus:scale-75 peer-focus:-translate-y-6 rtl:peer-focus:translate-x-1/4 rtl:peer-focus:left-auto">Email</label>
      </div>
      <div class="relative z-0 mt-5">
          <input required type="password" id="password" v-model="password" autocomplete="password"  class="block py-2.5 px-0 w-full text-sm text-gray-900 bg-transparent border-0 border-b-2 border-gray-300 appearance-none dark:text-black dark:border-gray-600 dark:focus:border-blue-500 focus:outline-none focus:ring-0 focus:border-blue-600 peer" placeholder=" " />
          <label for="password" class="absolute text-sm text-gray-500 dark:text-gray-[#757575] duration-300 transform -translate-y-6 scale-75 top-3 -z-10 origin-[0] peer-focus:start-0 peer-focus:text-blue-600 peer-focus:dark:text-blue-500 peer-placeholder-shown:scale-100 peer-placeholder-shown:translate-y-0 peer-focus:scale-75 peer-focus:-translate-y-6 rtl:peer-focus:translate-x-1/4 rtl:peer-focus:left-auto">Password</label>
      </div>
      <div class="relative z-0 mt-5">
          <input required type="password" id="confirmPassword" v-model="confirmPassword" autocomplete="password"  class="block py-2.5 px-0 w-full text-sm text-gray-900 bg-transparent border-0 border-b-2 border-gray-300 appearance-none dark:text-black dark:border-gray-600 dark:focus:border-blue-500 focus:outline-none focus:ring-0 focus:border-blue-600 peer" placeholder=" " />
          <label for="confirmPassword" class="absolute text-sm text-gray-500 dark:text-gray-[#757575] duration-300 transform -translate-y-6 scale-75 top-3 -z-10 origin-[0] peer-focus:start-0 peer-focus:text-blue-600 peer-focus:dark:text-blue-500 peer-placeholder-shown:scale-100 peer-placeholder-shown:translate-y-0 peer-focus:scale-75 peer-focus:-translate-y-6 rtl:peer-focus:translate-x-1/4 rtl:peer-focus:left-auto">Confirm Password</label>
      </div>
      <button class="w-full mt-6 shadow shadow-hover-sm cursor-pointer h-11 rounded-[3px] text-[14px] font-medium text-[#757575]" type="submit">
        <span v-if="isLoading"> 
          <div class="animate-spin inline-block size-6 border-3 border-current border-t-transparent text-[#757575] rounded-full" role="status" aria-label="loading">
            <span class="sr-only">Loading...</span>
          </div>
        </span>
        <span v-else>Sign up</span>
      </button> 
    </form>
    <router-link to="/login" class="text-[#757575] text-center text-sm underline">Already have an account?</router-link>
    <div v-if="error" class="error">{{ error }}</div>
    <div v-if="message" class="message">{{ message }}</div>
  </div>
</template>

<script setup>
import { ref } from 'vue';
import axios from 'axios';
import "@/assets/main.css"

const username = ref('');
const email = ref('');
const password = ref('');
const confirmPassword = ref('');
const error = ref('');
const message = ref('');
const isLoading = ref(false);

const register = async () => {
  // Check if passwords match
  if (password.value !== confirmPassword.value) {
    error.value = "Passwords do not match";
    return;
  }

  isLoading.value = true;

  try {
    const response = await axios.post('/v1/register', {
      username: username.value,
      password: password.value,
      email: email.value,
    });
    message.value = response.data.message;
  } catch (err) {
    error.value = err.response ? err.response.data.error : 'An error occurred';
  } finally {
    isLoading.value = false
  }
};
</script>

<style scoped>
.signup-container {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  gap: 1;
}
</style>