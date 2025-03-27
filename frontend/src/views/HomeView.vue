<template>
  <div v-if="!store.user.isAuthenticated">
    <LoginView />
  </div>
  <div v-else>
    <h1>Home page</h1>
  </div>
</template>

<script setup>
import { userStore } from '@/stores/user';
import LoginView from './LoginView.vue';

const store = userStore();

</script>