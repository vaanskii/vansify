<template>
  <div v-if="!store.user.isAuthenticated">
    <unauthenticatedNavigation />
    <unauthenticatedMain />
  </div>
  <div v-else>
    <h1>Home page</h1>
  </div>
</template>

<script setup>
import UnauthenticatedMain from '@/components/UnauthenticatedMain.vue';
import UnauthenticatedNavigation from '@/components/UnauthenticatedNavigation.vue';
import { userStore } from '@/stores/user';

const store = userStore();

</script>