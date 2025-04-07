<template>
  <div v-if="isAuthenticated" class="bg-transparent pt-10 md:pt-20 flex justify-center">
    <div v-if="userFound" class="max-w-[500px] w-[90%] py-14 rounded-lg shadow-2xl">

      <!-- Profile Header -->
      <div class="text-center my-4">
        <img v-if="imageIsLoaded" class="h-32 w-32 rounded-full border-4 border-white dark:border-gray-800 mx-auto my-4"
          :src="resolveProfilePicture(user.profile_picture)" alt="Profile Picture"
        />
        
        <div class="py-2">
          <h3 class="font-bold text-2xl text-gray-800 mb-1">{{ user.username }}</h3>
        </div>
      </div>

      <!-- Followers & Followings Section -->
      <div class="px-4 py-2 text-gray-700 flex flex-col justify-center items-center gap-2">

        <!-- Followers -->
        <div @click="openModal('followers')" 
            class="cursor-pointer flex items-center justify-between w-1/2 shadow-2xl px-4 py-2 rounded-md transition-all bg-[#F0F0F0] hover:bg-[#E8E8E8]">
          <p class="font-semibold text-gray-800">Followers:</p> 
          <button class="hover:underline text-blue-600">{{ user.followers_count }}</button>
        </div>

        <!-- Followings -->
        <div @click="openModal('followings')" 
            class="cursor-pointer flex items-center justify-between w-1/2 shadow-2xl px-4 py-2 rounded-md transition-all bg-[#F0F0F0] hover:bg-[#E8E8E8]">
          <p class="font-semibold text-gray-800">Followings:</p> 
          <button class="hover:underline text-blue-600">{{ user.followings_count }}</button>
        </div>

      </div>

      <!-- Popup Modal -->
      <div v-if="showModal" class="fixed inset-0 flex items-center justify-center backdrop-blur-sm bg-black/30 z-50">
        <div class="bg-[#D4D4D4]  p-6 rounded-lg shadow-lg max-w-sm w-full h-96 overflow-y-auto relative">
          
          <!-- Close Button -->
          <button @click="closeModal" class="absolute top-2 right-2 text-gray-500 hover:text-gray-900 cursor-pointer">
            ✖
          </button>

          <h3 class="text-xl font-bold text-gray-800 mb-4">
            {{ modalType === 'followers' ? 'Followers' : 'Followings' }}
          </h3>
          
          <ul class="space-y-2">
            <li @click="goToProfile(user.username)" v-for="user in modalType === 'followers' ? followers : followings" :key="user.username" 
                class="flex items-center gap-3 hover:bg-[#BEBEBE] rounded-l-3xl rounded-r-xl cursor-pointer px-2 py-1">
              <img :src="resolveProfilePicture(user.profile_picture)" alt="Profile Picture" 
                  class="h-10 w-10 rounded-full border border-gray-300 dark:border-gray-700">
              <span class="text-gray-800">
                {{ user.username }}
              </span>
            </li>
          </ul>
        </div>
      </div>

      <div class="flex justify-center items-center pt-12">
        <!-- If it's someone else's profile -->
        <div v-if="!isCurrentUser" class="w-1/2 gap-4 flex justify-between">
          <button 
            @click="toggleFollow"
            class="rounded-full bg-blue-600 text-white cursor-pointer font-bold hover:bg-blue-800 w-40 h-10 p-2">
            {{ 
              isFollowing 
              ? 'Unfollow' 
              : (isFollowedByMe ? 'Follow Back' : 'Follow') 
            }}
          </button>

          <button v-if="isAuthenticated"
            @click="handleChat"
            class="rounded-full border-2 border-gray-400 cursor-pointer font-semibold text-black w-40 h-10 p-2">
            Message
          </button>
        </div>

        <!-- If it's the current user, show Delete Profile button -->
        <button v-if="isCurrentUser && isAuthenticated"
          @click="showDeleteModal = true"
          class="rounded-full bg-red-600 text-white cursor-pointer font-bold hover:bg-red-700 w-[90%] px-4 py-2">
          Delete Profile
        </button>
      </div>

      <!-- Delete Confirmation Modal -->
      <div v-if="showDeleteModal" class="fixed inset-0 flex items-center justify-center backdrop-blur-sm bg-black/30 z-50">
        <div class="bg-white p-6 rounded-lg shadow-lg max-w-sm w-[90%] relative ">

          <h3 class="text-xl font-bold text-gray-800 mb-4">Are you sure?</h3>
          <p class="text-gray-600 mb-6">Deleting your profile cannot be undone.</p>

          <div class="flex justify-between">
            <button @click="confirmDeleteProfile"
                    class="rounded-full bg-red-600 cursor-pointer text-white font-bold hover:bg-red-700 px-4 py-2">
              Delete
            </button>
            <button @click="showDeleteModal = false"
                    class="rounded-full border-2 cursor-pointer border-gray-400 font-semibold text-black px-4 py-2">
              Cancel
            </button>
          </div>

        </div>
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
const showDeleteModal = ref(false);
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
const isFollowedByMe = ref(false);
const followers = ref([]);
const followings = ref([]);

const showModal = ref(false);
const modalType = ref("");

const openModal = async (type) => {
  modalType.value = type;
  showModal.value = true;

  if (type === "followers") {
    await fetchFollowers(user.value.username);
  } else {
    await fetchFollowings(user.value.username);
  }
};

const closeModal = () => {
  showModal.value = false;
};

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

const confirmDeleteProfile = async () => {
  try {
    await deleteProfile(); 
    showDeleteModal.value = false;
  } catch (error) {
    console.error("Error deleting profile:", error);
  }
}

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

onMounted(async () => {
  if (!isAuthenticated.value) return;
  const username = route.params.username;
  const userData = JSON.parse(localStorage.getItem('user_data'));
  const loggedInUsername = userData.username;

  isCurrentUser.value = username === loggedInUsername;
  
  try {
    const response = await axios.get(`/v1/user/${username}`, {
      headers: { Authorization: `Bearer ${store.user.access}` },
    });
    user.value = response.data;
    user.value.profile_picture = `/${user.value.profile_picture}`;
    imageIsLoaded.value = true;

    // ✅ Only check `isFollowedByMe` if viewing someone else's profile
    if (!isCurrentUser.value) {
      const followStatusResponse = await axios.get(`/v1/is-following/${loggedInUsername}/${username}`, {
        headers: { Authorization: `Bearer ${store.user.access}` },
      });

      isFollowing.value = followStatusResponse.data.is_following;
      isFollowedByMe.value = followStatusResponse.data.is_followed_by;
    }
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
  closeModal();
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

const updateFollowCount = (increment) => {
  if (increment) {
    user.value.followers_count += 1;
  } else {
    user.value.followers_count -= 1;
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

</style>
