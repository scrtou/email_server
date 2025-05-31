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
      pageSize: 10,
      totalItems: 0,
    },
    sort: { // New state for sorting
      orderBy: 'created_at',
      sortDirection: 'desc',
    },
  }),
  actions: {
    async fetchEmailAccounts(page = 1, pageSize = 10, sortOptions = {}) {
      this.loading = true;
      this.error = null;

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
          sortDirection: sortDirection
        };
        // api.js interceptor returns { data: [...], meta: {...} } for paginated responses
        const result = await emailAccountAPI.getAll(params);
        if (result && result.data) {
          this.emailAccounts = result.data;
          if (result.meta) {
            this.pagination.currentPage = result.meta.current_page;
            this.pagination.pageSize = result.meta.page_size;
            this.pagination.totalItems = result.meta.total_records;
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
        return result || { data: [], meta: { total_records: 0, current_page: 1, page_size: params.pageSize } };
      } catch (err) {
        this.error = err.message || '获取关联平台注册信息失败';
        ElMessage.error(this.error);
        return { data: [], meta: { total_records: 0, current_page: 1, page_size: params.pageSize } }; // Return empty structure on error
      } finally {
        this.loading = false;
      }
    }
  },
});