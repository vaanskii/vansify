import { defineStore } from "pinia";
import { ref} from 'vue';
import { useRouter } from 'vue-router';
import { userStore } from "./user";

export const useMessageUtilsStore = defineStore('messageUtilsStore', () => {
    const openMessageId = ref(null);
    const router = useRouter();
    const messages = ref([]);
    const useStore = userStore()
    const username = useStore.user.username;

    // Function to toggle message options
    const toggleMessageOptions = (messageId, event) => {
        if (event) event.stopPropagation();
        openMessageId.value = openMessageId.value === messageId ? null : messageId;
    };
  
  
    const closeMenu = (event) => {
        if (!event.target.closest('.message-options')) { 
        openMessageId.value = null;
        }
    };

    const copyMessage = (text) => {
        navigator.clipboard.writeText(text).then(() => {
        console.log('Copied:', text);
        openMessageId.value = null;
        });
    };
    
    const goToProfile = (username) => {
        router.push({ name: 'userprofile', params: { username }})
    };
    
    const formatProfilePictureUrl = (url) => {
        return url.startsWith('/') ? url.substring(1) : url;
    }

    return {
        openMessageId,
        messages,
        username,
        toggleMessageOptions,
        closeMenu,
        copyMessage,
        goToProfile,
        formatProfilePictureUrl,
    }
})