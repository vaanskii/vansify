<template>
  <div class="reset-password flex flex-col gap-3 max-w-[26rem] w-full mx-auto px-10 py-10">
    <h1 class="text-2xl uppercase text-[#757575]">Reset Password</h1>
    <div class="flex items-center my-4">
        <p class="flex-grow border-t border-gray-300"></p>
    </div>
    <form v-if="!passwordChanged" @submit.prevent="resetPassword">
      <div class="relative z-0 mt-5">
          <input required type="password" id="newPassword" v-model="newPassword" autocomplete="password"  class="block py-2.5 px-0 w-full text-sm text-gray-900 bg-transparent border-0 border-b-2 border-gray-300 appearance-none dark:text-black dark:border-gray-600 dark:focus:border-blue-500 focus:outline-none focus:ring-0 focus:border-blue-600 peer" placeholder=" " />
          <label for="newPassword" class="absolute text-sm text-gray-500 dark:text-gray-[#757575] duration-300 transform -translate-y-6 scale-75 top-3 -z-10 origin-[0] peer-focus:start-0 peer-focus:text-blue-600 peer-focus:dark:text-blue-500 peer-placeholder-shown:scale-100 peer-placeholder-shown:translate-y-0 peer-focus:scale-75 peer-focus:-translate-y-6 rtl:peer-focus:translate-x-1/4 rtl:peer-focus:left-auto">New password</label>
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
        <span v-else>Reset password</span>
      </button> 
    </form>
    <p v-if="message">{{ message }}</p>
  </div>
</template>

<script setup>
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import axios from 'axios';

const newPassword = ref('');
const confirmPassword = ref('');
const message = ref('');
const passwordChanged = ref(false);
const router = useRouter();
const isLoading = ref(false);

const resetPassword = async () => {
  if (newPassword.value !== confirmPassword.value) {
    message.value = 'Passwords do not match';
    return;
  }

  // Extract token from the URL
  const urlParams = new URLSearchParams(window.location.search);
  const token = urlParams.get('token');

  if (!token) {
    message.value = 'Invalid or missing token';
    return;
  }

  try {
    isLoading.value = true;
    const response = await axios.post('/v1/reset-password', {
      token: token,
      new_password: newPassword.value,
    });
    message.value = response.data.message;
    passwordChanged.value = true;
    if (response.status === 200) {
      message.value += " You will be redirected to the login page.";
      setTimeout(() => {
        router.push('/login');
      }, 5000);
    }
  } catch (error) {
    const errorMsg = error.response ? error.response.data.error : 'An error occurred';
    message.value = errorMsg;
  } finally {
    isLoading.value = false;
  }
};
</script>

<style scoped>
.reset-password {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  gap: 1;
}
</style>
