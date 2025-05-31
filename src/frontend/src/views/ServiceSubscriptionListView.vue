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

      <el-form :inline="true" :model="filters" class="filter-form">
        <el-form-item label="平台注册信息">
          <el-select 
            v-model="filters.platformRegistrationId" 
            placeholder="选择平台注册信息" 
            clearable 
            filterable
            @change="applyFilters"
            style="width: 300px;"
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
          <el-select v-model="filters.status" placeholder="选择状态" clearable @change="applyFilters">
            <el-option label="活跃 (active)" value="active" />
            <el-option label="已取消 (cancelled)" value="cancelled" />
            <el-option label="试用 (free_trial)" value="free_trial" />
            <el-option label="已过期 (expired)" value="expired" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="applyFilters">查询</el-button>
          <el-button @click="resetFilters">重置</el-button>
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
import { onMounted, reactive } from 'vue';
import { useRouter } from 'vue-router';
import { useServiceSubscriptionStore } from '@/stores/serviceSubscription';
import { usePlatformRegistrationStore } from '@/stores/platformRegistration';
import { ElIcon } from 'element-plus';
import { Plus, Edit, Delete } from '@element-plus/icons-vue';

const router = useRouter();
const serviceSubscriptionStore = useServiceSubscriptionStore();
const platformRegistrationStore = usePlatformRegistrationStore();

const filters = reactive({
  platformRegistrationId: null,
  status: '',
});

onMounted(async () => {
  // Fetch platform registrations for the filter dropdown
  // Fetching all, or implement a paginated/searchable select if list is too long
  await platformRegistrationStore.fetchPlatformRegistrations({ page: 1, pageSize: 10000 }); 
  fetchData();
});

const fetchData = (
  page = serviceSubscriptionStore.pagination.currentPage,
  pageSize = serviceSubscriptionStore.pagination.pageSize,
  sortOptions = { orderBy: serviceSubscriptionStore.sort.orderBy, sortDirection: serviceSubscriptionStore.sort.sortDirection }
) => {
  const params = {
    page,
    pageSize,
    orderBy: sortOptions.orderBy,
    sortDirection: sortOptions.sortDirection
  };
  if (filters.platformRegistrationId) {
    params.platform_registration_id = filters.platformRegistrationId;
  }
  if (filters.status) {
    params.status = filters.status;
  }
  serviceSubscriptionStore.fetchServiceSubscriptions(params);
};

const handleSortChange = ({ prop, order }) => {
  const sortDirection = order === 'descending' ? 'desc' : 'asc';
  // Note: Backend currently only supports sorting by ServiceSubscription's own fields.
  // If prop is 'platform_name' or 'email_address', it will default to 'created_at' on backend.
  fetchData(1, serviceSubscriptionStore.pagination.pageSize, { orderBy: prop, sortDirection });
};

const applyFilters = () => {
    fetchData(1, serviceSubscriptionStore.pagination.pageSize);
}

const resetFilters = () => {
    filters.platformRegistrationId = null;
    filters.status = '';
    fetchData(1, serviceSubscriptionStore.pagination.pageSize);
}

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
  fetchData(1, newSize); 
};

const handleCurrentChange = (newPage) => {
  fetchData(newPage, serviceSubscriptionStore.pagination.pageSize);
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
}
.filter-form {
  margin-bottom: 20px;
}
.mt-4 {
  margin-top: 1.5rem;
}
</style>