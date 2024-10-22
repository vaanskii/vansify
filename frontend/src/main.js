import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { userStore } from '@/stores/user';

import App from './App.vue'
import router from './router'

const app = createApp(App)

app.use(createPinia())
app.use(router)

app.mount('#app')
const store = userStore();
store.initStore();