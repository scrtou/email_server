<template>
    <div class="notification-list-container">
      <div v-if="notificationStore.hasUnread" class="unread-count-header">
        <span>未读通知: {{ notificationStore.unreadCount }}</span>
        <button @click="handleMarkAllAsRead" class="mark-all-read-button">全部标记为已读</button>
      </div>
      <!-- 修改：使用 reminders 而不是 notifications -->
      <ul v-if="reminders && reminders.length > 0" class="notification-list">
        <li
          v-for="reminder in reminders"
          :key="reminder.id"
          :class="['notification-item', { unread: !reminder.read }]"
          @click="handleReminderClick(reminder)"
        >
          <span v-if="!reminder.read" class="unread-dot"></span>
          <span class="notification-message">
            {{ formatReminderMessage(reminder) }}
          </span>
        </li>
      </ul>
      <p v-else class="no-notifications">暂无通知</p>
    </div>
  </template>
  
  <script setup>
  import { computed, onMounted } from 'vue';
  import { useNotificationStore } from '@/stores/notification';
  
  const notificationStore = useNotificationStore();
  
  // 修改：访问store中的reminders数组
  const reminders = computed(() => {
    // 只显示未读的通知
    return notificationStore.reminders ? notificationStore.reminders.filter(r => !r.read) : [];
  });
  
  // 格式化提醒消息，使其更易读
  const formatReminderMessage = (reminder) => {
    if (reminder.days_until_expiry <= 0) {
      return `${reminder.service_name} 已到期`;
    } else if (reminder.days_until_expiry <= 7) {
      return `${reminder.service_name} 将在 ${reminder.days_until_expiry} 天后到期`;
    } else {
      return `${reminder.service_name} 将在 ${reminder.days_until_expiry} 天后到期`;
    }
  };
  
  const handleReminderClick = (reminder) => {
    if (!reminder.read) {
      notificationStore.markAsRead(reminder.id);
    }
  };
  
  const handleMarkAllAsRead = () => {
    notificationStore.markAllAsRead();
  };
  
  // 组件挂载时获取提醒数据
  onMounted(() => {
    if (!notificationStore.remindersLoaded) {
      notificationStore.fetchReminders();
    }
  });
  </script>

<style scoped>
.notification-list-container {
  position: relative;
  min-width: 300px;
  max-width: 450px;
  background-color: #fff;
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  padding: 10px;
  font-family: 'Arial', sans-serif;
}

.unread-count-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  border-bottom: 1px solid #f0f0f0;
  margin-bottom: 8px;
  position: sticky;
  top: 0;
  background-color: #fff; /* 确保背景不透明 */
  z-index: 1; /* 确保在其他内容之上 */
}

.unread-count-header span {
  font-size: 0.9em;
  color: #333;
  font-weight: bold;
}

.mark-all-read-button {
  background-color: #007bff;
  color: white;
  border: none;
  padding: 6px 12px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.8em;
  transition: background-color 0.2s;
}

.mark-all-read-button:hover {
  background-color: #0056b3;
}

.notification-list {
  list-style-type: none;
  padding: 0;
  margin: 0;
  max-height: 400px;
  overflow-y: auto;
}

.notification-item {
  padding: 10px 12px;
  border-bottom: 1px solid #f5f5f5;
  cursor: pointer;
  display: flex;
  align-items: center;
  transition: background-color 0.2s;
}

.notification-item:last-child {
  border-bottom: none;
}

.notification-item:hover {
  background-color: #f9f9f9;
}

.notification-item.unread {
  background-color: #e6f7ff; /* 淡蓝色背景表示未读 */
  font-weight: bold;
  color: #1890ff;
}

.unread-dot {
  width: 8px;
  height: 8px;
  background-color: #1890ff; /* 蓝色小圆点 */
  border-radius: 50%;
  margin-right: 10px;
  flex-shrink: 0;
}

.notification-message {
  font-size: 0.95em;
  color: #555;
  line-height: 1.4;
  word-break: break-word; /* 防止长文本溢出导致横向滚动 */
}

.notification-item.unread .notification-message {
  color: #0056b3; /* 未读消息的文字颜色加深 */
}

.no-notifications {
  text-align: center;
  padding: 20px;
  color: #888;
  font-size: 0.9em;
}
</style>