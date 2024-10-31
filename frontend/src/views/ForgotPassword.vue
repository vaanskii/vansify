<template>
    <div class="forgot-password">
        <h1>Forgot Password</h1>
        <form @submit.prevent="handleSubmit">
            <div>
                <label for="email">Email:</label>
                <input type="email" id="email" v-model="email" required />
            </div>
            <button type="submit">Submit</button>
        </form>
        <p v-if="message">{{ message }}</p>
    </div>
</template>

<script setup>
import { ref } from 'vue';
import axios from 'axios';

const email = ref('');
const message = ref('');

const handleSubmit = async () => {
    try {
        const response = await axios.post('/v1/forgot-password', { email: email.value });
        message.value = response.data.message;
    } catch (error) {
        message.value = 'An error occurred. Please try again.';
    }
};
</script>

<style scoped>
.forgot-password {
    max-width: 400px;
    margin: 0 auto;
    padding: 20px;
    border: 1px solid #ccc;
    border-radius: 4px;
}

.forgot-password h1 {
    text-align: center;
}

.forgot-password form {
    display: flex;
    flex-direction: column;
}

.forgot-password label {
    margin-bottom: 5px;
}

.forgot-password input {
    margin-bottom: 10px;
    padding: 8px;
    border: 1px solid #ccc;
    border-radius: 4px;
}

.forgot-password button {
    padding: 10px;
    background-color: #007bff;
    color: white;
    border: none;
    border-radius: 4px;
    cursor: pointer;
}

.forgot-password button:hover {
    background-color: #0056b3;
}

.forgot-password p {
    text-align: center;
    color: green;
}
</style>