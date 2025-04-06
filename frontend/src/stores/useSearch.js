import { defineStore } from "pinia";
import { ref } from "vue";
import axios from "axios";

export const useSearchStore = defineStore("search", () => {
  const searchResults = ref([]);
  const isLoading = ref(false);

  const searchUsers = async (query) => {
    if (!query.trim()) {
      searchResults.value = [];
      return;
    }

    isLoading.value = true;
    try {
      const { data } = await axios.get(`/v1/search?q=${query}`);
      searchResults.value = data.results;
    } catch (error) {
      console.error("Search failed:", error);
      searchResults.value = [];
    } finally {
      isLoading.value = false;
    }
  };

  return { searchUsers, searchResults, isLoading };
});
