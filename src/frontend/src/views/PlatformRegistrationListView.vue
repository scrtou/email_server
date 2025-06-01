<template>
  <div class="platform-registration-list-view">
    <el-card class="box-card">
      <template #header>
        <div class="card-header">
          <span class="card-title">平台注册信息管理</span>
          <el-button type="primary" :icon="Plus" @click="handleAdd">
            添加注册信息
          </el-button>
          <el-button type="success" :icon="Upload" @click="handleImportBitwarden" style="margin-left: 10px;">
            导入 Bitwarden 数据
          </el-button>
        </div>
      </template>

      <el-form :inline="true" :model="platformRegistrationStore.filters" class="filter-form">
        <el-row :gutter="20">
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
            <el-form-item label="用户名">
              <el-select
                v-model="platformRegistrationStore.filters.login_username"
                placeholder="搜索用户名"
                clearable
                filterable
                @change="triggerFetchWithCurrentFilters"
                @clear="triggerFetchWithCurrentFilters"
                style="width: 180px;"
              >
                <el-option
                  v-for="item in usernameOptions"
                  :key="item"
                  :label="item"
                  :value="item"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :xs="24" :sm="24" :md="8" :lg="6">
            <el-form-item>
              <el-button type="primary" @click="triggerFetchWithCurrentFilters" :loading="isQuerying">查询</el-button>
              <el-button @click="triggerClearFilters" :loading="isResetting">重置</el-button>
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
      <div class="table-container" style="flex-grow: 1; overflow-y: scroll;">
        <el-table
          :data="platformRegistrationStore.platformRegistrations"
          v-loading="platformRegistrationStore.loading"
          style="width: 100%"
          height="100%"
          @sort-change="handleSortChange"
          :default-sort="{ prop: platformRegistrationStore.sort.orderBy, order: platformRegistrationStore.sort.sortDirection === 'desc' ? 'descending' : 'ascending' }"
          border
          stripe
          resizable
        >
        <el-table-column prop="platform_name" label="平台" min-width="120" sortable="custom" show-overflow-tooltip />
        <el-table-column prop="login_username" label="用户名/ID" min-width="130" sortable="custom" show-overflow-tooltip />
        <el-table-column prop="email_address" label="邮箱账户" min-width="180" sortable="custom" show-overflow-tooltip />
<el-table-column prop="phone_number" label="手机号" min-width="150" sortable="custom" show-overflow-tooltip />
        <el-table-column prop="notes" label="备注" min-width="200" sortable="custom" show-overflow-tooltip />
        <el-table-column prop="created_at" label="创建时间" width="200" sortable="custom" />
        <el-table-column prop="updated_at" label="更新时间" width="200" sortable="custom" />
        <el-table-column label="操作" width="140" fixed="right" :sortable="false">
          <template #default="scope">
            <el-button link type="primary" :icon="Edit" @click="handleEdit(scope.row)">编辑</el-button>
            <el-button link type="danger" :icon="Delete" @click="confirmDeleteRegistration(scope.row.id)" :loading="platformRegistrationStore.loading">删除</el-button>
          </template>
        </el-table-column>
        </el-table>
      </div>

      <el-pagination
        v-if="platformRegistrationStore.pagination.totalItems > 0"
        class="mt-4"
        background
        layout="total, prev, pager, next, jumper"
        :total="platformRegistrationStore.pagination.totalItems"
        :current-page="platformRegistrationStore.pagination.currentPage"
        :page-size="pageSize.value"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
      />
    </el-card>

    <ModalDialog
      :visible="showModal"
      :title="modalTitle"
      @confirm="() => platformRegistrationFormRef?.triggerSubmit()"
      @cancel="closeModal"
      width="60%"
      :confirm-button-text="isEditMode ? '保存更新' : '立即创建'"
      :show-confirm-button="false"
      :show-cancel-button="false"
    >
      <!-- Form content remains the same -->
      <PlatformRegistrationForm
        ref="platformRegistrationFormRef"
        v-if="showModal"
        :platform-registration="currentRegistration"
        :is-edit="isEditMode"
        @submit-form="handleFormSubmit"
        @cancel="closeModal"
      />
      <!-- Custom Footer with loading state -->
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="closeModal">取消</el-button>
          <el-button
            type="primary"
            @click="() => platformRegistrationFormRef?.triggerSubmit()"
            :loading="platformRegistrationStore.loading"
            :disabled="platformRegistrationStore.loading"
          >
            {{ isEditMode ? '保存更新' : '立即创建' }}
          </el-button>
        </div>
      </template>
    </ModalDialog>
  </div>
