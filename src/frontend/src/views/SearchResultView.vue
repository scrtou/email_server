<template>
  <div class="search-results-view">
    <h1>搜索结果</h1>
    <div v-if="searchStore.isLoading" class="loading-indicator">
      <el-skeleton :rows="5" animated />
    </div>
    <div v-else-if="searchStore.error" class="error-message">
      <el-alert type="error" :title="searchStore.error" show-icon :closable="false" />
    </div>
    <div v-else-if="!searchStore.results || searchStore.results.length === 0 && searchStore.searchTerm" class="no-results">
      <el-empty :description="`未能找到与 '${searchStore.searchTerm}' 相关的内容`" />
    </div>
    <div v-else-if="!searchStore.searchTerm" class="no-results">
      <el-empty description="请输入关键词进行搜索" />
    </div>
    <div v-else class="results-list">
      <el-card v-for="item in searchStore.results" :key="item.id + '-' + item.type" class="result-item">
        <template #header>
          <div class="result-item-header">
            <span>{{ getItemTypeDisplayName(item.type) }}: {{ getItemDisplayName(item) }}</span>
            <el-button type="primary" link @click="navigateToItem(item)">查看详情</el-button>
          </div>
        </template>
        <div class="result-item-details">
          <p v-if="item.email_address"><strong>邮箱地址:</strong> {{ item.email_address }}</p>
          <p v-if="item.provider"><strong>服务商:</strong> {{ item.provider }}</p>
          <p v-if="item.platform_name"><strong>平台名称:</strong> {{ item.platform_name }}</p>
          <p v-if="item.url"><strong>网址:</strong> <a :href="item.url" target="_blank">{{ item.url }}</a></p>
          <p v-if="item.username"><strong>用户名:</strong> {{ item.username }}</p>
          <p v-if="item.registration_date"><strong>注册日期:</strong> {{ formatDate(item.registration_date) }}</p>
          <p v-if="item.service_name"><strong>服务名称:</strong> {{ item.service_name }}</p>
          <p v-if="item.status"><strong>状态:</strong> {{ item.status }}</p>
          <p v-if="item.renewal_date"><strong>续费日期:</strong> {{ formatDate(item.renewal_date) }}</p>
          <p v-if="item.notes"><strong>备注:</strong> {{ item.notes }}</p>
        </div>
      </el-card>
    </div>
  </div>
</template>

<script setup>
import { onMounted, watch } from 'vue';
import { useSearchStore } from '@/stores/search';
import { useRouter, useRoute } from 'vue-router';
import { ElMessage } from 'element-plus';

const searchStore = useSearchStore();
const router = useRouter();
const route = useRoute();

const getItemTypeDisplayName = (type) => {
  const typeMap = {
    email_account: '邮箱账户',
    platform: '平台',
    platform_registration: '平台注册',
    service_subscription: '服务订阅',
  };
  return typeMap[type] || '未知类型';
};

const getItemDisplayName = (item) => {
  switch (item.type) {
    case 'email_account':
      return item.email_address || 'N/A';
    case 'platform':
      return item.name || 'N/A';
    case 'platform_registration':
      return `${item.platform_name} (${item.username || 'N/A'})`;
    case 'service_subscription':
      return item.service_name || 'N/A';
    default:
      return 'N/A';
  }
};

const formatDate = (dateString) => {
  if (!dateString) return 'N/A';
  const date = new Date(dateString);
  return date.toLocaleDateString();
};

const navigateToItem = (item) => {
  switch (item.type) {
    case 'email_account':
      ElMessage.info(`导航到邮箱账户详情: ${item.id} (功能待实现或调整)`);
      router.push('/email-accounts');
      break;
    case 'platform':
      ElMessage.info(`导航到平台详情: ${item.id} (功能待实现或调整)`);
      router.push('/platforms');
      break;
    case 'platform_registration':
      ElMessage.info(`导航到平台注册详情: ${item.id} (功能待实现或调整)`);
      router.push('/platform-registrations');
      break;
    case 'service_subscription':
      ElMessage.info(`导航到服务订阅详情: ${item.id} (功能待实现或调整)`);
      router.push('/service-subscriptions');
      break;
    default:
      ElMessage.warning('未知的结果类型，无法导航。');
  }
};

watch(
  () => route.query,
  (newQuery) => {
    if (route.name === 'search-results') {
        if (newQuery.term !== searchStore.searchTerm || newQuery.type !== searchStore.searchType) {
            searchStore.loadSearchTermFromQuery(newQuery);
        }
    }
  },
  { deep: true }
);

onMounted(() => {
  if (route.name === 'search-results' && (route.query.term || route.query.type)) {
    if (searchStore.searchTerm !== route.query.term || searchStore.searchType !== route.query.type) {
        searchStore.loadSearchTermFromQuery(route.query);
    } else if (!searchStore.results.length && searchStore.searchTerm) {
        searchStore.performSearch(searchStore.searchTerm, searchStore.searchType);
    }
  }
});
</script>

<style scoped>
.search-results-view {
  padding: 20px;
}
.loading-indicator, .error-message, .no-results {
  margin-top: 20px;
  text-align: center;
}
.results-list {
  margin-top: 20px;
}
.result-item {
  margin-bottom: 20px;
}
.result-item-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.result-item-details p {
  margin: 5px 0;
  font-size: 0.9em;
  color: #606266;
}
.result-item-details strong {
  color: #303133;
}
</style>