<template>
  <div v-if="isAuthenticated">
    <h1 v-if="userFound">User Profile: {{ user.username }}</h1>
    <div v-if="userFound">
      <p><strong>Followers:</strong> {{ user.followers_count }}</p>
      <p><strong>Followings:</strong> {{ user.followings_count }}</p>
      <button v-if="!isCurrentUser" @click="toggleFollow">
        {{ isFollowing ? 'Unfollow' : 'Follow' }}
      </button>
      <button v-if="!isCurrentUser && isAuthenticated" @click="handleChat">
        Chat
      </button>
    </div>
    <div v-else>
      <p>User not found.</p>
    </div>
  </div>
  <div v-else>
    <p>Please make authorization first.</p>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import axios from 'axios';
import { userStore } from '@/stores/user';

const route = useRoute();
const router = useRouter();
const store = userStore();
const isAuthenticated = ref(store.user.isAuthenticated);
const userFound = ref(true);
const user = ref({ id: '', username: '', followers_count: 0, followings_count: 0 });
const isCurrentUser = ref(false);
const isFollowing = ref(false);

const updateFollowCount = (increment) => {
  if (increment) {
    user.value.followers_count += 1;
  } else {
    user.value.followers_count -= 1;
  }
};

onMounted(async () => {
  if (!isAuthenticated.value) return;

  const username = route.params.username;
  const loggedInUsername = localStorage.getItem('username');
  isCurrentUser.value = username === loggedInUsername;

  try {
    // Fetch user details
    const response = await axios.get(`http://localhost:8080/v1/user/${username}`);
    user.value = response.data;

    // Check follow status
    const followStatusResponse = await axios.get(`http://localhost:8080/v1/is-following/${loggedInUsername}/${username}`);
    isFollowing.value = followStatusResponse.data.is_following;
  } catch (error) {
    console.error('Error fetching user details or follow status:', error);
    userFound.value = false; 
  }
});

const toggleFollow = async () => {
  try {
    if (isFollowing.value) {
      await axios.delete(`http://localhost:8080/v1/unfollow/${user.value.username}`);
      updateFollowCount(false);
    } else {
      await axios.post(`http://localhost:8080/v1/follow/${user.value.username}`);
      updateFollowCount(true);
    }
    isFollowing.value = !isFollowing.value;
  } catch (error) {
    console.error('Error toggling follow status:', error);
  }
};

const handleChat = async () => {
  try {
    // Check if chat exists
    const chatExistsResponse = await axios.get(`http://localhost:8080/v1/check-chat/${store.user.username}/${user.value.username}`);
    
    // If chat exists, redirect to chat
    if (chatExistsResponse.data.chat_id) {
      router.push({ path: `/chat/${chatExistsResponse.data.chat_id}`, query: { user: user.value.username } });
    } else {
      // If chat does not exist, create a new chat
      const createChatResponse = await axios.post('http://localhost:8080/v1/create-chat', { user2: user.value.username }, {
        headers: { Authorization: `Bearer ${store.user.access}` }
      });
      const chatID = createChatResponse.data.chat_id;
      router.push({ path: `/chat/${chatID}`, query: { user: user.value.username } });
    }
  } catch (error) {
    console.error('Error creating or retrieving chat:', error);
  }
};
</script>
