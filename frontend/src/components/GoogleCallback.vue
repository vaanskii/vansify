<template>
  <div>
    <p>Logging in with Google...</p>
  </div>
</template>

<script setup>
import { onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { userStore } from '@/stores/user';
import { useActiveUsersStore } from '@/stores/activeUsers';

const router = useRouter();
const store = userStore();
const activeUsersStore = useActiveUsersStore();

onMounted(() => {
  try {
    const query = router.currentRoute.value.query;
    const email = query.email;
    const username = query.username;
    const accessToken = query.access_token;
    const refreshToken = query.refresh_token;
    const id = query.id;
    const oauthUser = query.oauth_user;

    // Handle success - save user info
    store.setToken({
      access: accessToken,
      refresh: refreshToken,
      id: id,
      username: username,
      email: email,
      oauth_user: oauthUser,
    });

    // Establish WebSocket connection for active users
    activeUsersStore.connectWebSocket();

    // Redirect after successful login
    router.push('/');
  } catch (err) {
    console.error('Error during Google OAuth callback:', err);
    router.push('/login');
  }
});
</script>
