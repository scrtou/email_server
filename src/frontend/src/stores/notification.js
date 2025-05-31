import { defineStore } from 'pinia';
import { authAPI } from '@/utils/api';

export const useNotificationStore = defineStore('notification', {
  state: () => ({
    reminders: [],
    isLoading: false,
    error: null,
  }),
  getters: {
    unreadCount: (state) => {
      return Array.isArray(state.reminders) ? state.reminders.filter(r => !r.read).length : 0;
    },
    hasUnread: (state) => {
      return Array.isArray(state.reminders) && state.reminders.some(r => !r.read);
    }
  },
  actions: {
    async fetchReminders() {
      console.log('Attempting to fetch reminders in notificationStore...');
      this.isLoading = true;
      this.error = null;
      try {
        const responseData = await authAPI.getReminders();
        console.log('responseData in notificationStore:', responseData);
        
        // 数据转换：将API返回的字段名转换为组件期望的格式
        const transformedReminders = responseData && Array.isArray(responseData.reminders)
          ? responseData.reminders.map(reminder => ({
              id: reminder.id,
              service_subscription_id: reminder.id, // 假设用reminder.id作为subscription_id，需要根据实际情况调整
              service_name: reminder.serviceName,
              platform_name: reminder.platformName,
              expiry_date: reminder.renewalDate,
              days_until_expiry: reminder.daysRemaining,
              status: reminder.status,
              read: reminder.is_read // 从后端获取 is_read 状态
            }))
          : [];
            
          console.log('Transformed reminders:', transformedReminders);
        this.reminders = transformedReminders;
      } catch (error) {
        console.error('Failed to fetch reminders:', error);
        this.error = error.response?.data?.message || '获取提醒失败';
        this.reminders = [];
      } finally {
        this.isLoading = false;
      }
    },
    clearReminders() {
      this.reminders = [];
      this.error = null;
    },
    async markAsRead(id) {
      try {
        await authAPI.markReminderAsRead(id);
        const reminder = this.reminders.find(r => r.id === id);
        if (reminder) {
          reminder.read = true;
        }
      } catch (error) {
        console.error('Failed to mark reminder as read:', error);
        // 可以选择性地向用户显示错误信息
        // this.error = error.response?.data?.message || '标记已读失败';
      }
    },
    async markAllAsRead() { // 虽然本次任务不要求，但保持异步风格一致性
      // 注意：后端可能没有批量标记已读的接口，这里仅修改前端状态
      // 如果需要后端支持，则需要额外实现批量标记接口并调用
      this.reminders.forEach(r => r.read = true);
      // 示例：如果后端支持批量标记，则类似这样调用
      // try {
      //   await authAPI.markAllRemindersAsRead(this.reminders.map(r => r.id));
      //   this.reminders.forEach(r => r.read = true);
      // } catch (error) {
      //   console.error('Failed to mark all reminders as read:', error);
      // }
    }
  },
});