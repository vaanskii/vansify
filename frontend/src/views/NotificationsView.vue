<template>
  <div>
    <h1>Notifications</h1>
    <ul>
      <li v-for="notification in notifications" :key="notification.id">
        <span 
          :class="{ unread: !notification.is_read }"
          @click="markAsReadAndRedirect(notification.id, notification.message)"
        >
          {{ notification.message }}
        </span>
        <button @click="deleteNotification(notification.id)">Delete</button>
      </li>
    </ul>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import axios from 'axios';
import { userStore } from '@/stores/user';
import { useRouter } from 'vue-router';

const store = userStore();
const notifications = ref([]);
const router = useRouter();

const fetchNotifications = async () => {
  try {
    const response = await axios.get('/v1/notifications');
    notifications.value = response.data.notifications;
  } catch (error) {
    console.error('Error fetching notifications:', error);
  }
};

const markAsReadAndRedirect = async (notificationId, message) => {
  try {
    await axios.post(`/v1/notifications/general/mark-read/${notificationId}`);
    // Extract the username from the message
    const username = message.split(' ')[0];
    router.push(`/${username}`);
  } catch (error) {
    console.error('Error marking notification as read:', error);
  }
};

const deleteNotification = async (notificationId) => {
  try {
    const token = store.user.access;
    await axios.delete(`/v1/notifications/delete/${notificationId}`, {
      headers: { Authorization: `Bearer ${token}` }
    });
    fetchNotifications();
  } catch (error) {
    console.error('Error deleting notification:', error);
  }
};

onMounted(() => {
  fetchNotifications();
});
</script>

<style scoped>
.unread {
  font-weight: bold;
  cursor: pointer;
}
</style>
