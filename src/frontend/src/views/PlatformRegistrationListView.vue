<template>
  <div class="platform-registration-list-view">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>平台注册信息管理</span>
          <el-button type="primary" @click="handleAdd">
            <el-icon><Plus /></el-icon> 添加注册信息
          </el-button>
        </div>
      </template>

      <el-form :inline="true" :model="filters" class="filter-form">
        <el-form-item label="邮箱账户">
          <el-select v-model="filters.emailAccountId" placeholder="选择邮箱账户" clearable @change="handleFilterChange" style="width: 240px;">
            <el-option
              v-for="item in emailAccountStore.emailAccounts"
              :key="item.id"
              :label="item.email_address"
              :value="item.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="平台">
          <el-select v-model="filters.platformId" placeholder="选择平台" clearable @change="handleFilterChange" style="width: 240px;">
            <el-option
              v-for="item in platformStore.platforms"
              :key="item.id"
              :label="item.name"
              :value="item.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="applyFilters">查询</el-button>
          <el-button @click="resetFilters">重置</el-button>
        </el-form-item>
      </el-form>

      <el-table
        :data="platformRegistrationStore.platformRegistrations"
        v-loading="platformRegistrationStore.loading"
        style="width: 100%"
        @sort-change="handleSortChange"
        :default-sort="{ prop: platformRegistrationStore.sort.orderBy, order: platformRegistrationStore.sort.sortDirection === 'desc' ? 'descending' : 'ascending' }"
      >
        <el-table-column prop="email_address" label="邮箱账户" min-width="200" sortable="custom" />
        <!-- <el-table-column prop="id" label="ID" width="80" /> -->
        <el-table-column prop="platform_name" label="平台" min-width="150" sortable="custom" />
        <el-table-column prop="login_username" label="登录用户名/ID" min-width="180" sortable="custom" />
        <el-table-column prop="notes" label="备注" min-width="200" sortable="custom" show-overflow-tooltip />
        <el-table-column prop="created_at" label="创建时间" width="180" sortable="custom" />
        <el-table-column prop="updated_at" label="更新时间" width="180" sortable="custom" />
        <el-table-column label="操作" width="180" fixed="right" :sortable="false">
          <template #default="scope">
            <el-button size="small" @click="handleEdit(scope.row)">
              <el-icon><Edit /></el-icon> 编辑
            </el-button>
            <el-button size="small" type="danger" @click="confirmDeleteRegistration(scope.row.id)">
              <el-icon><Delete /></el-icon> 删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-if="platformRegistrationStore.pagination.totalItems > 0"
        class="mt-4"
        background
        layout="total, sizes, prev, pager, next, jumper"
        :total="platformRegistrationStore.pagination.totalItems"
        :page-sizes="[10, 20, 50, 100]"
        :current-page="platformRegistrationStore.pagination.currentPage"
        :page-size="platformRegistrationStore.pagination.pageSize"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
      />
    </el-card>
  </div>
</template>

<script setup>
import { onMounted, reactive } from 'vue';
import { useRouter } from 'vue-router';
import { usePlatformRegistrationStore } from '@/stores/platformRegistration';
import { useEmailAccountStore } from '@/stores/emailAccount';
import { usePlatformStore } from '@/stores/platform';
import { ElIcon, ElMessageBox } from 'element-plus';
import { Plus, Edit, Delete } from '@element-plus/icons-vue';

const router = useRouter();
const platformRegistrationStore = usePlatformRegistrationStore();
const emailAccountStore = useEmailAccountStore();
const platformStore = usePlatformStore();

const filters = reactive({
  emailAccountId: null,
  platformId: null,
});

onMounted(async () => {
  await emailAccountStore.fetchEmailAccounts(1, 1000); // Fetch all for dropdown, or implement paginated select
  await platformStore.fetchPlatforms(1, 1000); // Fetch all for dropdown
  fetchData();
});

const fetchData = (
  page = platformRegistrationStore.pagination.currentPage,
  pageSize = platformRegistrationStore.pagination.pageSize,
  sortOptions = { orderBy: platformRegistrationStore.sort.orderBy, sortDirection: platformRegistrationStore.sort.sortDirection }
) => {
  const params = {
    page,
    pageSize,
    orderBy: sortOptions.orderBy,
    sortDirection: sortOptions.sortDirection
  };
  if (filters.emailAccountId) {
    params.email_account_id = filters.emailAccountId;
  }
  if (filters.platformId) {
    params.platform_id = filters.platformId;
  }
  platformRegistrationStore.fetchPlatformRegistrations(params);
};

const handleSortChange = ({ prop, order }) => {
  const sortDirection = order === 'descending' ? 'desc' : 'asc';
  // Note: Backend currently only supports sorting by PlatformRegistration's own fields.
  // If prop is 'email_address' or 'platform_name', it will default to 'created_at' on backend.
  // For a better UX, we might disable sorting on these columns or adjust backend.
  // For now, we pass the prop as is.
  fetchData(1, platformRegistrationStore.pagination.pageSize, { orderBy: prop, sortDirection });
};

const handleFilterChange = () => {
 // fetchData(1, platformRegistrationStore.pagination.pageSize); // Optionally auto-filter on change
};

const applyFilters = () => {
    fetchData(1, platformRegistrationStore.pagination.pageSize);
}

const resetFilters = () => {
    filters.emailAccountId = null;
    filters.platformId = null;
    fetchData(1, platformRegistrationStore.pagination.pageSize);
}

const handleAdd = () => {
  router.push({ name: 'PlatformRegistrationCreate' }); 
};

const handleEdit = (row) => {
  router.push({ name: 'PlatformRegistrationEdit', params: { id: row.id } });
};

const confirmDeleteRegistration = (id) => {
  ElMessageBox.confirm(
    '删除此平台注册信息将同时删除其下所有关联的服务订阅数据。是否确认删除？',
    '警告',
    {
      confirmButtonText: '确定删除',
      cancelButtonText: '取消',
      type: 'warning',
    }
  )
    .then(async () => {
      await platformRegistrationStore.deletePlatformRegistration(id);
      // Data is re-fetched by the store action on success
    })
    .catch(() => {
      // User cancelled
    });
};

const handleSizeChange = (newSize) => {
  fetchData(1, newSize); 
};

const handleCurrentChange = (newPage) => {
  fetchData(newPage, platformRegistrationStore.pagination.pageSize);
};
</script>

<style scoped>
.platform-registration-list-view {
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