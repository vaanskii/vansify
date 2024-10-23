<template>
  <div>
    <form @submit.prevent="register">
      <div>
        <label for="username">Username:</label>
        <input type="text" v-model="username" id="username" required>
      </div>
      <div>
        <label for="email">Email:</label>
        <input type="email" v-model="email" id="email" required>
      </div>
      <div>
        <label for="password">Password:</label>
        <input type="password" v-model="password" id="password" required>
      </div>
      <div>
        <label for="confirmPassword">Confirm Password:</label>
        <input type="password" v-model="confirmPassword" id="confirmPassword" required>
      </div>
      <div>
        <label for="gender">Gender:</label>
        <select v-model="gender" id="gender" required>
          <option value="male">Male</option>
          <option value="female">Female</option>
        </select>
      </div>
      <button type="submit">Register</button>
    </form>
    <div v-if="error" class="error">{{ error }}</div>
    <div v-if="message" class="message">{{ message }}</div>
  </div>
</template>

<script setup>
import { ref } from 'vue';
import axios from 'axios';

const username = ref('');
const email = ref('');
const password = ref('');
const confirmPassword = ref('');
const gender = ref(''); // Add gender field
const error = ref('');
const message = ref('');

const register = async () => {
  // Check if passwords match
  if (password.value !== confirmPassword.value) {
    error.value = "Passwords do not match";
    return;
  }

  try {
    const response = await axios.post('/v1/register', {
      username: username.value,
      password: password.value,
      email: email.value,
      gender: gender.value 
    });
    message.value = response.data.message;
  } catch (err) {
    error.value = err.response ? err.response.data.error : 'An error occurred';
  }
};
</script>
