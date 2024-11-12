import { createApp, h } from 'vue';
import Notification from '@/components/Notification.vue';

let notificationInstance;

const notify = (message, type = 'info', duration = 3000) => {
  if (notificationInstance) {
    notificationInstance.unmount();
  }

  const container = document.createElement('div');
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
    notificationInstance.unmount();
    document.body.removeChild(container);
    notificationInstance = null;
  }, duration + 500); 
};

export default notify;