</template>

<script setup>
import { onMounted, ref, computed } from 'vue';
import { useRouter } from 'vue-router'; // Added back
const MIN_LOADING_TIME = 300; // 最小加载时间，单位毫秒
import { usePlatformRegistrationStore } from '@/stores/platformRegistration';
import { useEmailAccountStore } from '@/stores/emailAccount';
import { usePlatformStore } from '@/stores/platform';
import { ElMessageBox, ElMessage } from 'element-plus';
import { Plus, Edit, Delete, Upload } from '@element-plus/icons-vue'; // Added Upload icon
import ModalDialog from '@/components/ui/ModalDialog.vue';
import PlatformRegistrationForm from '@/components/forms/PlatformRegistrationForm.vue';

const router = useRouter(); // Added back
const platformRegistrationStore = usePlatformRegistrationStore();
const emailAccountStore = useEmailAccountStore();
const platformStore = usePlatformStore();
const pageSize = ref(platformRegistrationStore.pagination.pageSize || 10);

// Modal state
const showModal = ref(false);
const modalTitle = ref('');
const currentRegistration = ref(null);
const isEditMode = ref(false); // This is already used for the form's :is-edit prop
const platformRegistrationFormRef = ref(null); // Ref for the form component
const isQuerying = ref(false); // 用于“查询”按钮的 loading 状态
const isResetting = ref(false); // 用于“重置”按钮的 loading 状态

// const filters = reactive({ // Removed, use store directly
//   emailAccountId: null,
//   platformId: null,
// });

const usernameOptions = computed(() => {
  const usernames = platformRegistrationStore.platformRegistrations
    .map(pr => pr.login_username)
    .filter(username => username && username.trim() !== ''); // Filter out empty or null usernames
  return [...new Set(usernames)].sort();
});

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
  // Always fetch platforms for the dropdown with a large page size (e.g., 200)
  // to ensure all platforms are available, overriding any previously fetched list
  // that might have been paginated with a smaller pageSize (like the default 8).
  // The backend API's max pageSize is 100. Requesting 200 here aims to get
  // as many as possible, ideally all if the backend handles pageSize > max gracefully,
  // or at least the maximum of 100. This directly addresses the issue of the
  // dropdown showing only 8 items due to an earlier fetch with a small pageSize.
  await platformStore.fetchPlatforms(1, 200, { orderBy: 'name', sortDirection: 'asc' });
  
  // Initial data fetch for the table, using current store state for filters, sort, pagination
  platformRegistrationStore.fetchPlatformRegistrations(
    platformRegistrationStore.pagination.currentPage,
    pageSize.value, // Use the local pageSize ref
    platformRegistrationStore.sort
  );
});

// Removed local fetchData, applyFilters, resetFilters as store handles this.

const handleEmailAccountFilterChange = (value) => {
  platformRegistrationStore.setFilter('email_account_id', value);
};

const handlePlatformFilterChange = (value) => {
  platformRegistrationStore.setFilter('platform_id', value);
};

const triggerFetchWithCurrentFilters = async () => {
  if (isQuerying.value) return;
  isQuerying.value = true;
  const startTime = Date.now();

  try {
    // v-model has updated the store's filters.
    // fetchPlatformRegistrations will use them. Reset to page 1.
    await platformRegistrationStore.fetchPlatformRegistrations(1, pageSize.value);
  } finally {
    const elapsedTime = Date.now() - startTime;
    if (elapsedTime < MIN_LOADING_TIME) {
      setTimeout(() => {
        isQuerying.value = false;
      }, MIN_LOADING_TIME - elapsedTime);
    } else {
      isQuerying.value = false;
    }
  }
};

