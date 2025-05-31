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
if (route.query.q) { // Changed from term to q
  searchTerm.value = route.query.q;
}
if (route.query.type) {
  searchType.value = route.query.type;
}
// If on search results page and store is empty, populate from query
if (route.name === 'search-results' && !searchStore.searchTerm && (route.query.q || route.query.type)) { // Changed from term to q
    searchStore.loadSearchTermFromQuery(route.query);
}

</script>

<style scoped>
.search-bar {
  display: flex;
  align-items: center;
  min-width: 400px;
  max-width: 600px;
  flex-grow: 1;
}

.search-input {
  width: 100%;
  border-radius: var(--radius-lg);
  overflow: hidden;
  box-shadow: var(--shadow-sm);
  transition: all var(--transition-base);
}

.search-input:hover {
  box-shadow: var(--shadow-md);
}

.search-input:focus-within {
  box-shadow: var(--shadow-lg);
  transform: translateY(-1px);
}

/* Enhanced input styling */
:deep(.el-input__wrapper) {
  border-radius: var(--radius-lg);
  border: 2px solid var(--color-gray-200);
  background: rgba(255, 255, 255, 0.9);
  backdrop-filter: blur(10px);
  transition: all var(--transition-base);
}

:deep(.el-input__wrapper:hover) {
  border-color: var(--color-primary-300);
  background: rgba(255, 255, 255, 0.95);
}

:deep(.el-input__wrapper.is-focus) {
  border-color: var(--color-primary-500);
  background: rgba(255, 255, 255, 1);
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

:deep(.el-input__inner) {
  font-size: var(--text-sm);
  font-weight: var(--font-medium);
  color: var(--color-gray-700);
}

:deep(.el-input__inner::placeholder) {
  color: var(--color-gray-400);
  font-weight: var(--font-normal);
}

/* Enhanced select styling */
:deep(.el-input-group__prepend) {
  background: linear-gradient(135deg, var(--color-gray-50), var(--color-gray-100));
  border: none;
  border-right: 1px solid var(--color-gray-200);
}

:deep(.el-input-group__append) {
  background: linear-gradient(135deg, var(--color-primary-500), var(--color-primary-600));
  border: none;
  padding: 0;
}

:deep(.el-input-group__append .el-button) {
  background: transparent;
  border: none;
  color: white;
  font-weight: var(--font-medium);
  transition: all var(--transition-base);
}

:deep(.el-input-group__append .el-button:hover) {
  background: rgba(255, 255, 255, 0.1);
  transform: scale(1.05);
}

/* Select dropdown styling */
:deep(.el-select .el-input__wrapper) {
  border: none;
  background: transparent;
  box-shadow: none;
}

:deep(.el-select .el-input__inner) {
  font-size: var(--text-xs);
  font-weight: var(--font-semibold);
  color: var(--color-gray-600);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

/* Responsive design */
@media (max-width: 768px) {
  .search-bar {
    min-width: 250px;
    max-width: 100%;
  }

  :deep(.el-input-group__prepend) {
    padding: 0 var(--space-2);
  }

  :deep(.el-select .el-input__inner) {
    font-size: var(--text-xs);
  }
}
</style>