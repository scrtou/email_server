import { defineStore } from 'pinia';
import { platformRegistrationAPI } from '@/utils/api';
import { ElMessage } from 'element-plus';

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
    loading: false,
    error: null,
    pagination: {
      currentPage: 1,
      pageSize: 10,
      totalItems: 0,
    },
    sort: { // New state for sorting
      orderBy: 'created_at', // Default sort
      sortDirection: 'desc',
    },
  }),
  actions: {
    async fetchPlatformRegistrations(params = { page: 1, pageSize: 10, orderBy: 'created_at', sortDirection: 'desc' }) {
      this.loading = true;
      this.error = null;
      
      const page = params.page || this.pagination.currentPage;
      const pageSize = params.pageSize || this.pagination.pageSize;
      const orderBy = params.orderBy || this.sort.orderBy;
      const sortDirection = params.sortDirection || this.sort.sortDirection;

      // Update sort state if new options are provided directly in params
      if (params.orderBy) this.sort.orderBy = params.orderBy;
      if (params.sortDirection) this.sort.sortDirection = params.sortDirection;

      const apiParams = { ...params, page, pageSize, orderBy, sortDirection };

      try {
        const result = await platformRegistrationAPI.getAll(apiParams);
        if (result && result.data) {
          this.platformRegistrations = result.data;
          if (result.meta) {
            this.pagination.currentPage = result.meta.current_page;
            this.pagination.pageSize = result.meta.page_size;
            this.pagination.totalItems = result.meta.total_records;
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
      this.loading = true;
      this.error = null;
      this.currentPlatformRegistration = null;
      try {
        const data = await platformRegistrationAPI.getById(id);
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
      this.loading = true;
      this.error = null;
      try {
        const createdData = await platformRegistrationAPI.create(data);
        ElMessage.success('平台注册信息创建成功');
        // Decide on re-fetch strategy, e.g., based on current view or always re-fetch first page
        await this.fetchPlatformRegistrations({
          page: 1,
          pageSize: this.pagination.pageSize,
          orderBy: this.sort.orderBy,
          sortDirection: this.sort.sortDirection,
        });
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
      this.loading = true;
      this.error = null;
      try {
        // 假设 platformRegistrationAPI.createByName 将被添加到 api.js
        const createdData = await platformRegistrationAPI.createByName(data);
        ElMessage.success('平台注册信息创建成功 (by name)');
        await this.fetchPlatformRegistrations({
          page: 1,
          pageSize: this.pagination.pageSize,
          orderBy: this.sort.orderBy,
          sortDirection: this.sort.sortDirection,
        });
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
      this.loading = true;
      this.error = null;
      try {
        const updatedData = await platformRegistrationAPI.update(id, data);
        ElMessage.success('平台注册信息更新成功');
        await this.fetchPlatformRegistrations({
            page: this.pagination.currentPage,
            pageSize: this.pagination.pageSize,
            orderBy: this.sort.orderBy,
            sortDirection: this.sort.sortDirection,
        });
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
      this.loading = true;
      this.error = null;
      try {
        await platformRegistrationAPI.delete(id);
        ElMessage.success('平台注册信息删除成功');
        // Re-fetch, considering current page might become empty
         const currentPage = (this.platformRegistrations.length === 1 && this.pagination.currentPage > 1) 
                            ? this.pagination.currentPage - 1
                            : this.pagination.currentPage;
       await this.fetchPlatformRegistrations({
           page: currentPage,
           pageSize: this.pagination.pageSize,
           orderBy: this.sort.orderBy,
           sortDirection: this.sort.sortDirection,
       });
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