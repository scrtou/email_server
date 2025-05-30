import { defineStore } from 'pinia';
import { emailAccountAPI } from '@/utils/api';
import { ElMessage } from 'element-plus';

export const useEmailAccountStore = defineStore('emailAccount', {
  state: () => ({
    emailAccounts: [],
    currentEmailAccount: null,
    loading: false,
    error: null,
    pagination: {
      currentPage: 1,
      pageSize: 7,
      totalItems: 0,
    },
    sort: { // New state for sorting
      orderBy: 'created_at',
      sortDirection: 'desc',
    },
    filters: { // New state for filters
      provider: '',
      emailAddressSearch: '', // For email address fuzzy search
    },
    uniqueProviders: [], // For provider filter dropdown
  }),
  actions: {
    // Action to update filter values
    setFilter(filterName, value) {
      if (Object.prototype.hasOwnProperty.call(this.filters, filterName)) {
        this.filters[filterName] = value;
        // Reset to first page when filters change
        this.pagination.currentPage = 1;
        // fetchEmailAccounts will be called by the component after filter change
        // or if we want the store to be self-contained for this, then call:
        // this.fetchEmailAccounts(1, this.pagination.pageSize, this.sort, this.filters);
      }
    },
    // Action to clear all filters
    clearFilters() {
      this.filters.provider = '';
      this.filters.emailAddressSearch = '';
      this.pagination.currentPage = 1;
      // this.fetchEmailAccounts(1, this.pagination.pageSize, this.sort, this.filters);
    },
    async fetchEmailAccounts(page = this.pagination.currentPage, pageSize = this.pagination.pageSize, sortOptions = {}, filters = {}) {
      this.loading = true;
      this.error = null;

      // Update store's filter state if new filters are passed and different
      if (filters.provider !== undefined && this.filters.provider !== filters.provider) {
        this.filters.provider = filters.provider;
        // If filters are externally updated, reset to page 1
        if (page !== 1) page = 1;
        this.pagination.currentPage = page;
      }
      if (filters.emailAddressSearch !== undefined && this.filters.emailAddressSearch !== filters.emailAddressSearch) {
        this.filters.emailAddressSearch = filters.emailAddressSearch;
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
          provider: this.filters.provider || undefined, // Use the potentially updated store filter
          email_address: this.filters.emailAddressSearch || undefined, // Add email address search
        };
        // api.js interceptor returns { data: [...], meta: {...} } for paginated responses
        const result = await emailAccountAPI.getAll(params);
        if (result && result.data) {
          this.emailAccounts = result.data;
          if (result.meta) {
            this.pagination.currentPage = result.meta.current_page;
            this.pagination.pageSize = result.meta.page_size;
            this.pagination.totalItems = result.meta.total_items;
          } else {
            // Fallback if meta is somehow not present, though API should provide it
            this.pagination = { currentPage: page, pageSize: pageSize, totalItems: result.data.length };
          }
        } else {
          this.emailAccounts = [];
          this.pagination = { currentPage: 1, pageSize: 10, totalItems: 0 };
        }
      } catch (err) {
        this.error = err.message || '获取邮箱账户列表失败';
        ElMessage.error(this.error);
        this.emailAccounts = [];
        this.pagination = { currentPage: 1, pageSize: 10, totalItems: 0 };
      } finally {
        this.loading = false;
      }
    },
    async fetchEmailAccountById(id) {
      this.loading = true;
      this.error = null;
      this.currentEmailAccount = null;
      try {
        // api.js interceptor returns the 'data' part of the backend response
        const data = await emailAccountAPI.getById(id);
        this.currentEmailAccount = data;
        return data;
      } catch (err) {
        this.error = err.message || '获取邮箱账户详情失败';
        ElMessage.error(this.error);
        return null;
      } finally {
        this.loading = false;
      }
    },
    async createEmailAccount(data) {
      this.loading = true;
      this.error = null;
      try {
        const createdData = await emailAccountAPI.create(data);
        ElMessage.success('邮箱账户创建成功');
        await this.fetchEmailAccounts(this.pagination.currentPage, this.pagination.pageSize);
        return createdData;
      } catch (err) {
        const errorMessage = err.message || '创建邮箱账户失败';
        if (errorMessage.includes('UNIQUE constraint failed') && errorMessage.includes('email_accounts.email_address')) {
          this.error = '此邮箱地址已被注册。请尝试使用其他邮箱地址，或使用此邮箱地址登录。';
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
    async updateEmailAccount(id, data) {
      this.loading = true;
      this.error = null;
      try {
        const updatedData = await emailAccountAPI.update(id, data);
        ElMessage.success('邮箱账户更新成功');
        await this.fetchEmailAccounts(this.pagination.currentPage, this.pagination.pageSize);
        if (this.currentEmailAccount && this.currentEmailAccount.id === id) {
            this.currentEmailAccount = updatedData;
        }
        return updatedData;
      } catch (err) {
        this.error = err.message || '更新邮箱账户失败';
        ElMessage.error(this.error);
        return null;
      } finally {
        this.loading = false;
      }
    },
    async deleteEmailAccount(id) {
      this.loading = true;
      this.error = null;
      try {
        await emailAccountAPI.delete(id);
        ElMessage.success('邮箱账户删除成功');
        // Re-fetch list
        await this.fetchEmailAccounts(this.pagination.currentPage, this.pagination.pageSize);
        return true;
      } catch (err) {
        this.error = err.message || '删除邮箱账户失败';
        ElMessage.error(this.error);
        return false;
      } finally {
        this.loading = false;
      }
    },
    setCurrentEmailAccount(account) {
        this.currentEmailAccount = account;
    },
    async fetchAssociatedPlatformRegistrations(emailAccountId, params = { page: 1, pageSize: 5 }) {
      this.loading = true;
      this.error = null;
      try {
        // api.js interceptor returns { data: [...], meta: {...} } for paginated responses
        const result = await emailAccountAPI.getAssociatedPlatformRegistrations(emailAccountId, params);
        return result || { data: [], meta: { total_items: 0, current_page: 1, page_size: params.pageSize } };
      } catch (err) {
        this.error = err.message || '获取关联平台注册信息失败';
        ElMessage.error(this.error);
        return { data: [], meta: { total_items: 0, current_page: 1, page_size: params.pageSize } }; // Return empty structure on error
      } finally {
        this.loading = false;
      }
    },
    async fetchUniqueProviders() {
      // this.loading = true; // Optional: manage loading state for this specific fetch
      try {
        const providers = await emailAccountAPI.getUniqueProviders(); // Assumes this method will be added to api.js
        this.uniqueProviders = providers || [];
      } catch (err) {
        // ElMessage.error(err.message || '获取服务商列表失败'); // Optional: show error
        this.uniqueProviders = []; // Ensure it's an array on error
      } finally {
        // this.loading = false; // Optional: manage loading state
      }
    }
  },
});