import { defineStore } from 'pinia';
import { dashboardAPI } from '@/utils/api';
import { ElMessage } from 'element-plus';
import { useAuthStore } from './auth'; // 导入 Auth Store

export const useDashboardStore = defineStore('dashboard', {
  state: () => ({
    summaryData: null,
    loading: false,
    error: null
  }),
  actions: {
    async fetchDashboardSummary() {
      const authStore = useAuthStore();
      if (!authStore.isAuthenticated) {
        console.warn('[DashboardStore] fetchDashboardSummary called while not authenticated.');
        // Decide what to do: return, clear data, or show message?
        // For now, just return to prevent API call.
        this.summaryData = null; // Clear data if not authenticated
        this.loading = false;
        return;
      }

      this.loading = true;
      this.error = null;
      try {
        const response = await dashboardAPI.getSummary();
        // The response interceptor in api.js already unwraps response.data.data
        // So, 'response' here should be the actual summary data object
        this.summaryData = response
        console.log('Dashboard summary data fetched:', this.summaryData)
      } catch (error) {
        this.error = error.message || '获取仪表盘数据失败'
        ElMessage.error(this.error)
        console.error('Error fetching dashboard summary:', error)
      } finally {
        this.loading = false
      }
    }
  },
  getters: {
    isLoading: (state) => state.loading,
    hasError: (state) => state.error !== null,
    getSummaryData: (state) => state.summaryData,
    // Example getters for specific pieces of data, can be expanded
    activeSubscriptionsCount: (state) => state.summaryData?.active_subscriptions_count || 0,
    estimatedMonthlySpending: (state) => state.summaryData?.estimated_monthly_spending || 0,
    estimatedYearlySpending: (state) => state.summaryData?.estimated_yearly_spending || 0,
    upcomingRenewals: (state) => state.summaryData?.upcoming_renewals || [],
    subscriptionsByPlatform: (state) => state.summaryData?.subscriptions_by_platform || [],
    totalEmailAccounts: (state) => state.summaryData?.total_email_accounts || 0,
    totalPlatforms: (state) => state.summaryData?.total_platforms || 0,
    totalPlatformRegistrations: (state) => state.summaryData?.total_platform_registrations || 0,
    totalServiceSubscriptions: (state) => state.summaryData?.total_service_subscriptions || 0,
  }
})