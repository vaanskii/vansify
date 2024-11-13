<template>
  <h1 v-if="notificationStore.loader">Loading...</h1>
  <div v-else>
    <h1>Notifications</h1>
    <ul v-if="notificationStore.notifications.length > 0">
      <li v-for="notification in notificationStore.notifications" :key="notification.id">
        <img :src="notification.profile_picture" alt="Profile Picture" width="30" height="30" />
        <span 
          :class="{ unread: !notification.is_read }"
          @click="notificationStore.markAsReadAndRedirect(notification.id, notification.message)"
        >
          {{ notification.message }}
        </span>
        <span class="time">{{ notificationStore.formatTime(notification.created_at) }}</span>
        <button @click="notificationStore.deleteNotification(notification.id)">Delete</button>
      </li>
    </ul>
    <p v-else>No notifications</p>
  </div>
</template>


<script setup>
import { onMounted, onUnmounted } from 'vue';
import { useRouter } from 'vue-router';
import { userStore } from '@/stores/user';
import { useNotificationStore } from '@/stores/notificationStore';
import emitter from '@/eventBus';

const store = userStore();
const notificationStore = useNotificationStore();

onMounted(() => {
  if (store.user.isAuthenticated) {
    if (notificationStore.notifications.length === 0) {
      notificationStore.fetchNotifications();
    }
    setInterval(notificationStore.updateMessageTimes, 60000);
  }
});

onUnmounted(() => {
  emitter.off('notification-updated');
});
</script>


<style scoped>
.unread {
  font-weight: bold;
  cursor: pointer;
}
.time {
  margin-left: 10px;
  color: gray;
  font-size: 0.9em;
}
</style>
