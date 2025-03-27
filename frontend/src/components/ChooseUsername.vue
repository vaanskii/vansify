<template>
    <div class="username-selection flex flex-col gap-3 max-w-[30rem] w-full mx-auto px-10 py-10">
      <h1 class="uppercase text-center text-2xl">Choose Your Username</h1>
      <form @submit.prevent="submitUsername">
        <div class="relative z-0 mt-6">
          <input required type="text" id="username" v-model="username" autocomplete="username"  class="block py-2.5 px-0 w-full text-md text-gray-900 bg-transparent border-0 border-b-2 border-gray-300 appearance-none dark:text-black dark:border-gray-600 dark:focus:border-blue-500 focus:outline-none focus:ring-0 focus:border-blue-600 peer" placeholder=" " />
          <label for="username" class="absolute text-sm text-gray-500 dark:text-gray-[#757575] duration-300 transform -translate-y-6 scale-75 top-3 -z-10 origin-[0] peer-focus:start-0 peer-focus:text-blue-600 peer-focus:dark:text-blue-500 peer-placeholder-shown:scale-100 peer-placeholder-shown:translate-y-0 peer-focus:scale-75 peer-focus:-translate-y-6 rtl:peer-focus:translate-x-1/4 rtl:peer-focus:left-auto">Username</label>
        </div>
        <button class="w-full mt-6 shadow shadow-hover-sm cursor-pointer h-11 rounded-[3px] text-[14px] font-medium text-[#757575]" type="submit">
        <span v-if="isLoading"> 
          <div class="animate-spin inline-block size-6 border-3 border-current border-t-transparent text-[#757575] rounded-full" role="status" aria-label="loading">
            <span class="sr-only">Loading...</span>
          </div>
        </span>
        <span v-else>Submit</span>
      </button> 
      </form>
      <p v-if="error">{{ error }}</p>
    </div>
  </template>
  
<script setup>
import { ref, onMounted } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import axios from 'axios';
import { userStore } from '@/stores/user';

const router = useRouter();
const route = useRoute();
const store = userStore();
const username = ref('');
const error = ref('');
const isLoading = ref(false);
const token = ref('');
const email = ref('');

onMounted(async () => {
  try {
    isLoading.value = true;
    token.value = route.query.token; 

    if (!token.value) {
      error.value = 'Missing authentication token';
      router.push('/login');
      return;
    }

    const response = await axios.post('/v1/validate-token', { token: token.value });
    email.value = response.data.email;
  } catch (err) {
    error.value = 'Invalid or expired token. Please retry the authentication process.';
    console.error('Error validating token:', err);
    router.push('/login');
  } finally {
    isLoading.value = false;
  }
});

const submitUsername = async () => {
  try {
    isLoading.value = true;

    if (!email.value) {
      error.value = 'Email not available. Please retry the authentication process.';
      router.push('/login');
      return;
    }

    // Send the username and email to the backend
    const response = await axios.post('/v1/create-user', {
      username: username.value,
      email: email.value,
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
      email: email.value,
      oauth_user: oauth_user,
    });

    // Redirect to the home page
    router.push('/');
  } catch (err) {
    error.value = err.response ? err.response.data.error : 'An error occurred';
    console.error('Error submitting username:', err);
  } finally {
    isLoading.value = false;
  }
};
</script>

  
<style scoped>
.username-selection {
  position: absolute;
  top: 30%;
  left: 50%;
  transform: translate(-50%, -50%);
  gap: 1;
}
</style>
