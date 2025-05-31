<template>
  <div class="service-subscription-list-view">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>服务订阅管理</span>
  <el-button type="primary" @click="handleAdd">
            <el-icon><Plus /></el-icon> 添加订阅
          </el-button>
        </div>
      </template>

      <el-form :inline="true" class="filter-form">
        <el-form-item label="平台注册信息">
          <el-select
            v-model="serviceSubscriptionStore.filters.platform_registration_id"
            placeholder="选择平台注册信息"
            clearable
            filterable
            @change="handlePlatformRegistrationFilterChange"
            style="width: 220px;"
          >
            <el-option
              v-for="item in platformRegistrationStore.platformRegistrations"
              :key="item.id"
              :label="`${item.platform_name} - ${item.email_address} (${item.login_username || '无用户名'})`"
              :value="item.id"
            />
          </el-select>
        </el-form-item>
         <el-form-item label="订阅状态">
          <el-select
            v-model="serviceSubscriptionStore.filters.status"
            placeholder="选择状态"
            clearable
            filterable
            @change="handleStatusFilterChange"
            style="width: 180px;"
          >
            <el-option label="全部状态" value="" />
            <el-option label="活跃 (Active)" value="active" />
            <el-option label="不活跃 (Inactive)" value="inactive" />
            <el-option label="已过期 (Expired)" value="expired" />
            <el-option label="试用 (Trial)" value="free_trial" />
            <el-option label="已取消 (Cancelled)" value="cancelled" />
            <el-option label="待处理 (Pending)" value="pending" />
          </el-select>
        </el-form-item>
        <el-form-item label="计费周期">
          <el-select
            v-model="serviceSubscriptionStore.filters.billing_cycle"
            placeholder="选择计费周期"
            clearable
            filterable
            @change="handleBillingCycleFilterChange"
            style="width: 180px;"
          >
            <el-option label="全部周期" value="" />
            <el-option label="月付 (Monthly)" value="monthly" />
            <el-option label="年付 (Annually)" value="annually" />
            <el-option label="一次性 (One-time)" value="one_time" />
            <el-option label="其他 (Other)" value="other" />
          </el-select>
        </el-form-item>
        <el-form-item label="续费日期范围">
          <el-date-picker
            :model-value="[serviceSubscriptionStore.filters.renewal_date_start, serviceSubscriptionStore.filters.renewal_date_end]"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            clearable
            @change="handleRenewalDateChange"
            value-format="YYYY-MM-DD"
            style="width: 280px;"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="triggerFetchWithCurrentFilters">查询</el-button>
          <el-button @click="triggerClearFilters">重置</el-button>
        </el-form-item>
      </el-form>

      <el-table
        :data="serviceSubscriptionStore.serviceSubscriptions"
        v-loading="serviceSubscriptionStore.loading"
        style="width: 100%"
        @sort-change="handleSortChange"
        :default-sort="{ prop: serviceSubscriptionStore.sort.orderBy, order: serviceSubscriptionStore.sort.sortDirection === 'desc' ? 'descending' : 'ascending' }"
      >
        <el-table-column prop="service_name" label="服务名称" min-width="180" sortable="custom" />
        <!-- <el-table-column prop="id" label="ID" width="60" /> -->
        <el-table-column prop="platform_name" label="所属平台" min-width="150" sortable="custom" />
        <el-table-column label="平台邮箱" min-width="200" prop="email_address" sortable="custom">
          <template #default="scope">
            <span>{{ scope.row.email_address }}</span>
            <span v-if="scope.row.login_username" style="color: #909399; margin-left: 5px;">({{ scope.row.login_username }})</span>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100" sortable="custom" />
        <el-table-column prop="cost" label="费用" width="100" sortable="custom" />
        <el-table-column prop="billing_cycle" label="计费周期" width="120" sortable="custom" />
        <el-table-column prop="next_renewal_date" label="下次续费日" width="140" sortable="custom" />
        <el-table-column prop="notes" label="备注" min-width="150" :sortable="false" show-overflow-tooltip />
        <el-table-column prop="created_at" label="创建时间" width="170" sortable="custom" />
        <el-table-column label="操作" width="180" fixed="right" :sortable="false">
          <template #default="scope">
            <el-button size="small" @click="handleEdit(scope.row)">
              <el-icon><Edit /></el-icon> 编辑
            </el-button>
            <el-popconfirm
              title="确定要删除此订阅吗？"
              confirm-button-text="确定"
              cancel-button-text="取消"
              @confirm="handleDelete(scope.row.id)"
            >
              <template #reference>
                <el-button size="small" type="danger">
                  <el-icon><Delete /></el-icon> 删除
                </el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-if="serviceSubscriptionStore.pagination.totalItems > 0"
        class="mt-4"
        background
        layout="total, sizes, prev, pager, next, jumper"
        :total="serviceSubscriptionStore.pagination.totalItems"
        :page-sizes="[10, 20, 50, 100]"
        :current-page="serviceSubscriptionStore.pagination.currentPage"
        :page-size="serviceSubscriptionStore.pagination.pageSize"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
      />
    </el-card>
  </div>
