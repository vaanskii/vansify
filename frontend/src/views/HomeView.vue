<template>
  <button @click="logout">Logout</button>
</template>

<script setup>
import { useRouter } from 'vue-router';
import { userStore } from '@/stores/user';

const store = userStore();
const router = useRouter();

const logout = () => {
    store.removeToken();
    router.push('/login')
}
</script>