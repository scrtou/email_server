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

      <el-form :inline="true" class="filter-form">
        <el-form-item label="邮箱账户">
          <el-select
            v-model="platformRegistrationStore.filters.email_account_id"
            placeholder="选择邮箱账户"
            clearable
            filterable
            @change="handleEmailAccountFilterChange"
            style="width: 240px;"
          >
            <el-option
              v-for="item in emailAccountStore.emailAccounts"
              :key="item.id"
              :label="item.email_address"
              :value="item.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="平台">
          <el-select
            v-model="platformRegistrationStore.filters.platform_id"
            placeholder="选择平台"
            clearable
            filterable
            @change="handlePlatformFilterChange"
            style="width: 240px;"
          >
            <el-option
              v-for="item in platformStore.platforms"
              :key="item.id"
              :label="item.name"
              :value="item.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="triggerFetchWithCurrentFilters">查询</el-button>
          <el-button @click="triggerClearFilters">重置</el-button>
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
import { onMounted } from 'vue';
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

// const filters = reactive({ // Removed, use store directly
//   emailAccountId: null,
//   platformId: null,
// });

onMounted(async () => {
  // Fetch options for select dropdowns
  // Consider if these lists are large, might need a paginated/searchable select component later
  if (emailAccountStore.emailAccounts.length === 0) { // Fetch only if not already populated
    await emailAccountStore.fetchEmailAccounts(
      1,
      10000,
      { orderBy: 'email_address', sortDirection: 'asc' },
      { provider: '', emailAddressSearch: '' } // Explicitly pass empty filters
    );
  }
  if (platformStore.platforms.length === 0) { // Fetch only if not already populated
    await platformStore.fetchPlatforms(1, 10000, { orderBy: 'name', sortDirection: 'asc' }); // Fetch a large number for dropdown
  }
  
  // Initial data fetch for the table, using current store state for filters, sort, pagination
  platformRegistrationStore.fetchPlatformRegistrations();
});

// Removed local fetchData, applyFilters, resetFilters as store handles this.

const handleEmailAccountFilterChange = (value) => {
  platformRegistrationStore.setFilter('email_account_id', value);
};

const handlePlatformFilterChange = (value) => {
  platformRegistrationStore.setFilter('platform_id', value);
};

const triggerFetchWithCurrentFilters = () => {
  // v-model has updated the store's filters.
  // fetchPlatformRegistrations will use them. Reset to page 1.
  platformRegistrationStore.fetchPlatformRegistrations(1);
};

const triggerClearFilters = () => {
  platformRegistrationStore.clearFilters(); // This clears filters in store and fetches
};


const handleSortChange = ({ prop, order }) => {
  const sortDirection = order === 'descending' ? 'desc' : 'asc';
  // Store's fetchPlatformRegistrations will use current filters
  platformRegistrationStore.fetchPlatformRegistrations(
    1, // Reset to page 1 on sort change
    platformRegistrationStore.pagination.pageSize,
    { orderBy: prop, sortDirection }
  );
};

// handleFilterChange is now split into specific handlers above

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
  // Store's fetchPlatformRegistrations will use current filters and sort
  platformRegistrationStore.fetchPlatformRegistrations(1, newSize);
};

const handleCurrentChange = (newPage) => {
  // Store's fetchPlatformRegistrations will use current filters and sort
  platformRegistrationStore.fetchPlatformRegistrations(newPage, platformRegistrationStore.pagination.pageSize);
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