import { createRouter, createWebHistory } from 'vue-router';
import HomeView from '../views/HomeView.vue';
import LoginView from '../views/LoginView.vue';
import RegisterView from '../views/RegisterView.vue';
import ProfileView from '../views/ProfileView.vue';
import Chat from '../views/Chat.vue';
import ChatListView from '../views/ChatListView.vue';

const routes = [
  {
    path: '/',
    name: 'home',
    component: HomeView,
    meta: {
      title: 'Vansify'
    }
  },
  {
    path: '/login',
    name: 'login',
    component: LoginView,
    meta: {
      title: 'Login'
    }
  },
  {
    path: '/register',
    name: 'register',
    component: RegisterView,
    meta: {
      title: 'Register'
    }
  },
  {
    path: '/:username',
    name: 'userprofile',
    component: ProfileView,
  },
  {
    path: '/inbox/:chatID',
    name: 'chat',
    component: Chat,
    meta: {
      title: 'Chat'
    }
  },
  {
    path: '/inbox',
    name: 'chatlist',
    component: ChatListView,
    meta: {
      title: 'Chat'
    }
  },
];

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes
});

// Navigation guard to update document title
router.beforeEach((to, from, next) => {
  if (to.name === 'userprofile') {
    document.title = `${to.params.username} • Vansify`;
  } else {
    document.title = to.meta.title || 'Default Title';
  }
  next();
});

export default router;
