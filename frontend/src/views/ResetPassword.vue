<template>
  <div class="reset-password">
    <h1>Reset Password</h1>
    <form v-if="!passwordChanged" @submit.prevent="resetPassword">
      <div>
        <label for="newPassword">New Password:</label>
        <input type="password" v-model="newPassword" required />
      </div>
      <div>
        <label for="confirmPassword">Confirm Password:</label>
        <input type="password" v-model="confirmPassword" required />
      </div>
      <button type="submit">Reset Password</button>
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
const router = useRouter(); // Use Vue Router

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
  }
};
</script>

<style scoped>
.reset-password {
  max-width: 400px;
  margin: 0 auto;
  padding: 1em;
  border: 1px solid #ccc;
  border-radius: 4px;
}
.reset-password h1 {
  text-align: center;
}
.reset-password form {
  display: flex;
  flex-direction: column;
}
.reset-password div {
  margin-bottom: 1em;
}
.reset-password label {
  margin-bottom: 0.5em;
  font-weight: bold;
}
.reset-password input {
  padding: 0.5em;
  border: 1px solid #ccc;
  border-radius: 4px;
}
.reset-password button {
  padding: 0.5em;
  border: none;
  border-radius: 4px;
  background-color: #007bff;
  color: white;
  cursor: pointer;
}
.reset-password button:hover {
  background-color: #0056b3;
}
.reset-password p {
  text-align: center;
  color: red;
}
</style>
