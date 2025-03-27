import { createRouter, createWebHistory } from 'vue-router';
import HomeView from '../views/HomeView.vue';
import LoginView from '../views/LoginView.vue';
import RegisterView from '../views/RegisterView.vue';
import ProfileView from '../views/ProfileView.vue';
import Chat from '../views/Chat.vue';
import ChatListView from '../views/ChatListView.vue';
import ForgotPassword from '@/views/ForgotPassword.vue';
import ResetPassword from '@/views/ResetPassword.vue';
import VerifyRegister from '@/views/VerifyRegister.vue';
import NotificationsView from '@/views/NotificationsView.vue';
import { userStore } from '@/stores/user';
import GoogleCallback from '@/components/GoogleCallback.vue';
import ChooseUsername from '@/components/ChooseUsername.vue';
import InboxWrapper from '@/views/InboxWrapper.vue';

// Function to determine if the device is mobile
const isMobile = () => window.innerWidth <= 768; // Change breakpoint as needed

const mobileRoutes = [
  {
    path: '/inbox/:chatID',
    name: 'chat',
    component: Chat,
    meta: {
      title: 'Chat',
      requiresAuth: true
    }
  },
  {
    path: '/inbox',
    name: 'chatlist',
    component: ChatListView,
    meta: {
      title: 'Chat',
      requiresAuth: true
    }
  }
];

const desktopRoutes = [
  {
    path: '/inbox',
    component: InboxWrapper,
    meta: {
      title: "Inbox",
      requiresAuth: true,
    },
    children: [
      {
        path: ':chatID',
        name: 'chat',
        component: Chat,
        meta: {
          title: 'Chat',
          requiresAuth: true
        }
      }
    ]
  }
];

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
    path: '/signup',
    name: 'signup',
    component: RegisterView,
    meta: {
      title: 'signup'
    }
  },
  {
    path: '/:username',
    name: 'userprofile',
    component: ProfileView,
  },
  {
    path: '/forgot-password',
    name: 'forgot',
    component: ForgotPassword,
    meta: {
      title: 'Forgot Password'
    }
  },
  {
    path: '/reset-password',
    name: 'reset',
    component: ResetPassword,
    meta: {
      title: 'Reset Password'
    }
  },
  {
    path: '/verify',
    name: 'verify',
    component: VerifyRegister,
    meta: {
      title: 'Verify Email'
    }
  },
  {
    path: '/notifications',
    name: 'notifications',
    component: NotificationsView,
    meta: {
      title: 'Notifications',
    }
  },
  {
    path: '/auth/google/callback',
    name: 'google auth',
    component: GoogleCallback,
    meta: {
      title: 'google callback',
    }
  },
  {
    path: '/authset',
    name: 'ChooseUsername',
    component: ChooseUsername,
    meta: {
      title: 'Auth Set',
    }
  },
  // Add routes based on device type
  ...(isMobile() ? mobileRoutes : desktopRoutes)
];

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes
});


router.beforeEach((to, from, next) => {
  const store = userStore();
  const isAuthenticated = store.user.isAuthenticated;

  const authRoutes = [
    'google auth',
    'ChooseUsername',
    'verify',
    'reset',
    'forgot',
    'signup',
    'login'
  ];

  if (isAuthenticated && authRoutes.includes(to.name)) {
    next({ name: 'home' });
  } else if (to.matched.some(record => record.meta.requiresAuth) && !isAuthenticated) {
    next({ name: 'login' });
  } else {
    if (to.name === 'userprofile') {
      document.title = `${to.params.username} • Vansify`;
    } else if (to.name === 'chat' && to.query.user) {
      document.title = `Inbox • ${to.query.user}`;
    } else {
      document.title = to.meta.title || 'Default Title';
    }
    next();
  }
});

export default router;
