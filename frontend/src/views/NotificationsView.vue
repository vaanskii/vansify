<template>
  <div class="max-w-lg mx-auto p-8 bg-white rounded-xl shadow-lg">
    
    <!-- Loader -->
    <h1 v-if="notificationStore.loader" class="text-lg font-semibold text-gray-700 text-center">Loading...</h1>

    <!-- Notifications List -->
    <div v-else>
      <h1 class="text-2xl font-bold text-gray-800 mb-6">Notifications</h1>
      
      <ul v-if="notificationStore.notifications.length > 0" class="space-y-6">
        <li v-for="notification in notificationStore.notifications" :key="notification.id" 
            class="flex items-center justify-between bg-gray-100 p-3 rounded-lg shadow-md hover:bg-gray-200 transition gap-x-6">
          
          <!-- Profile Picture -->
          <img :src="notification.profile_picture" alt="Profile Picture" 
              class="w-12 h-12 rounded-full border border-gray-400">

          <!-- Notification Message -->
          <span 
            :class="{ 'font-bold text-blue-600': !notification.is_read }"
            @click="notificationStore.markAsReadAndRedirect(notification.id, notification.message)"
            class="cursor-pointer flex-grow text-gray-800 hover:underline text-lg"
          >
            {{ notification.message }}
          </span>

          <!-- Time -->
          <span class="text-sm text-gray-500">{{ notificationStore.formatTimeAgo(notification.created_at) }}</span>

          <!-- Delete Button -->
          <button @click="notificationStore.deleteNotification(notification.id)" 
                  class="bg-red-600 text-white px-4 py-2 rounded-md hover:bg-red-700 transition">
            Delete
          </button>
        </li>
      </ul>

      <!-- No Notifications -->
      <p v-else class="text-gray-500 text-center text-lg">No notifications</p>
    </div>
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
    intervalId.value = setInterval(notificationStore.updateNotificationTimes, 60000);
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
