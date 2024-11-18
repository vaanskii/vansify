<template>
  <div v-if="isAuthenticated">
    <h1 v-if="userFound">User Profile: {{ user.username }}</h1>
    <div v-if="userFound">
      <div class="image-container">
        <img v-if="imageIsLoaded" class="image" :src="resolveProfilePicture(user.profile_picture)" alt="Profile Picture"/>
        <div v-else class="lds-ellipsis"><div></div><div></div><div></div><div></div></div>
      </div>
      <p><strong>Gender:</strong> {{ user.gender }}</p>
      <p><strong>Followers:</strong> <button @click="toggleFollowers">{{ user.followers_count }}</button></p>
      <p><strong>Followings:</strong> <button @click="toggleFollowings">{{ user.followings_count }}</button></p>
      <button v-if="!isCurrentUser" @click="toggleFollow">{{ isFollowing ? 'Unfollow' : 'Follow' }}</button>
      <button v-if="!isCurrentUser && isAuthenticated" @click="handleChat">Chat</button>
      <button v-if="isCurrentUser && isAuthenticated" @click="deleteProfile">Delete Profile</button>

      <!-- Followers List -->
      <div v-if="showFollowers">
        <h3>Followers</h3>
        <ul>
          <li v-for="follower in followers" :key="follower.username">
            <img :src="resolveProfilePicture(follower.profile_picture)" alt="Profile Picture" class="small-image" />
            <span @click="goToProfile(follower.username)">{{ follower.username }}</span>
          </li>
        </ul>
      </div>

      <!-- Followings List -->
      <div v-if="showFollowings">
        <h3>Followings</h3>
        <ul>
          <li v-for="following in followings" :key="following.username">
            <img :src="resolveProfilePicture(following.profile_picture)" alt="Profile Picture" class="small-image" />
            <span @click="goToProfile(following.username)">{{ following.username }}</span>
          </li>
        </ul>
      </div>
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
import { onMounted, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import axios from 'axios';
import { userStore } from '@/stores/user';

const route = useRoute();
const router = useRouter();
const store = userStore();
const isAuthenticated = ref(store.user.isAuthenticated);
const userFound = ref(true);
const imageIsLoaded = ref(false);
const user = ref({
  id: '',
  username: '',
  profile_picture: '',
  gender: '',
  followers_count: 0,
  followings_count: 0,
});
const isCurrentUser = ref(false);
const isFollowing = ref(false);
const showFollowers = ref(false);
const showFollowings = ref(false);
const followers = ref([]);
const followings = ref([]);

function resolveProfilePicture(profilePicture) {
  if (profilePicture.startsWith('/')) {
    profilePicture = profilePicture.substring(1);
  }
  return profilePicture.startsWith('http') ? profilePicture : `${profilePicture}`;
}

// Fetch followers for current or other user
const fetchFollowers = async (username) => {
  try {
    const response = await axios.get(`/v1/followers/${username}`, {
      headers: {
        Authorization: `Bearer ${store.user.access}`,
      },
    });
    followers.value = response.data.followers;
  } catch (error) {
    console.error('Error fetching followers:', error);
  }
};

const deleteProfile = async () => {
  try {
    await axios.delete(`/v1/delete-account`, {
      headers: {
        Authorization: `Bearer ${store.user.access}`,
      },
    });
    store.removeToken();
    router.push({ path: '/login' });
  } catch (error) {
    console.error('Error deleting profile:', error);
  }
};

// Fetch followings for current or other user
const fetchFollowings = async (username) => {
  try {
    const response = await axios.get(`/v1/following/${username}`, {
      headers: {
        Authorization: `Bearer ${store.user.access}`,
      },
    });
    followings.value = response.data.followings;
  } catch (error) {
    console.error('Error fetching followings:', error);
  }
};

const toggleFollowers = async () => {
  showFollowers.value = !showFollowers.value;
  if (showFollowers.value) {
    await fetchFollowers(user.value.username);
  }
};

const toggleFollowings = async () => {
  showFollowings.value = !showFollowings.value;
  if (showFollowings.value) {
    await fetchFollowings(user.value.username);
  }
};

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
  const userData = JSON.parse(localStorage.getItem('user_data'));
  const loggedInUsername = userData.username;

  isCurrentUser.value = username === loggedInUsername;
  try {
    const response = await axios.get(`/v1/user/${username}`, {
      headers: {
        Authorization: `Bearer ${store.user.access}`,
      },
    });
    user.value = response.data;
    user.value.profile_picture = `/${user.value.profile_picture}`;
    imageIsLoaded.value = true;
    // Check follow status
    const followStatusResponse = await axios.get(`/v1/is-following/${loggedInUsername}/${username}`, {
      headers: {
        Authorization: `Bearer ${store.user.access}`,
      },
    });
    isFollowing.value = followStatusResponse.data.is_following;
  } catch (error) {
    console.error('Error fetching user details or follow status:', error);
    userFound.value = false;
  }
});

