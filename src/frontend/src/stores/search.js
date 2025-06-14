import { defineStore } from 'pinia';
import api from '@/utils/api';
import router from '@/router';
import { useAuthStore } from './auth'; // 导入 Auth Store

export const useSearchStore = defineStore('search', {
  state: () => ({
    searchTerm: '',
    searchType: 'all', // Default search type
    results: [],
    isLoading: false,
    error: null,
  }),
  actions: {
    async performSearch(term, type = 'all') {
      const authStore = useAuthStore();
      if (!authStore.isAuthenticated) {
        console.warn('[SearchStore] performSearch called while not authenticated.');
        this.results = [];
        this.isLoading = false;
        // Optionally clear term and type, or show a message
        // ElMessage.error('请先登录再执行搜索');
        return;
      }

      if (!term || term.trim() === '') {
        this.results = [];
        this.searchTerm = '';
        this.searchType = type;
        // Optionally navigate to a blank search results page or clear current results
        if (router.currentRoute.value.name !== 'search-results') {
          router.push({ name: 'search-results' });
        }
        return;
      }

      this.isLoading = true;
      this.error = null;
      this.searchTerm = term;
      this.searchType = type;

      try {
        // Ensure the query parameter name matches backend expectation ('q' instead of 'term')
        const response = await api.searchAPI.search({ q: term, type });
        // The response from api.searchAPI.search, after interceptor, should be the actual data object
        // which is { results: [...] } as defined by models.GlobalSearchResponse
        this.results = response.results || []; // Access .results property
        if (router.currentRoute.value.name !== 'search-results') {
          router.push({ name: 'search-results', query: { q: this.searchTerm, type: this.searchType } });
        } else {
          // Update query params if already on the search results page
          router.replace({ query: { q: this.searchTerm, type: this.searchType } });
        }
      } catch (error) {
        console.error('Error performing search:', error);
        this.error = error.response?.data?.message || 'An unknown error occurred during search.';
        this.results = [];
      } finally {
        this.isLoading = false;
      }
    },
    clearSearch() {
      this.searchTerm = '';
      this.searchType = 'all';
      this.results = [];
      this.error = null;
      // Optionally navigate away or clear query params
      if (router.currentRoute.value.name === 'search-results') {
         router.replace({ query: {} });
      }
    },
    // Action to update search term from URL query on page load/refresh
    loadSearchTermFromQuery(query) {
      if (query.q) { // Changed from term to q
        this.searchTerm = query.q;
      }
      if (query.type) {
        this.searchType = query.type;
      }
      // If there's a search term, perform the search
      if (this.searchTerm) {
        this.performSearch(this.searchTerm, this.searchType);
      }
    }
  },
});