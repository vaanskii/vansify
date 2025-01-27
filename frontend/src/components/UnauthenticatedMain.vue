<template>
    <div>
        <button @click="loginWithGoogle" type="button" class="login-with-google-btn" >
            Sign in with Google
        </button>
        <button @click="signinWithUsername">Sign in with username</button>
        <div v-if="error" class="error">{{ error }}</div>
        <div v-if="message" class="message">{{ message }}</div>
    </div>
</template>
  
<script setup>
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import "@/assets/buttons.css"

const error = ref('');
const message = ref('');
const router = useRouter();

const signinWithUsername = () => {
router.push('/login')
}

const loginWithGoogle = () => {
console.log('Login with Google called');
const apiUrl = import.meta.env.VITE_API_URL;
window.location.href = `${apiUrl}/v1/auth/google`;
};
</script>

<style scoped>
</style>