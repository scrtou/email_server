import { defineStore } from 'pinia';
import { platformAPI } from '@/utils/api';
import { ElMessage } from 'element-plus';
import { useAuthStore } from './auth'; // 导入 Auth Store

// First, define platformAPI within api.js or ensure it's added there.
// For now, assuming it will be added to api.js like emailAccountAPI:
/*
// In api.js:
export const platformAPI = {
  getAll: (params = {}) => api.get('/platforms', { params }),
  getById: (id) => api.get(`/platforms/${id}`),
  create: (data) => api.post('/platforms', data),
  update: (id, data) => api.put(`/platforms/${id}`, data),
  delete: (id) => api.delete(`/platforms/${id}`),
};
*/

export const usePlatformStore = defineStore('platform', {
  state: () => ({
    platforms: [],
    currentPlatform: null,
    loading: false,
    error: null,
    pagination: {
      currentPage: 1,
      pageSize: 8,
      totalItems: 0,
    },
    sort: { // New state for sorting
      orderBy: 'name', // Default sort for platforms
      sortDirection: 'asc',
    },
    filters: { // New state for filters
      nameSearch: '',
    },
  }),
  actions: {
    setFilter(filterName, value) {
      if (Object.prototype.hasOwnProperty.call(this.filters, filterName)) {
        this.filters[filterName] = value;
        this.pagination.currentPage = 1; // Reset to first page
        // Trigger fetch, assuming the component will call it or we call it here
        // For now, let component trigger fetch or call directly if store manages all fetches
        // this.fetchPlatforms(); // Example: if store should auto-fetch
      }
    },
    clearFilters() {
      this.filters.nameSearch = '';
      this.pagination.currentPage = 1;
      // this.fetchPlatforms(); // Example: if store should auto-fetch
    },
    async fetchPlatforms(page = this.pagination.currentPage, pageSize = this.pagination.pageSize, sortOptions = {}, filterOptions = {}) {
      const authStore = useAuthStore();
      if (!authStore.isAuthenticated) {
        console.warn('[PlatformStore] fetchPlatforms called while not authenticated.');
        this.platforms = [];
        this.pagination.totalItems = 0;
        this.loading = false;
        return;
      }

      this.loading = true;
      this.error = null;

      // Update store's filter state if new filters are passed and different
      if (filterOptions.nameSearch !== undefined && this.filters.nameSearch !== filterOptions.nameSearch) {
        this.filters.nameSearch = filterOptions.nameSearch;
        // If filters are externally updated, reset to page 1
        if (page !== 1) page = 1;
        this.pagination.currentPage = page;
      }

      const orderBy = sortOptions.orderBy || this.sort.orderBy;
      const sortDirection = sortOptions.sortDirection || this.sort.sortDirection;

      // Update sort state if new options are provided
      if (sortOptions.orderBy) this.sort.orderBy = sortOptions.orderBy;
      if (sortOptions.sortDirection) this.sort.sortDirection = sortOptions.sortDirection;

      try {
        const params = {
          page,
          pageSize,
          orderBy: orderBy,
          sortDirection: sortDirection,
          name: this.filters.nameSearch || undefined, // Add name search filter
        };
        const result = await platformAPI.getAll(params);
        if (result && result.data) {
          this.platforms = result.data;
          if (result.meta) {
            this.pagination.currentPage = result.meta.current_page;
            this.pagination.pageSize = result.meta.page_size;
            this.pagination.totalItems = result.meta.total_items;
          } else {
            this.pagination = { currentPage: page, pageSize: pageSize, totalItems: result.data.length };
          }
        } else {
          this.platforms = [];
          this.pagination = { currentPage: 1, pageSize: 10, totalItems: 0 };
        }
      } catch (err) {
        this.error = err.message || '获取平台列表失败';
        ElMessage.error(this.error);
        this.platforms = [];
        this.pagination = { currentPage: 1, pageSize: 10, totalItems: 0 };
      } finally {
        this.loading = false;
      }
    },
    async fetchPlatformById(id) {
      const authStore = useAuthStore();
      if (!authStore.isAuthenticated) {
        console.warn('[PlatformStore] fetchPlatformById called while not authenticated.');
        this.currentPlatform = null;
        this.loading = false;
        return null;
      }

      this.loading = true;
      this.error = null;
      this.currentPlatform = null;
      try {
        const data = await platformAPI.getById(id);
        this.currentPlatform = data;
        return data;
      } catch (err) {
        this.error = err.message || '获取平台详情失败';
        ElMessage.error(this.error);
        return null;
      } finally {
        this.loading = false;
      }
    },
    async createPlatform(data) {
      const authStore = useAuthStore();
      if (!authStore.isAuthenticated) {
        console.warn('[PlatformStore] createPlatform called while not authenticated.');
        ElMessage.error('请先登录再创建平台');
        this.loading = false;
        return null;
      }

      this.loading = true;
      this.error = null;
      try {
        const createdData = await platformAPI.create(data);
        ElMessage.success('平台创建成功');
        await this.fetchPlatforms(this.pagination.currentPage, this.pagination.pageSize);
        return createdData;
      } catch (err) {
        const errorMessage = err.message || '创建平台失败';
        if (errorMessage.includes('UNIQUE constraint failed') && errorMessage.includes('platforms.name')) {
          this.error = '该平台名称已被使用。请输入一个不同的平台名称。';
          ElMessage.error(this.error);
        } else {
          this.error = errorMessage;
          ElMessage.error(this.error);
        }
        return null;
      } finally {
        this.loading = false;
      }
    },
    async updatePlatform(id, data) {
      const authStore = useAuthStore();
      if (!authStore.isAuthenticated) {
        console.warn('[PlatformStore] updatePlatform called while not authenticated.');
        ElMessage.error('请先登录再更新平台');
        this.loading = false;
        return null;
      }

      this.loading = true;
      this.error = null;
      try {
        const updatedData = await platformAPI.update(id, data);
        ElMessage.success('平台更新成功');
        await this.fetchPlatforms(this.pagination.currentPage, this.pagination.pageSize);
        if (this.currentPlatform && this.currentPlatform.id === id) {
          this.currentPlatform = updatedData;
        }
        return updatedData;
      } catch (err) {
        this.error = err.message || '更新平台失败';
        ElMessage.error(this.error);
        return null;
      } finally {
        this.loading = false;
      }
    },
    async deletePlatform(id) {
      const authStore = useAuthStore();
      if (!authStore.isAuthenticated) {
        console.warn('[PlatformStore] deletePlatform called while not authenticated.');
        ElMessage.error('请先登录再删除平台');
        this.loading = false;
        return false;
      }

      this.loading = true;
      this.error = null;
      try {
        await platformAPI.delete(id);
        ElMessage.success('平台删除成功');
        await this.fetchPlatforms(this.pagination.currentPage, this.pagination.pageSize);
        return true;
      } catch (err) {
        this.error = err.message || '删除平台失败';
        ElMessage.error(this.error);
        return false;
      } finally {
        this.loading = false;
      }
    },
    setCurrentPlatform(platform) {
        this.currentPlatform = platform;
    },
    async fetchAssociatedEmailRegistrations(platformId, params = { page: 1, pageSize: 5 }) {
      const authStore = useAuthStore();
      if (!authStore.isAuthenticated) {
        console.warn('[PlatformStore] fetchAssociatedEmailRegistrations called while not authenticated.');
        this.loading = false;
        // Return empty structure consistent with success case on error
        return { data: [], meta: { total_records: 0, current_page: 1, page_size: params.pageSize } };
      }

      this.loading = true;
      this.error = null;
      try {
        // api.js interceptor returns { data: [...], meta: {...} } for paginated responses
        const result = await platformAPI.getAssociatedEmailRegistrations(platformId, params);
        return result || { data: [], meta: { total_records: 0, current_page: 1, page_size: params.pageSize } };
      } catch (err) {
        this.error = err.message || '获取关联邮箱注册信息失败';
        ElMessage.error(this.error);
        return { data: [], meta: { total_records: 0, current_page: 1, page_size: params.pageSize } }; // Return empty structure on error
      } finally {
        this.loading = false;
      }
    }
  },
});