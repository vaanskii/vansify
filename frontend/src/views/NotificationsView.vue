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
import { ref, onMounted, onUnmounted } from 'vue';
import { userStore } from '@/stores/user';
import { useNotificationStore } from '@/stores/notificationStore';
import emitter from '@/eventBus';

const store = userStore();
const notificationStore = useNotificationStore();
const intervalId = ref(null);

onMounted(() => {
  if (store.user.isAuthenticated) {
    notificationStore.fetchNotifications();
    intervalId.value = setInterval(notificationStore.updateMessageTimes, 60000);
  }

  emitter.on('notification-updated', notificationStore.fetchNotifications);

});

onUnmounted(() => {
  if (intervalId.value) {
    clearInterval(intervalId.value);
  }
  emitter.off('notification-updated', notificationStore.fetchNotifications);
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
