<template>
  <h1 v-if="loader">Loading...</h1>
  <div v-else>
    <h1>Notifications</h1>
    <ul v-if="notifications.length > 0">
      <li v-for="notification in notifications" :key="notification.id">
        <img :src="notification.profile_picture" alt="Profile Picture" width="30" height="30" />
        <span 
          :class="{ unread: !notification.is_read }"
          @click="markAsReadAndRedirect(notification.id, notification.message)"
        >
          {{ notification.message }}
        </span>
        <span class="time">{{ formatTime(notification.created_at) }}</span>
        <button @click="deleteNotification(notification.id)">Delete</button>
      </li>
    </ul>
    <p v-else>No noticiations</p>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import axios from 'axios';
import { userStore } from '@/stores/user';
import { useRouter } from 'vue-router';
import { addData, getData } from '@/utils/notifDB';

const store = userStore();
const notifications = ref([]);
const router = useRouter();
const loader = ref(true);

const fetchNotifications = async () => {
  try {
    const response = await axios.get('/v1/notifications');
    notifications.value = response.data.notifications || [];
    notifications.value.forEach(notification => {
      console.log(notification.created_at)
    })
    addData('notifications', { id: 'notificationList', notification_list: notifications.value });
    loader.value = false;
  } catch (error) {
    console.error('Error fetching notifications:', error);
  }
};

const updateMessageTimes = () => {
  notifications.value = notifications.value.map(notification => ({
    ...notification,
    time: formatTime(notification.created_at) 
  }));
};

const markAsReadAndRedirect = async (notificationId, message) => {
  try {
    await axios.post(`/v1/notifications/general/mark-read/${notificationId}`, null)
    const username = message.split(' ')[0];
    router.push(`/${username}`);
  } catch (error) {
    console.error('Error marking notification as read:', error);
  }
};

const deleteNotification = async (notificationId) => {
  try {
    await axios.delete(`/v1/notifications/delete/${notificationId}`);
    notifications.value = notifications.value.filter(notification => notification.id !== notificationId);
    addData('notifications', { id: 'notificationList', notification_list: notifications.value });
  } catch (error) {
    console.error('Error deleting notification:', error);
  }
};

const formatTime = (timestamp) => {
  const timeDiff = Math.floor((Date.now() - new Date(timestamp)) / 1000);
  if (timeDiff < 60) return 'Just now';
  const minutes = Math.floor(timeDiff / 60);
  if (minutes < 60) return `${minutes} min ago`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours} ${hours === 1 ? 'hour' : 'hours'} ago`;
  const days = Math.floor(hours / 24);
  return `${days} ${days === 1 ? 'day' : 'days'} ago`;
};

onMounted(async () => {
  const storedNotifications = await getData('notifications', 'notificationList');
  if (storedNotifications) {
    notifications.value = storedNotifications.notification_list;
    loader.value = false;
  }
  
  if (store.user.isAuthenticated) {
    fetchNotifications();
    setInterval(updateMessageTimes, 60000);
  }
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
