import { defineStore } from 'pinia';
import { ref } from 'vue';
import axios from 'axios';
import { userStore } from '@/stores/user';
import emitter from '@/eventBus';

export const useNotificationStore = defineStore('notificationStore', () => {
  const store = userStore();
  const notifications = ref([]);
  const loader = ref(true);

  const fetchNotifications = async () => {
    try {
      const response = await axios.get('/v1/notifications', {
        headers: {
          Authorization: `Bearer ${store.user.access}`
        }
      });
      notifications.value = response.data.notifications || [];
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
      await axios.post(`/v1/notifications/general/mark-read/${notificationId}`, null, {
        headers: {
          Authorization: `Bearer ${store.user.access}`
        }
      });
      const username = message.split(' ')[0];
      notifications.value = notifications.value.map(notification => 
        notification.id === notificationId ? { ...notification, is_read: true } : notification
      );
      emitter.emit('notification-updated');
      router.push(`/${username}`);
    } catch (error) {
      console.error('Error marking notification as read:', error);
    }
  };

  const deleteNotification = async (notificationId) => {
    try {
      await axios.delete(`/v1/notifications/delete/${notificationId}`, {
        headers: {
          Authorization: `Bearer ${store.user.access}`
        }
      });
      notifications.value = notifications.value.filter(notification => notification.id !== notificationId);
      emitter.emit('notification-updated');
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

  return {
    notifications,
    loader,
    fetchNotifications,
    updateMessageTimes,
    markAsReadAndRedirect,
    deleteNotification,
    formatTime
  };
});
