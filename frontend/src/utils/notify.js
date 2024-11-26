import { createApp, h } from 'vue';
import Notification from '@/components/Notification.vue';

let notificationInstance;
let container;

const notify = (message, type = 'info', duration = 3000) => {
  if (notificationInstance) {
    notificationInstance.unmount();
    if (container && container.parentNode === document.body) {
      document.body.removeChild(container);
    }
    notificationInstance = null;
    container = null;
  }

  // Create a new container for the notification
  container = document.createElement('div');
  container.className = 'notification-container';
  document.body.appendChild(container);

  notificationInstance = createApp({
    render() {
      return h(Notification, {
        message,
        type,
        duration,
      });
    },
  });

  notificationInstance.mount(container);

  setTimeout(() => {
    if (notificationInstance) {
      notificationInstance.unmount();
      if (container && container.parentNode === document.body) {
        document.body.removeChild(container);
      }
      notificationInstance = null;
      container = null;
    }
  }, duration + 500); 
};

export default notify;
