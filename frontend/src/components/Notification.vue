<template>
    <div class="notification-container">
        <transition name="slide-down">
            <div class="notification" :class="type" v-if="visible">
                <span>{{ message }}</span>
                <!-- <button @click="close">x</button> -->
            </div>
        </transition>
    </div>
</template>

<script setup>
import { ref } from 'vue';

const props = defineProps({
message: String,
type: {
    type: String,
    default: 'info',
    validator: value => ['success', 'error', 'info', 'warning'].includes(value),
},
duration: {
    type: Number,
    default: 3000,
},
});

const visible = ref(true);

// const close = () => {
// visible.value = false;
// };

setTimeout(() => {
visible.value = false;
}, props.duration);
</script>

<style scoped>
.notification-container{
    display: flex;
    justify-content: center;
    align-items: center;
}

@keyframes slide-down {
    0% {
        transform: translateY(-100%);
        opacity: 0;
    }
    10% {
        transform: translateY(0);
        opacity: 1;
    }
    90% {
        transform: translateY(0);
        opacity: 1;
    }
    100% {
        transform: translateY(-100%);
        opacity: 0;
    }
}

.notification {
    display: flex;
    justify-content: center;
    align-items: center;
    padding: 15px;
    width: 200px;
    margin-bottom: 10px;
    border-radius: 5px;
    color: white;
    font-weight: bold;
    animation: slide-down 3.5s ease forwards;
}

.notification.success {
    background-color: #4caf50;
}

.notification.error {
    background-color: #f44336;
}

.notification.info {
    background-color: #2196f3;
}

.notification.warning {
    background-color: #ff9800;
}
</style>
  