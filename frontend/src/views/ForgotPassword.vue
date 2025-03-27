<template>
    <div class="forgot-password flex flex-col gap-3 max-w-[26rem] w-full mx-auto px-10 py-10">
        <h1 class="text-2xl uppercase text-[#757575]">Find your account</h1>
        <p class="text-[#757575]">Please enter your email to search for your account.</p>
        <div class="flex items-center my-4">
            <p class="flex-grow border-t border-gray-300"></p>
        </div>
        <form @submit.prevent="handleSubmit">
            <div class="relative z-0 mt-5">
                <input required type="email" id="email" v-model="email" autocomplete="email" class="block py-2.5 px-0 w-full text-sm text-gray-900 bg-transparent border-0 border-b-2 border-gray-300 appearance-none dark:text-black dark:border-gray-600 dark:focus:border-blue-500 focus:outline-none focus:ring-0 focus:border-blue-600 peer" placeholder=" " />
                <label for="email" class="absolute text-sm text-gray-500 dark:text-gray-[#757575] duration-300 transform -translate-y-6 scale-75 top-3 -z-10 origin-[0] peer-focus:start-0 peer-focus:text-blue-600 peer-focus:dark:text-blue-500 peer-placeholder-shown:scale-100 peer-placeholder-shown:translate-y-0 peer-focus:scale-75 peer-focus:-translate-y-6 rtl:peer-focus:translate-x-1/4 rtl:peer-focus:left-auto">Email</label>
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
        <p v-if="message">{{ message }}</p>
    </div>
</template>

<script setup>
import { ref } from 'vue';
import axios from 'axios';

const email = ref('');
const message = ref('');
const isLoading = ref(false);

const handleSubmit = async () => {
    try {
        isLoading.value = true
        const response = await axios.post('/v1/forgot-password', { email: email.value });
        message.value = response.data.message;
    } catch (error) {
        if (error.response && error.response.data && error.response.data.error === "OAuth users cannot reset password") {
            message.value = 'OAuth users cannot reset their password through this form. Please use your OAuth provider to manage your password.';
        } else {
            message.value = 'An error occurred. Please try again.';
        }
    } finally {
        isLoading.value = false
    }
};
</script>

<style scoped>
.forgot-password {
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    gap: 1;
}
</style>