import { defineStore } from 'pinia';
import { platformRegistrationAPI } from '@/utils/api';
import { ElMessage } from 'element-plus';
import { useAuthStore } from './auth'; // 导入 Auth Store

// Assuming platformRegistrationAPI will be added to api.js:
/*
// In api.js:
export const platformRegistrationAPI = {
  getAll: (params = {}) => api.get('/platform-registrations', { params }),
  getById: (id) => api.get(`/platform-registrations/${id}`),
  create: (data) => api.post('/platform-registrations', data),
  update: (id, data) => api.put(`/platform-registrations/${id}`, data),
  delete: (id) => api.delete(`/platform-registrations/${id}`),
};
*/

export const usePlatformRegistrationStore = defineStore('platformRegistration', {
  state: () => ({
    platformRegistrations: [],
    currentPlatformRegistration: null,
// currentPlatformRegistration will include phone_number when fetched/set
    loading: false,
    error: null,
    pagination: {
      currentPage: 1,
      pageSize: 8,
      totalItems: 0,
    },
    sort: { // New state for sorting
      orderBy: 'created_at', // Default sort
      sortDirection: 'desc',
    },
    filters: { // New state for filters
      email_account_id: null,
      platform_id: null,
      login_username: '', // 添加用户名筛选
    }
  }),
  actions: {
    // Action to update filter values
    setFilter(filterName, value) {
      if (Object.prototype.hasOwnProperty.call(this.filters, filterName)) {
        this.filters[filterName] = value;
        this.pagination.currentPage = 1; // Reset to first page
        this.fetchPlatformRegistrations(); // Re-fetch data
      }
    },
    // Action to clear all filters
    clearFilters() {
      this.filters.email_account_id = null;
      this.filters.platform_id = null;
      this.filters.login_username = ''; // 清空用户名筛选
      this.pagination.currentPage = 1;
      this.fetchPlatformRegistrations();
    },
    async fetchPlatformRegistrations(
        page = this.pagination.currentPage,
        pageSize = this.pagination.pageSize,
        sortOptions = {},
        // filterOptions = {} // filterOptions is not used as a parameter anymore, filters are taken from state
    ) {
      const authStore = useAuthStore();
      if (!authStore.isAuthenticated) {
        console.warn('[PlatformRegistrationStore] fetchPlatformRegistrations called while not authenticated.');
        this.platformRegistrations = [];
        this.pagination.totalItems = 0;
        this.loading = false;
        return;
      }

      this.loading = true;
      this.error = null;
      
      const orderBy = sortOptions.orderBy || this.sort.orderBy;
      const sortDirection = sortOptions.sortDirection || this.sort.sortDirection;

      // Update sort state
      if (sortOptions.orderBy) this.sort.orderBy = sortOptions.orderBy;
      if (sortOptions.sortDirection) this.sort.sortDirection = sortOptions.sortDirection;

      // Filter state is updated by setFilter or directly via v-model binding in components.
      // No need to pass filterOptions to this action anymore.
      // The action will use the current state of this.filters.
 
      const apiParams = {
        page,
        pageSize,
        orderBy,
        sortDirection,
        email_account_id: this.filters.email_account_id || undefined,
        platform_id: this.filters.platform_id || undefined,
        username: this.filters.login_username || undefined, // Changed from login_username to username
      };
 
      try {
        const result = await platformRegistrationAPI.getAll(apiParams);
        if (result && result.data) {
          this.platformRegistrations = result.data;
          if (result.meta) {
            this.pagination.currentPage = result.meta.current_page;
            this.pagination.pageSize = result.meta.page_size;
            this.pagination.totalItems = result.meta.total_items;
          } else {
            this.pagination = { currentPage: page, pageSize: pageSize, totalItems: result.data.length };
          }
        } else {
          this.platformRegistrations = [];
          this.pagination = { currentPage: 1, pageSize: 10, totalItems: 0 };
        }
      } catch (err) {
        this.error = err.message || '获取平台注册列表失败';
        ElMessage.error(this.error);
        this.platformRegistrations = [];
        this.pagination = { currentPage: 1, pageSize: 10, totalItems: 0 };
      } finally {
        this.loading = false;
      }
    },
    async fetchPlatformRegistrationById(id) {
      const authStore = useAuthStore();
      if (!authStore.isAuthenticated) {
        console.warn('[PlatformRegistrationStore] fetchPlatformRegistrationById called while not authenticated.');
        this.currentPlatformRegistration = null;
        this.loading = false;
        return null;
      }

      this.loading = true;
      this.error = null;
      this.currentPlatformRegistration = null;
      try {
        const data = await platformRegistrationAPI.getById(id);
// Ensure data includes phone_number if returned by API
        this.currentPlatformRegistration = data;
        return data;
      } catch (err) {
        this.error = err.message || '获取平台注册详情失败';
        ElMessage.error(this.error);
        return null;
      } finally {
        this.loading = false;
      }
    },
    async createPlatformRegistration(data) {
      const authStore = useAuthStore();
      if (!authStore.isAuthenticated) {
        console.warn('[PlatformRegistrationStore] createPlatformRegistration called while not authenticated.');
        ElMessage.error('请先登录再创建平台注册信息');
        this.loading = false;
        return null;
      }

      this.loading = true;
      this.error = null;
      try {
// data should include phone_number from the form
        const createdData = await platformRegistrationAPI.create(data);
        ElMessage.success('平台注册信息创建成功');
        // Decide on re-fetch strategy, e.g., based on current view or always re-fetch first page
        await this.fetchPlatformRegistrations(1, this.pagination.pageSize, this.sort, this.filters); // Pass current filters
        return createdData;
      } catch (err) {
        if (err.response && err.response.status === 409) {
          this.error = err.response.data.message || '该邮箱账户已在此平台注册。';
          ElMessage.error(this.error);
        } else {
          this.error = err.message || '创建平台注册信息失败';
          ElMessage.error(this.error);
        }
        return null;
      } finally {
        this.loading = false;
      }
},
    async createPlatformRegistrationByName(data) {
      const authStore = useAuthStore();
      if (!authStore.isAuthenticated) {
        console.warn('[PlatformRegistrationStore] createPlatformRegistrationByName called while not authenticated.');
        ElMessage.error('请先登录再创建平台注册信息');
        this.loading = false;
        return null;
      }

      this.loading = true;
      this.error = null;
      try {
        // 假设 platformRegistrationAPI.createByName 将被添加到 api.js
// data should include phone_number from the form
        const createdData = await platformRegistrationAPI.createByName(data);
        ElMessage.success('平台注册信息创建成功 (by name)');
        await this.fetchPlatformRegistrations(1, this.pagination.pageSize, this.sort, this.filters); // Pass current filters
        return createdData;
      } catch (err) {
        if (err.response && err.response.status === 409) {
          this.error = err.response.data.message || '该邮箱账户已在此平台注册。';
          ElMessage.error(this.error);
        } else {
          this.error = err.message || '创建平台注册信息 (by name) 失败';
          ElMessage.error(this.error);
        }
        return null;
      } finally {
        this.loading = false;
      }
    },
    async updatePlatformRegistration(id, data) {
      const authStore = useAuthStore();
      if (!authStore.isAuthenticated) {
        console.warn('[PlatformRegistrationStore] updatePlatformRegistration called while not authenticated.');
        ElMessage.error('请先登录再更新平台注册信息');
        this.loading = false;
        return null;
      }

      this.loading = true;
      this.error = null;
// data should include phone_number from the form
      try {
        const updatedData = await platformRegistrationAPI.update(id, data);
        ElMessage.success('平台注册信息更新成功');
        await this.fetchPlatformRegistrations(this.pagination.currentPage, this.pagination.pageSize, this.sort, this.filters); // Pass current filters
// Ensure updatedData includes phone_number if returned by API
        if (this.currentPlatformRegistration && this.currentPlatformRegistration.id === id) {
          this.currentPlatformRegistration = updatedData;
        }
        return updatedData;
      } catch (err) {
        this.error = err.message || '更新平台注册信息失败';
        ElMessage.error(this.error);
        return null;
      } finally {
        this.loading = false;
      }
    },
    async deletePlatformRegistration(id) {
      const authStore = useAuthStore();
      if (!authStore.isAuthenticated) {
        console.warn('[PlatformRegistrationStore] deletePlatformRegistration called while not authenticated.');
        ElMessage.error('请先登录再删除平台注册信息');
        this.loading = false;
        return false;
      }

      this.loading = true;
      this.error = null;
      try {
        await platformRegistrationAPI.delete(id);
        ElMessage.success('平台注册信息删除成功');
        // Re-fetch, considering current page might become empty
         const currentPage = (this.platformRegistrations.length === 1 && this.pagination.currentPage > 1) 
                            ? this.pagination.currentPage - 1
                            : this.pagination.currentPage;
       await this.fetchPlatformRegistrations(currentPage, this.pagination.pageSize, this.sort, this.filters); // Pass current filters
       return true;
     } catch (err) {
        this.error = err.message || '删除平台注册信息失败';
        ElMessage.error(this.error);
        return false;
      } finally {
        this.loading = false;
      }
    },
    setCurrentPlatformRegistration(registration) {
        this.currentPlatformRegistration = registration;
    },
    async fetchAssociatedServiceSubscriptions(registrationId, params = { page: 1, pageSize: 5 }) {
      const authStore = useAuthStore();
      if (!authStore.isAuthenticated) {
        console.warn('[PlatformRegistrationStore] fetchAssociatedServiceSubscriptions called while not authenticated.');
        this.loading = false;
        // Return empty structure consistent with success case on error
        return { data: [], meta: { total_records: 0, current_page: 1, page_size: params.pageSize } };
      }

      // This action will call the new backend API to get service subscriptions for a specific platform registration.
      // It's similar to fetchAssociatedPlatformRegistrations in emailAccount.js
      this.loading = true;
      this.error = null;
      try {
        // Ensure platformRegistrationAPI has a method for this:
        // getAssociatedServiceSubscriptions: (registrationId, params) => api.get(`/platform-registrations/${registrationId}/service-subscriptions`, { params }),
        const result = await platformRegistrationAPI.getAssociatedServiceSubscriptions(registrationId, params);
        return result || { data: [], meta: { total_records: 0, current_page: 1, page_size: params.pageSize } };
      } catch (err) {
        this.error = err.message || '获取关联服务订阅失败';
        ElMessage.error(this.error);
        return { data: [], meta: { total_records: 0, current_page: 1, page_size: params.pageSize } };
      } finally {
        this.loading = false;
      }
    }
  },
});