</template>

<script setup>
import { onMounted } from 'vue'; // Removed reactive, using ref for local renewalDateRange if needed
import { useRouter } from 'vue-router';
import { useServiceSubscriptionStore } from '@/stores/serviceSubscription';
import { usePlatformRegistrationStore } from '@/stores/platformRegistration';
import { ElIcon } from 'element-plus';
import { Plus, Edit, Delete } from '@element-plus/icons-vue';

const router = useRouter();
const serviceSubscriptionStore = useServiceSubscriptionStore();
const platformRegistrationStore = usePlatformRegistrationStore();

// const filters = reactive({...}); // Removed, use store directly

onMounted(async () => {
  // Fetch options for platform registration dropdown
  if (platformRegistrationStore.platformRegistrations.length === 0) {
    await platformRegistrationStore.fetchPlatformRegistrations({ page: 1, pageSize: 10000, orderBy: 'platform_name', sortDirection: 'asc' });
  }
  // Initial data fetch for the table using current store state
  serviceSubscriptionStore.fetchServiceSubscriptions();
});

// Removed local fetchData, applyFilters, resetFilters

const handlePlatformRegistrationFilterChange = (value) => {
  serviceSubscriptionStore.setFilter('platform_registration_id', value);
};

const handleStatusFilterChange = (value) => {
  serviceSubscriptionStore.setFilter('status', value);
};

const handleBillingCycleFilterChange = (value) => {
  serviceSubscriptionStore.setFilter('billing_cycle', value);
};

const handleRenewalDateChange = (dateRange) => {
  const start = dateRange && dateRange.length > 0 ? dateRange[0] : '';
  const end = dateRange && dateRange.length > 1 ? dateRange[1] : '';
  // Update store filters separately then fetch, or enhance setFilter to handle multiple
  serviceSubscriptionStore.filters.renewal_date_start = start;
  serviceSubscriptionStore.filters.renewal_date_end = end;
  serviceSubscriptionStore.pagination.currentPage = 1; // Reset page
  serviceSubscriptionStore.fetchServiceSubscriptions(); // Fetch with updated dates
};

const triggerFetchWithCurrentFilters = () => {
  // v-models have updated the store's filters (except date range which is handled by handleRenewalDateChange)
  // fetchServiceSubscriptions will use them. Reset to page 1.
  serviceSubscriptionStore.fetchServiceSubscriptions(1);
};

const triggerClearFilters = () => {
  serviceSubscriptionStore.clearFilters(); // This clears all filters in store and fetches
};

const handleSortChange = ({ prop, order }) => {
  const sortDirection = order === 'descending' ? 'desc' : 'asc';
  serviceSubscriptionStore.fetchServiceSubscriptions(
    1, // Reset to page 1 on sort change
    serviceSubscriptionStore.pagination.pageSize,
    { orderBy: prop, sortDirection }
  );
};

const handleAdd = () => {
  router.push({ name: 'ServiceSubscriptionCreate' }); 
};

const handleEdit = (row) => {
  router.push({ name: 'ServiceSubscriptionEdit', params: { id: row.id } });
};

const handleDelete = async (id) => {
  await serviceSubscriptionStore.deleteServiceSubscription(id);
  // Data is re-fetched by the store action on success
};

const handleSizeChange = (newSize) => {
  serviceSubscriptionStore.fetchServiceSubscriptions(1, newSize);
};

const handleCurrentChange = (newPage) => {
  serviceSubscriptionStore.fetchServiceSubscriptions(newPage, serviceSubscriptionStore.pagination.pageSize);
};
</script>

<style scoped>
.service-subscription-list-view {
  padding: 20px;
}
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-bottom: 10px; /* Add some padding below the header */
}
.filter-form {
  margin-bottom: 20px;
  display: flex; /* Use flexbox for better alignment */
  flex-wrap: wrap; /* Allow items to wrap to the next line */
  gap: 15px; /* Add gap between form items */
}
.filter-form .el-form-item {
  margin-right: 0; /* Remove default margin-right from el-form-item */
  margin-bottom: 0; /* Remove default margin-bottom from el-form-item */
}
.mt-4 {
  margin-top: 1.5rem;
}
</style>