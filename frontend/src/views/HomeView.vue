<template>
  <div>
    <button v-if="store.user.isAuthenticated" @click="logout">Logout</button>
    <button v-if="store.user.isAuthenticated" @click="goToChats">My Chats</button>
    <button v-if="store.user.isAuthenticated" @click="goToProfile">My Profile</button>
    <div v-else>
      <button @click="goToLogin">Login</button>
      <button @click="goToRegister">Register</button>
    </div>
  </div>
</template>

<script setup>
import { useRouter } from 'vue-router';
import { userStore } from '@/stores/user';

const store = userStore();
const router = useRouter();

const logout = () => {
  store.removeToken();
  router.push('/login');
};

const goToChats = () => {
  router.push('/inbox');
};

const goToProfile = () => {
  router.push(`/${store.user.username}`);
};

const goToLogin = () => {
  router.push('/login');
};

const goToRegister = () => {
  router.push('/register');
};
</script>
