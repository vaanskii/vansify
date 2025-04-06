<template>
    <div class="flex items-center justify-center flex-col">
      <div class="max-w-[400px] w-[90%]">
        <h1 class="font-bold uppercase text-2xl">Users</h1>
        <div v-if="searchResults.length" class="flex items-center py-4">
          <ul class="w-full">
          <li v-for="user in searchResults" :key="user.id">
            <div class="flex flex-row items-center justify-between">
              <div class="flex flex-row items-center gap-2">
                <img :src="user.profile_picture" alt="Profile" class="profile-img w-10 h-10 rounded-full cursor-pointer"/>
                <p @click="goToProfile(user.username)" class="hover:underline cursor-pointer">{{ user.username }}</p>
              </div>
              <button @click="goToProfile(user.username)" type="button" class="text-white cursor-pointer bg-blue-700 hover:bg-blue-800 focus:ring-4 focus:ring-blue-300 font-medium rounded-lg text-sm px-5 py-2.5 me-2 mb-2 dark:bg-blue-600 dark:hover:bg-blue-700 focus:outline-none dark:focus:ring-blue-800">Profile</button>
            </div>
          </li>
        </ul>
        </div>
        <p v-else>No results found.</p>
      </div>
    </div>
</template>

<script setup>
import { ref, watch, onMounted } from "vue";
import { useRoute, useRouter } from "vue-router";
import axios from "axios";

const route = useRoute();
const router = useRouter();
const searchQuery = ref(route.query.query);
const searchResults = ref([]);

const fetchSearchResults = async () => {
  if (!searchQuery.value) return;

  try {
    const { data } = await axios.get(`/v1/search?q=${searchQuery.value}`);
    searchResults.value = data.results;
  } catch (error) {
    console.error("Search failed:", error);
  }
};

// ✅ Watch for changes in the search query and fetch new results
watch(() => route.query.query, (newQuery) => {
  searchQuery.value = newQuery;
  fetchSearchResults();
});

const goToProfile = (username) => {
    router.push({ name: 'userprofile', params: { username }})
};

onMounted(fetchSearchResults);
</script>

