import { defineStore } from 'pinia';
import { serviceSubscriptionAPI } from '@/utils/api';
import { ElMessage } from 'element-plus';

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
  state: () => ({
    serviceSubscriptions: [],
    currentServiceSubscription: null,
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
      status: '',
      billing_cycle: '',
      renewal_date_start: '',
      renewal_date_end: '',
      // platform_registration_id: null, // Removed platform_registration_id
    }
  }),
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
      // this.filters.platform_registration_id = null; // Removed platform_registration_id
      this.pagination.currentPage = 1;
      this.fetchServiceSubscriptions(this.pagination.currentPage, this.pagination.pageSize, this.sort, this.filters);
    },
    async fetchServiceSubscriptions(
        page = this.pagination.currentPage,
        pageSize = this.pagination.pageSize,
        sortOptions = {},
        filterOptions = {}
    ) {
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
      // if (filterOptions.platform_registration_id !== undefined) this.filters.platform_registration_id = filterOptions.platform_registration_id; // Removed
      
      const apiParams = {
        page,
        pageSize,
        orderBy,
        sortDirection,
        status: this.filters.status || undefined,
        billing_cycle: this.filters.billing_cycle || undefined,
        renewal_date_start: this.filters.renewal_date_start || undefined,
        renewal_date_end: this.filters.renewal_date_end || undefined,
        // platform_registration_id: this.filters.platform_registration_id || undefined, // Removed
      };
      
      try {
        const result = await serviceSubscriptionAPI.getAll(apiParams);
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
        this.error = err.message || '获取服务订阅列表失败';
        ElMessage.error(this.error);
        this.serviceSubscriptions = [];
        this.pagination = { currentPage: 1, pageSize: 10, totalItems: 0 };
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
        const createdData = await serviceSubscriptionAPI.create(data);
        ElMessage.success('服务订阅创建成功');
        await this.fetchServiceSubscriptions(1, this.pagination.pageSize, this.sort, this.filters);
        return createdData;
      } catch (err) {
        this.error = err.message || '创建服务订阅失败';
        ElMessage.error(this.error);
        return null;
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
        return updatedData;
      } catch (err) {
        this.error = err.message || '更新服务订阅失败';
        ElMessage.error(this.error);
        return null;
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