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
      pageSize: 10,
      totalItems: 0,
    },
  }),
  actions: {
    async fetchServiceSubscriptions(params = { page: 1, pageSize: 10 }) {
      this.loading = true;
      this.error = null;
      const page = params.page || this.pagination.currentPage;
      const pageSize = params.pageSize || this.pagination.pageSize;
      
      try {
        const result = await serviceSubscriptionAPI.getAll({ params });
        if (result && result.data) {
          this.serviceSubscriptions = result.data;
          if (result.meta) {
            this.pagination.currentPage = result.meta.current_page;
            this.pagination.pageSize = result.meta.page_size;
            this.pagination.totalItems = result.meta.total_records;
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
        await this.fetchServiceSubscriptions({ page: 1, pageSize: this.pagination.pageSize });
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
        await this.fetchServiceSubscriptions({
            page: this.pagination.currentPage, 
            pageSize: this.pagination.pageSize 
        });
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
        await this.fetchServiceSubscriptions({ page: currentPage, pageSize: this.pagination.pageSize });
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