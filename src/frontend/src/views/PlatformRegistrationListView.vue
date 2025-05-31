<template>
  <div class="platform-registration-list-view">
    <el-card class="box-card">
      <template #header>
        <div class="card-header">
          <span class="card-title">平台注册信息管理</span>
          <el-button type="primary" :icon="Plus" @click="handleAdd">
            添加注册信息
          </el-button>
        </div>
      </template>

      <el-form :inline="true" :model="platformRegistrationStore.filters" class="filter-form">
        <el-row :gutter="20">
          <el-col :xs="24" :sm="12" :md="8" :lg="6">
            <el-form-item label="邮箱账户">
              <el-select
                v-model="platformRegistrationStore.filters.email_account_id"
                placeholder="选择邮箱账户"
                clearable
                filterable
                @change="handleEmailAccountFilterChange"
                class="full-width-select"
              >
                <el-option
                  v-for="item in emailAccountStore.emailAccounts"
                  :key="item.id"
                  :label="item.email_address"
                  :value="item.id"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :xs="24" :sm="12" :md="8" :lg="6">
            <el-form-item label="平台">
              <el-select
                v-model="platformRegistrationStore.filters.platform_id"
                placeholder="选择平台"
                clearable
                filterable
                @change="handlePlatformFilterChange"
                class="full-width-select"
              >
                <el-option
                  v-for="item in platformStore.platforms"
                  :key="item.id"
                  :label="item.name"
                  :value="item.id"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :xs="24" :sm="24" :md="8" :lg="6">
            <el-form-item>
              <el-button type="primary" @click="triggerFetchWithCurrentFilters">查询</el-button>
              <el-button @click="triggerClearFilters">重置</el-button>
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>

      <el-table
        :data="platformRegistrationStore.platformRegistrations"
        v-loading="platformRegistrationStore.loading"
        style="width: 100%"
        @sort-change="handleSortChange"
        :default-sort="{ prop: platformRegistrationStore.sort.orderBy, order: platformRegistrationStore.sort.sortDirection === 'desc' ? 'descending' : 'ascending' }"
        border
        stripe
        resizable
      >
        <el-table-column prop="email_address" label="邮箱账户" min-width="180" sortable="custom" show-overflow-tooltip />
        <el-table-column prop="platform_name" label="平台" min-width="120" sortable="custom" show-overflow-tooltip />
        <el-table-column prop="login_username" label="登录用户名/ID" min-width="170" sortable="custom" show-overflow-tooltip />
        <el-table-column prop="notes" label="备注" min-width="200" sortable="custom" show-overflow-tooltip />
        <el-table-column prop="created_at" label="创建时间" width="200" sortable="custom" />
        <el-table-column prop="updated_at" label="更新时间" width="200" sortable="custom" />
        <el-table-column label="操作" width="140" fixed="right" :sortable="false">
          <template #default="scope">
            <el-button link type="primary" :icon="Edit" @click="handleEdit(scope.row)">编辑</el-button>
            <el-button link type="danger" :icon="Delete" @click="confirmDeleteRegistration(scope.row.id)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-if="platformRegistrationStore.pagination.totalItems > 0"
        class="pagination-container"
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
import { ElMessageBox } from 'element-plus';
import { Plus, Edit, Delete } from '@element-plus/icons-vue'; // 移除 ElIcon 导入，直接在模板中使用图标组件

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
  padding: 20px; /* 增加整体内边距 */
  background-color: #f0f2f5; /* Light grey background for the whole view */
}

.box-card {
  border-radius: 8px; /* 卡片圆角 */
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.05); /* 增加阴影效果 */
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 0;
}

.card-title {
  font-size: 20px; /* 标题字体大小 */
  font-weight: bold; /* 标题加粗 */
  color: #303133; /* 标题颜色 */
}

.filter-form {
  margin-bottom: 20px; /* 筛选表单底部间距 */
  padding: 16px; /* 筛选表单内边距 */
  background-color: #f9fafc; /* 筛选表单背景色 */
  border-radius: 4px; /* 筛选表单圆角 */
}

.filter-form .el-form-item {
  margin-bottom: 16px; /* 表单项底部间距 */
  margin-right: 0; /* 移除默认右边距 */
  width: 100%; /* 确保表单项在col中占满宽度 */
}

.full-width-select {
  width: 100%; /* 下拉选择框占满宽度 */
}


/* 表格核心样式 - 与 EmailAccountListView 统一 */
:deep(.el-table) {
  margin-top: 20px;
  border-radius: 8px;
  overflow: hidden;
  border: none; /* 移除 Element Plus 默认边框 */
}
:deep(.el-table::before) { /* 移除表格底部默认横线 */
  height: 0;
}

/* 表格行 hover 效果 */
:deep(.el-table .el-table__row:hover > td) {
  background: linear-gradient(135deg, var(--color-primary-50), rgba(59, 130, 246, 0.05));
}

/* 表格数据单元格 (td) */
:deep(.el-table td.el-table__cell) {
  padding: 4px 8px; /* 更紧凑的内边距 */
  border-bottom: 1px solid var(--color-gray-100);
  border-right: none; /* 移除竖线 */
  line-height: 1.4;
}

/* 表格头部单元格 (th) */
:deep(.el-table th.el-table__cell) {
  padding: 4px 8px; /* 更紧凑的内边距 */
  background: linear-gradient(135deg, var(--color-gray-50), var(--color-gray-100));
  color: var(--color-gray-700);
  font-weight: var(--font-semibold);
  border-bottom: 2px solid var(--color-gray-200);
  border-right: none; /* 移除竖线 */
  line-height: 1.4;
}

/* --- 操作列特定样式 (与 EmailAccountListView 统一) --- */
/* 操作列按钮容器 */
:deep(.el-table .el-table__cell:last-child .cell) {
  display: flex !important;
  justify-content: center !important;
  align-items: center !important;
  gap: 2px !important;
}
/* 操作列按钮本身 */
:deep(.el-table td.el-table__cell .el-button) {
  margin-right: 0px !important;
  margin-bottom: 0 !important;
  padding: 4px 8px !important;
  font-size: 12px !important;
  height: 28px !important;
  line-height: 1.2 !important;
  display: inline-block !important;
}
/* 操作列最后一个按钮 */
:deep(.el-table td.el-table__cell .el-button:last-child) {
  margin-right: 0 !important;
}
/* 操作列单元格本身 */
:deep(.el-table td.el-table__cell:last-child) {
  padding: 8px 4px !important;
  white-space: nowrap !important;
}

.pagination-container {
  margin-top: 20px; /* 分页器顶部间距 */
  display: flex;
  justify-content: flex-end; /* 分页器右对齐 */
}

/* 响应式调整 */
@media (max-width: 768px) {
  .filter-form .el-form-item {
    margin-bottom: 10px; /* 移动端表单项间距 */
  }
  .card-header {
    flex-direction: column; /* 移动端标题和按钮垂直堆叠 */
    align-items: flex-start;
  }
  .card-header .el-button {
    margin-top: 10px; /* 移动端按钮顶部间距 */
    width: 100%; /* 按钮占满宽度 */
  }
  .pagination-container {
    justify-content: center; /* 移动端分页器居中 */
  }
}
</style>