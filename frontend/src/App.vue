<template>
  <router-view />
</template>

<script setup>
import { onMounted } from 'vue';
import { userStore } from '@/stores/user';
import { RouterView } from 'vue-router'

const store = userStore();

onMounted(() => {
  store.initStore();
});
</script>