watch(route, async (newRoute) => {
  if (isAuthenticated.value) {
    const username = newRoute.params.username;
    const userData = JSON.parse(localStorage.getItem('user_data'));
    const loggedInUsername = userData.username;

    isCurrentUser.value = username === loggedInUsername;
    try {
      const response = await axios.get(`/v1/user/${username}`, {
        headers: {
          Authorization: `Bearer ${store.user.access}`,
        },
      });
      user.value = response.data;
      user.value.profile_picture = `/${user.value.profile_picture}`;
      const followStatusResponse = await axios.get(`/v1/is-following/${loggedInUsername}/${username}`, {
        headers: {
          Authorization: `Bearer ${store.user.access}`,
        },
      });
      isFollowing.value = followStatusResponse.data.is_following;
      await fetchFollowers(username);
      await fetchFollowings(username);
    } catch (error) {
      console.error('Error fetching user details or follow status:', error);
      userFound.value = false;
    }
  }
});

const goToProfile = (username) => {
  router.push({ path: `/${username}` });
};

const toggleFollow = async () => {
  try {
    if (isFollowing.value) {
      await axios.delete(`/v1/unfollow/${user.value.username}`, {
        headers: {
          Authorization: `Bearer ${store.user.access}`,
        },
      });
      updateFollowCount(false);
    } else {
      await axios.post(`/v1/follow/${user.value.username}`, {}, {
        headers: {
          Authorization: `Bearer ${store.user.access}`,
        },
      });
      updateFollowCount(true);
    }
    isFollowing.value = !isFollowing.value;
  } catch (error) {
    console.error('Error toggling follow status:', error);
  }
};

const handleChat = async () => {
  try {
    const userData = JSON.parse(localStorage.getItem('user_data'));

    // Check if chat exists
    const chatExistsResponse = await axios.get(`/v1/check-chat/${userData.username}/${user.value.username}`, {
      headers: {
        Authorization: `Bearer ${store.user.access}`,
      },
    });
    if (chatExistsResponse.data.chat_id) {
      router.push({ path: `/inbox/${chatExistsResponse.data.chat_id}`, query: { user: user.value.username } });
    } else {
      const createChatResponse = await axios.post('/v1/create-chat', { user2: user.value.username }, {
        headers: {
          Authorization: `Bearer ${store.user.access}`,
        },
      });
      const chatID = createChatResponse.data.chat_id;
      router.push({ path: `/inbox/${chatID}`, query: { user: user.value.username } });
    }
  } catch (error) {
    console.error('Error creating or retrieving chat:', error);
  }
};
</script>


<style scoped>
  .image-container {
    width: 150px;
    height: 150px;
    border-radius: 50%;
    border: 1px solid black;
    display: flex;
    justify-content: center;
    align-items: center;
  }
  .image {
    width: 149px;
    height: 149px;
    border-radius: 50%;
  }
  .small-image {
    width: 50px;
    height: 50px;
    border-radius: 50%;
    margin-right: 10px;
  }


.lds-ellipsis,
.lds-ellipsis div {
  box-sizing: border-box;
}
.lds-ellipsis {
  display: inline-block;
  position: relative;
  width: 80px;
  height: 80px;
}
.lds-ellipsis div {
  position: absolute;
  top: 33.33333px;
  width: 13.33333px;
  height: 13.33333px;
  border-radius: 50%;
  background: currentColor;
  animation-timing-function: cubic-bezier(0, 1, 1, 0);
}
.lds-ellipsis div:nth-child(1) {
  left: 8px;
  animation: lds-ellipsis1 0.6s infinite;
}
.lds-ellipsis div:nth-child(2) {
  left: 8px;
  animation: lds-ellipsis2 0.6s infinite;
}
.lds-ellipsis div:nth-child(3) {
  left: 32px;
  animation: lds-ellipsis2 0.6s infinite;
}
.lds-ellipsis div:nth-child(4) {
  left: 56px;
  animation: lds-ellipsis3 0.6s infinite;
}
@keyframes lds-ellipsis1 {
  0% {
    transform: scale(0);
  }
  100% {
    transform: scale(1);
  }
}
@keyframes lds-ellipsis3 {
  0% {
    transform: scale(1);
  }
  100% {
    transform: scale(0);
  }
}
@keyframes lds-ellipsis2 {
  0% {
    transform: translate(0, 0);
  }
  100% {
    transform: translate(24px, 0);
  }
}

</style>
