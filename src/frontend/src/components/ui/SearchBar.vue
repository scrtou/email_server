<template>
  <div class="search-bar">
    <el-input
      v-model="searchTerm"
      placeholder="全局搜索..."
      clearable
      @keyup.enter="submitSearch"
      class="search-input"
    >
      <template #prepend>
        <el-select v-model="searchType" placeholder="类型" style="width: 110px;">
          <el-option label="全部" value="all"></el-option>
          <el-option label="邮箱账户" value="email_accounts"></el-option>
          <el-option label="平台" value="platforms"></el-option>
          <el-option label="平台注册" value="platform_registrations"></el-option>
          <el-option label="服务订阅" value="service_subscriptions"></el-option>
          <!-- Add more types as needed -->
        </el-select>
      </template>
      <template #append>
        <el-button @click="submitSearch" :icon="SearchIcon" />
      </template>
    </el-input>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue';
import { useSearchStore } from '@/stores/search';
import { Search } from '@element-plus/icons-vue';
import { useRouter, useRoute } from 'vue-router';

const SearchIcon = Search; // Alias for template

const searchStore = useSearchStore();
const router = useRouter();
const route = useRoute();

const searchTerm = ref(searchStore.searchTerm);
const searchType = ref(searchStore.searchType || 'all');

// Watch for changes in the store's search term (e.g., cleared or updated from URL)
watch(() => searchStore.searchTerm, (newTerm) => {
  searchTerm.value = newTerm;
});

watch(() => searchStore.searchType, (newType) => {
  searchType.value = newType;
});

// Update store if local component state changes (e.g. user types in search bar)
watch(searchTerm, (newTerm) => {
  if (searchStore.searchTerm !== newTerm) {
    // Only update store, don't trigger search here to avoid loop
    // searchStore.searchTerm = newTerm; // Let performSearch handle this
  }
});

watch(searchType, (newType) => {
  if (searchStore.searchType !== newType) {
    // searchStore.searchType = newType; // Let performSearch handle this
  }
});


const submitSearch = () => {
  if (searchTerm.value.trim() === '' && route.name === 'search-results') {
    // If search term is cleared on search results page, clear results and update URL
    searchStore.clearSearch();
    router.replace({ name: 'search-results', query: {} });
    return;
  }
  if (searchTerm.value.trim() !== '') {
    searchStore.performSearch(searchTerm.value, searchType.value);
  }
};

// Initialize searchTerm and searchType from route query if present on component mount
// This is useful if the user navigates directly to a search URL or refreshes the page
if (route.query.term) {
  searchTerm.value = route.query.term;
}
if (route.query.type) {
  searchType.value = route.query.type;
}
// If on search results page and store is empty, populate from query
if (route.name === 'search-results' && !searchStore.searchTerm && (route.query.term || route.query.type)) {
    searchStore.loadSearchTermFromQuery(route.query);
}

</script>

<style scoped>
.search-bar {
  display: flex;
  align-items: center;
  min-width: 300px; /* Adjust as needed */
}
.search-input {
  width: 100%;
}
/* Optional: Adjust select width within input group */
.search-bar .el-select .el-input__inner {
  width: 100px; /* Or whatever fits your design */
}
</style>