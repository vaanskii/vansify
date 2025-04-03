import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { userStore } from '@/stores/user';
import axios from 'axios';
import "@/assets/main.css"

import App from './App.vue'
import router from './router'

axios.defaults.baseURL = import.meta.env.VITE_API_URL

const app = createApp(App)
const pinia = createPinia()

if (import.meta.env.PROD) {
  console.log = () => {};
  console.warn = () => {};
  console.error = () => {};
}

app.use(pinia)
app.use(router, axios)

const store = userStore();
store.initStore();
app.mount('#app')
