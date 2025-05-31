<template>
  <el-card class="reminder-card" shadow="hover" @click="goToSubscriptionDetail">
    <template #header>
      <div class="card-header">
        <span>{{ reminder.service_name }}</span>
        <el-tag :type="tagType" size="small">{{ reminder.days_until_expiry }} 天后到期</el-tag>
      </div>
    </template>
    <div class="card-content">
      <p><strong>平台:</strong> {{ reminder.platform_name }}</p>
      <p><strong>到期日期:</strong> {{ formatDate(reminder.expiry_date) }}</p>
    </div>
  </el-card>
</template>

<script setup>
import { computed, defineProps } from 'vue';
import { useRouter } from 'vue-router';
import { ElCard, ElTag } from 'element-plus';

const props = defineProps({
  reminder: {
    type: Object,
    required: true,
    default: () => ({
      id: null, // Reminder ID
      service_subscription_id: null, // Subscription ID to navigate to
      service_name: 'N/A',
      platform_name: 'N/A',
      expiry_date: new Date().toISOString(),
      days_until_expiry: 0,
    })
  }
});

const router = useRouter();

const formatDate = (dateString) => {
  if (!dateString) return 'N/A';
  const date = new Date(dateString);
  return date.toLocaleDateString('zh-CN', { year: 'numeric', month: 'long', day: 'numeric' });
};

const tagType = computed(() => {
  if (props.reminder.days_until_expiry <= 7) {
    return 'danger';
  } else if (props.reminder.days_until_expiry <= 30) {
    return 'warning';
  }
  return 'info';
});

const goToSubscriptionDetail = () => {
  if (props.reminder.service_subscription_id) {
    router.push(`/service-subscriptions/${props.reminder.service_subscription_id}/edit`);
  } else {
    console.warn('Reminder card clicked, but no service_subscription_id found in reminder data:', props.reminder);
    // Optionally, show a message to the user or handle differently
  }
};
</script>

<style scoped>
.reminder-card {
  cursor: pointer;
  margin-bottom: 16px;
  transition: transform 0.2s ease-in-out, box-shadow 0.2s ease-in-out;
}

.reminder-card:hover {
  transform: translateY(-3px);
  box-shadow: var(--el-box-shadow-light); /* Using Element Plus variable for consistency */
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-weight: bold;
}

.card-content p {
  margin: 8px 0;
  font-size: 14px;
  color: #606266;
}

.card-content p strong {
  color: #303133;
}
</style>