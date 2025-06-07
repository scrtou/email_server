import { defineStore } from 'pinia';
import { serviceSubscriptionAPI } from '@/utils/api';
import { ElMessage } from 'element-plus';
import { useAuthStore } from './auth'; // Import auth store
import { useSettingsStore } from './settings'; // 导入 Settings Store

// Assuming serviceSubscriptionAPI will be added to api.js:
/*
// In api.js:
export const serviceSubscriptionAPI = {
  getAll: (params = {}) => api.get('/service-subscriptions', { params }),
  getById: (id) => api.get(`/service-subscriptions/${id}`),
  create: (data) => api.post('/service-subscriptions', data),
  update: (id, data) => api.put(`/service-subscriptions/${id}`, data),
  delete: (id) => api.delete(`/service-subscriptions/${id}`),
};
*/

export const useServiceSubscriptionStore = defineStore('serviceSubscription', {
  state: () => {
    const settingsStore = useSettingsStore();
    return {
      serviceSubscriptions: [],
      currentServiceSubscription: null,
      loading: false,
      error: null,
      pagination: {
        currentPage: 1,
        pageSize: settingsStore.getDefaultPageSize,
        totalItems: 0,
      },
      sort: { // New state for sorting
        orderBy: 'created_at', // Default sort
        sortDirection: 'desc',
      },
      filters: { // New state for filters
        status: '',
        billing_cycle: '',
        renewal_date_start: '',
        renewal_date_end: '',
        filterPlatformName: '', // New filter
        filterEmail: '', // New filter
        filterUsername: '', // New filter
      },
    };
  },
  actions: {
    // Action to update filter values
    setFilter(filterName, value) {
      if (Object.prototype.hasOwnProperty.call(this.filters, filterName)) {
        this.filters[filterName] = value;
        this.pagination.currentPage = 1; // Reset to first page
        // Re-fetch data with all current filters, sort, and pagination
        this.fetchServiceSubscriptions(this.pagination.currentPage, this.pagination.pageSize, this.sort, this.filters);
      }
    },
    // Action to clear all filters
    clearFilters() {
      this.filters.status = '';
      this.filters.billing_cycle = '';
      this.filters.renewal_date_start = '';
      this.filters.renewal_date_end = '';
      this.filters.filterPlatformName = ''; // Clear new filter
      this.filters.filterEmail = ''; // Clear new filter
      this.filters.filterUsername = ''; // Clear new filter
      this.pagination.currentPage = 1;
      this.fetchServiceSubscriptions(this.pagination.currentPage, this.pagination.pageSize, this.sort, this.filters);
    },
    async fetchServiceSubscriptions(
        page = this.pagination.currentPage,
        pageSize = this.pagination.pageSize,
        sortOptions = {},
        filterOptions = {},
        signal // Add signal parameter
    ) {
      // --- BEGIN AUTH CHECK ---
      const authStore = useAuthStore();
      if (!authStore.isAuthenticated) {
        console.log('[ServiceSubscriptionStore] User not authenticated. Skipping fetchServiceSubscriptions.');
        // We simply return early. The component's loading state might still be true
        // if it was set before calling this action, but no API call will be made,
        // and no error will be shown for this specific case.
        // If necessary, we could set this.loading = false here, but it might interfere
        // with component logic expecting loading state during transitions.
        return; // Stop execution if not authenticated
      }
      // --- END AUTH CHECK ---

      this.loading = true;
      this.error = null;

      const orderBy = sortOptions.orderBy || this.sort.orderBy;
      const sortDirection = sortOptions.sortDirection || this.sort.sortDirection;

      // Update sort state
      if (sortOptions.orderBy) this.sort.orderBy = sortOptions.orderBy;
      if (sortOptions.sortDirection) this.sort.sortDirection = sortOptions.sortDirection;
      
      // Update filter state if new options are provided
      if (filterOptions.status !== undefined) this.filters.status = filterOptions.status;
      if (filterOptions.billing_cycle !== undefined) this.filters.billing_cycle = filterOptions.billing_cycle;
      if (filterOptions.renewal_date_start !== undefined) this.filters.renewal_date_start = filterOptions.renewal_date_start;
      if (filterOptions.renewal_date_end !== undefined) this.filters.renewal_date_end = filterOptions.renewal_date_end;
      if (filterOptions.filterPlatformName !== undefined) this.filters.filterPlatformName = filterOptions.filterPlatformName; // New
      if (filterOptions.filterEmail !== undefined) this.filters.filterEmail = filterOptions.filterEmail; // New
      if (filterOptions.filterUsername !== undefined) this.filters.filterUsername = filterOptions.filterUsername; // New
      
      const apiParams = {
        page,
        pageSize,
        orderBy,
        sortDirection,
        status: this.filters.status || undefined,
        billing_cycle: this.filters.billing_cycle || undefined,
        renewal_date_start: this.filters.renewal_date_start || undefined,
        renewal_date_end: this.filters.renewal_date_end || undefined,
        platform_name: this.filters.filterPlatformName || undefined, // New
        email: this.filters.filterEmail || undefined, // New
        username: this.filters.filterUsername || undefined, // New
      };
      
      try {
        // Pass signal to the API call
        const result = await serviceSubscriptionAPI.getAll(apiParams, signal);
        if (result && result.data) {
          this.serviceSubscriptions = result.data;
          if (result.meta) {
            this.pagination.currentPage = result.meta.current_page;
            this.pagination.pageSize = result.meta.page_size;
            this.pagination.totalItems = result.meta.total_items;
          } else {
            this.pagination = { currentPage: page, pageSize: pageSize, totalItems: result.data.length };
          }
        } else {
          this.serviceSubscriptions = [];
          this.pagination = { currentPage: 1, pageSize: 10, totalItems: 0 };
        }
      } catch (err) {
        // Check if the error is an AbortError
        if (err.name === 'AbortError') {
          console.log('Fetch aborted by user action.'); // Optional: log cancellation
          // Do not set error state or show message for aborted requests
        } else {
          this.error = err.message || '获取服务订阅列表失败';
          ElMessage.error(this.error);
          this.serviceSubscriptions = [];
          this.pagination = { currentPage: 1, pageSize: 10, totalItems: 0 };
        }
      } finally {
        this.loading = false;
      }
    },
    async fetchServiceSubscriptionById(id) {
      this.loading = true;
      this.error = null;
      this.currentServiceSubscription = null;
      try {
        const data = await serviceSubscriptionAPI.getById(id);
        this.currentServiceSubscription = data;
        return data;
      } catch (err) {
        this.error = err.message || '获取服务订阅详情失败';
        ElMessage.error(this.error);
        return null;
      } finally {
        this.loading = false;
      }
    },
    async createServiceSubscription(data) {
      this.loading = true;
      this.error = null;
      try {
        await serviceSubscriptionAPI.create(data);
        ElMessage.success('服务订阅创建成功');
        await this.fetchServiceSubscriptions(1, this.pagination.pageSize, this.sort, this.filters);
        return true; // 返回 true 表示成功
      } catch (err) {
        console.error('[ServiceSubscriptionStore] createServiceSubscription error:', err);
        this.error = err.message || '创建服务订阅失败';
        ElMessage.error(this.error);
        return false; // 返回 false 表示失败
      } finally {
        this.loading = false;
      }
    },
    async updateServiceSubscription(id, data) {
      this.loading = true;
      this.error = null;
      try {
        const updatedData = await serviceSubscriptionAPI.update(id, data);
        ElMessage.success('服务订阅更新成功');
        await this.fetchServiceSubscriptions(this.pagination.currentPage, this.pagination.pageSize, this.sort, this.filters);
        if (this.currentServiceSubscription && this.currentServiceSubscription.id === id) {
          this.currentServiceSubscription = updatedData;
        }
        return true; // 返回 true 表示成功
      } catch (err) {
        console.error('[ServiceSubscriptionStore] updateServiceSubscription error:', err);
        this.error = err.message || '更新服务订阅失败';
        ElMessage.error(this.error);
        return false; // 返回 false 表示失败
      } finally {
        this.loading = false;
      }
    },
    async deleteServiceSubscription(id) {
      this.loading = true;
      this.error = null;
      try {
        await serviceSubscriptionAPI.delete(id);
        ElMessage.success('服务订阅删除成功');
        const currentPage = (this.serviceSubscriptions.length === 1 && this.pagination.currentPage > 1)
                            ? this.pagination.currentPage - 1
                            : this.pagination.currentPage;
       await this.fetchServiceSubscriptions(currentPage, this.pagination.pageSize, this.sort, this.filters);
       return true;
     } catch (err) {
        this.error = err.message || '删除服务订阅失败';
        ElMessage.error(this.error);
        return false;
      } finally {
        this.loading = false;
      }
    },
    setCurrentServiceSubscription(subscription) {
        this.currentServiceSubscription = subscription;
    }
  },
});