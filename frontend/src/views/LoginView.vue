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
            <label>
            <input type="checkbox" v-model="rememberMe">
            Remember Me
            </label>
        </div>
        <button type="submit">Login</button>
        </form>
        <div v-if="error" class="error">{{ error }}</div>
        <div v-if="message" class="message">{{ message }}</div>
    </div>
</template>
  
<script setup>
import { ref } from 'vue';
import axios from 'axios';

const username = ref('');
const password = ref('');
const rememberMe = ref(false);
const error = ref('');
const message = ref('');

const login = async () => {
try {
    const response = await axios.post('http://localhost:8080/v1/login', {
    username: username.value,
    password: password.value,
    remember_me: rememberMe.value,
    });
    console.log(response.data)
    
    message.value = response.data.message;
    // Handle success - save token or navigate to another page
    const token = response.data.token;
    // Save the token to localStorage or a cookie
    localStorage.setItem('authToken', token);

    // Redirect to another page or show success message
} catch (err) {
    error.value = err.response ? err.response.data.error : 'An error occurred';
}
};
</script>