const triggerClearFilters = async () => {
  if (isResetting.value) return;
  isResetting.value = true;
  const startTime = Date.now();
  try {
    await platformRegistrationStore.clearFilters(); // This clears filters in store and fetches
  } finally {
    const elapsedTime = Date.now() - startTime;
    if (elapsedTime < MIN_LOADING_TIME) {
      setTimeout(() => {
        isResetting.value = false;
      }, MIN_LOADING_TIME - elapsedTime);
    } else {
      isResetting.value = false;
    }
  }
};


const handleSortChange = ({ prop, order }) => {
  const sortDirection = order === 'descending' ? 'desc' : 'asc';
  // Store's fetchPlatformRegistrations will use current filters
  platformRegistrationStore.fetchPlatformRegistrations(
    1, // Reset to page 1 on sort change
    pageSize.value,
    { orderBy: prop, sortDirection }
  );
};

// handleFilterChange is now split into specific handlers above

const openModalForCreate = () => {
  isEditMode.value = false;
  modalTitle.value = '添加平台注册信息';
  currentRegistration.value = null; // Or an empty object with default structure if your form expects it
  showModal.value = true;
};

const openModalForEdit = (registration) => {
  isEditMode.value = true;
  modalTitle.value = '编辑平台注册信息';
  currentRegistration.value = { ...registration }; // Pass a copy to avoid direct mutation
  showModal.value = true;
};

const closeModal = () => {
  showModal.value = false;
  currentRegistration.value = null;
};

const handleFormSubmit = async (eventData) => {
  // eventData is { payload, id?, isEdit, useByNameApi? }
  const { payload, id, isEdit: formIsEdit, useByNameApi } = eventData;
  let success = false;

  if (formIsEdit) { // Corresponds to isEditMode.value when the form was opened
    if (!id) {
      ElMessage.error('编辑错误：缺少注册信息ID');
      return;
    }
    success = await platformRegistrationStore.updatePlatformRegistration(id, payload);
  } else { // Create mode
    if (useByNameApi) {
      success = await platformRegistrationStore.createPlatformRegistrationByName(payload);
    } else {
      success = await platformRegistrationStore.createPlatformRegistration(payload);
    }
  }

  if (success) {
    closeModal();
    // ElMessage is handled by store actions now
    // Refresh data
    await platformRegistrationStore.fetchPlatformRegistrations(
      platformRegistrationStore.pagination.currentPage,
      pageSize.value,
      platformRegistrationStore.sort,
      platformRegistrationStore.filters
    );
  }
};


// Original handleAdd and handleEdit are replaced by openModalForCreate and openModalForEdit
const handleAdd = () => {
  // router.push({ name: 'PlatformRegistrationCreate' });
  openModalForCreate();
};

const handleEdit = (row) => {
  // router.push({ name: 'PlatformRegistrationEdit', params: { id: row.id } });
  openModalForEdit(row);
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
  pageSize.value = newSize;
  platformRegistrationStore.fetchPlatformRegistrations(1, pageSize.value);
};

const handleCurrentChange = (newPage) => {
  // Store's fetchPlatformRegistrations will use current filters and sort
  platformRegistrationStore.fetchPlatformRegistrations(newPage, pageSize.value);
};

const handleImportBitwarden = () => {
  router.push('/import/bitwarden');
};
</script>

<style scoped>
.platform-registration-list-view {
  padding: 20px; /* 增加整体内边距 */
  background-color: #f0f2f5; /* Light grey background for the whole view */
  display: flex;
  flex-direction: column;
  height: 100%;
  box-sizing: border-box;
}

.box-card {
  border-radius: 8px; /* 卡片圆角 */
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.05); /* 增加阴影效果 */
  display: flex;
  flex-direction: column;
  flex-grow: 1;
  overflow: hidden;
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
  margin-top: 0px;
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
  padding: 4px 8px; /* 增加垂直内边距 */
  border-bottom: 1px solid var(--color-gray-100);
  border-right: none; /* 移除竖线 */
  line-height: 1.4;
}

/* 表格头部单元格 (th) */
:deep(.el-table th.el-table__cell) {
  padding: 4px 8px; /* 增加垂直内边距 */
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

/* Pagination styles moved to utilities.css */
/* .pagination-container class is still used in the template, but styled globally now */

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