
<script setup>
import { onMounted, ref, reactive } from 'vue';
// import { useRouter } from 'vue-router'; // No longer needed for add/edit
import { useEmailAccountStore } from '@/stores/emailAccount';
import { ElMessage, ElMessageBox, ElButton } from 'element-plus'; // ElDialog might not be directly used if ModalDialog handles it
import { Plus, Edit, Delete, View as ViewIcon } from '@element-plus/icons-vue';
import AssociatedInfoDialog from '@/components/AssociatedInfoDialog.vue';
import ModalDialog from '@/components/ui/ModalDialog.vue'; // Import ModalDialog
import EmailAccountForm from '@/components/forms/EmailAccountForm.vue'; // Import EmailAccountForm

// const router = useRouter(); // Keep for other navigation if any, or remove if not used at all
const emailAccountStore = useEmailAccountStore();

// --- Modal Dialog State for EmailAccountForm ---
const emailAccountFormDialog = reactive({
  visible: false,
  isEditMode: false,
  title: '',
  currentAccount: null, // Store the account being edited
});
// --- End Modal Dialog State ---

const loading = ref(false); // Define loading state for the view
const MIN_LOADING_TIME = 300; // Minimum loading time in milliseconds
const isQuerying = ref(false);
const isResetting = ref(false);

// const providerFilter = ref(emailAccountStore.filters.provider || ''); // Removed, use store directly

const currentEmailAccountForDialog = ref(null); // To store the email account context for pagination

const associatedInfoDialog = reactive({
  visible: false,
  title: '',
  items: [],
  layout: [
    { label: '平台名称', prop: 'platform_name', minWidth: '150px' },
    { label: '平台网址', prop: 'platform_website_url', type: 'link', minWidth: '200px' },
    { label: '注册备注', prop: 'registration_notes', minWidth: '200px' },
  ],
  pagination: {
    currentPage: 1,
    pageSize: 5, // Default page size for dialog
    totalItems: 0,
  },
  loading: false,
});

onMounted(() => {
  // emailAccountStore.fetchEmailAccounts(); // Fetch with current store state (page, size, sort, filters)
  // If filters might be pre-populated from elsewhere (e.g. URL params in future), ensure they are set before first fetch
  // For now, direct call is fine as store initializes filters.
  // If store's filters.provider is already set (e.g. from a previous session if persisted),
  // this will use it.
  emailAccountStore.fetchEmailAccounts(
    emailAccountStore.pagination.currentPage,
    emailAccountStore.pagination.pageSize,
    { orderBy: emailAccountStore.sort.orderBy, sortDirection: emailAccountStore.sort.sortDirection }
    // Filters are now part of the store's fetchEmailAccounts internal logic
  );
  emailAccountStore.fetchUniqueProviders(); // Fetch providers for the dropdown
});

// Removed local fetchData, applyFilters, clearFilters as store handles this now.

// const triggerSetFilter = () => { // Replaced by handleProviderFilterChange
//   emailAccountStore.setFilter('provider', emailAccountStore.filters.provider);
// };

// const triggerClearFilterAndFetch = () => { // Handled by el-select clearable and change
//   emailAccountStore.setFilter('provider', '');
// }

const handleProviderFilterChange = (value) => {
  // The v-model on el-select updates emailAccountStore.filters.provider directly.
  // The setFilter action in the store will handle fetching.
  emailAccountStore.setFilter('provider', value || ''); // Ensure empty string if null/undefined from clearable
  // Also trigger a general fetch if other filters might be active
  emailAccountStore.fetchEmailAccounts(1, emailAccountStore.pagination.pageSize, emailAccountStore.sort, emailAccountStore.filters);
};

const handleEmailAddressSearchChange = (value) => {
  emailAccountStore.setFilter('emailAddressSearch', value || '');
  // Also trigger a general fetch
  emailAccountStore.fetchEmailAccounts(1, emailAccountStore.pagination.pageSize, emailAccountStore.sort, emailAccountStore.filters);
};

const triggerApplyAllFilters = async () => {
  isQuerying.value = true;
  const startTime = Date.now();
  try {
    // This is for a dedicated "Query/Search" button if present
    // It ensures that the current state of all filters in the store is used to fetch.
    // The individual filter handlers already call setFilter which then calls fetch.
    // So this button might be redundant if instant filtering on change is preferred.
    // If used, it should fetch with all current filters.
    await emailAccountStore.fetchEmailAccounts(1, emailAccountStore.pagination.pageSize, emailAccountStore.sort, emailAccountStore.filters);
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

const triggerClearAllFilters = async () => {
  isResetting.value = true;
  const startTime = Date.now();
  try {
    emailAccountStore.clearFilters(); // This will clear all filters
    // After clearing, explicitly call fetch to ensure UI updates
    await emailAccountStore.fetchEmailAccounts(1, emailAccountStore.pagination.pageSize, emailAccountStore.sort, emailAccountStore.filters);
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

// Watch for external changes to store filters (e.g. if they were persisted and reloaded)
// This might not be strictly necessary if all filter changes go through setFilter/clearFilters actions
// and v-model directly binds to store.filters.provider.
// However, keeping it can be a safeguard or useful if store.filters.provider could be changed externally
// without calling setFilter (which would be an anti-pattern).
// For now, let's assume direct v-model binding and action calls are sufficient.
// watch(() => emailAccountStore.filters.provider, (newProvider) => {
//   // This watcher might cause a loop if not handled carefully with setFilter
//   // Let's rely on explicit calls to setFilter / clearFilters for now.
// });


const handleSortChange = ({ prop, order }) => {
  const sortDirection = order === 'descending' ? 'desc' : 'asc';
  // Store's fetchEmailAccounts will use current filters
  emailAccountStore.fetchEmailAccounts(1, emailAccountStore.pagination.pageSize, { orderBy: prop, sortDirection });
};

const handleAddEmailAccount = () => {
  // router.push({ name: 'EmailAccountCreate' }); // Replaced by dialog
  emailAccountFormDialog.isEditMode = false;
  emailAccountFormDialog.title = '添加邮箱账户';
  associatedInfoDialog.visible = false; // 确保关联信息弹窗已关闭
  emailAccountFormDialog.currentAccount = null; // Clear any previous edit data
  emailAccountFormDialog.visible = true;
};

const handleEdit = (row) => {
  // router.push({ name: 'EmailAccountEdit', params: { id: row.id } }); // Replaced by dialog
  associatedInfoDialog.visible = false; // 确保关联信息弹窗已关闭
  emailAccountFormDialog.isEditMode = true;
  emailAccountFormDialog.title = '编辑邮箱账户';
  // Create a shallow copy for editing to avoid direct mutation of the list item
  // Deep copy might be needed if form internally mutates nested objects of currentAccount
  emailAccountFormDialog.currentAccount = { ...row };
  emailAccountFormDialog.visible = true;
};

const handleFormSubmit = async (payloadFromForm) => {
  // payloadFromForm is the object emitted by EmailAccountForm's submit-form event
  loading.value = true; // Consider adding a loading state to the view if not already present
  let success = false;
  if (emailAccountFormDialog.isEditMode && emailAccountFormDialog.currentAccount && emailAccountFormDialog.currentAccount.id) {
    success = await emailAccountStore.updateEmailAccount(emailAccountFormDialog.currentAccount.id, payloadFromForm);
  } else if (!emailAccountFormDialog.isEditMode) {
    success = await emailAccountStore.createEmailAccount(payloadFromForm);
  } else {
    ElMessage.error('操作失败：无法确定是新增还是编辑模式，或编辑ID丢失。');
    loading.value = false;
    return;
  }
  loading.value = false;
  if (success) {
    emailAccountFormDialog.visible = false;
    // Data is refreshed by store actions, or we can call fetch here if needed.
    // Assuming store actions (create/update) handle list refresh.
    // If not, uncomment and adjust:
    // emailAccountStore.fetchEmailAccounts(
    //   emailAccountStore.pagination.currentPage,
    //   emailAccountStore.pagination.pageSize,
    //   emailAccountStore.sort,
    //   emailAccountStore.filters
    // );
  }
  // Error messages are handled by the store actions
};

const handleFormCancel = () => {
  emailAccountFormDialog.visible = false;
};


const confirmDeleteEmailAccount = (id) => {
  ElMessageBox.confirm(
    '删除此邮箱账户将同时删除其下所有关联的平台注册信息以及这些平台注册信息下的服务订阅数据。是否确认删除？',
    '警告',
    {
      confirmButtonText: '确定删除',
      cancelButtonText: '取消',
      type: 'warning',
    }
  )
    .then(async () => {
      const success = await emailAccountStore.deleteEmailAccount(id);
      if (success) {
        // Data is re-fetched by the store action on success.
        // The store's deleteEmailAccount should ideally handle fetching the correct page.
        // If not, we might need to call fetchEmailAccounts here.
        // The store's fetchEmailAccounts will use current filters.
        // Let's assume the store's delete action correctly refreshes.
        // If current page becomes empty, the store's fetch in delete should handle it or we adjust here.
        // For now, relying on store's delete action to call fetchEmailAccounts appropriately.
      }
    })
    .catch(() => {
      // User cancelled
    });
};

const handleSizeChange = (newSize) => {
  // Store's fetchEmailAccounts will use current filters and sort
  // pageSize.value = newSize; // No longer needed to update local ref
  emailAccountStore.fetchEmailAccounts(1, newSize, emailAccountStore.sort, emailAccountStore.filters); // Reset to page 1 when size changes
};

const handleCurrentChange = (newPage) => {
  // Store's fetchEmailAccounts will use current filters and sort
  emailAccountStore.fetchEmailAccounts(newPage, emailAccountStore.pagination.pageSize, emailAccountStore.sort, emailAccountStore.filters);
};

const fetchAssociatedPlatformsData = async (emailAccountId, page = 1, pageSize = 5) => {
  associatedInfoDialog.loading = true;
  try {
    const result = await emailAccountStore.fetchAssociatedPlatformRegistrations(emailAccountId, { page, pageSize });
    associatedInfoDialog.items = result.data;
    associatedInfoDialog.pagination.currentPage = result.meta.current_page;
    associatedInfoDialog.pagination.pageSize = result.meta.page_size;
    associatedInfoDialog.pagination.totalItems = result.meta.total_items;
  } catch (error) {
    // Error is handled by the store and ElMessage
    associatedInfoDialog.items = [];
    associatedInfoDialog.pagination.totalItems = 0;
  } finally {
    associatedInfoDialog.loading = false;
  }
};

const showAssociatedPlatforms = async (emailAccount) => {
  currentEmailAccountForDialog.value = emailAccount; // Store context
  associatedInfoDialog.title = `邮箱 "${emailAccount.email_address}" 关联的平台`;
  associatedInfoDialog.pagination.currentPage = 1; // Reset to first page
  emailAccountFormDialog.visible = false; // 确保编辑/新增弹窗已关闭
  await fetchAssociatedPlatformsData(emailAccount.id, 1, associatedInfoDialog.pagination.pageSize);
  associatedInfoDialog.visible = true;
};

const handleAssociatedPageChange = (payload) => {
  if (currentEmailAccountForDialog.value) {
    fetchAssociatedPlatformsData(currentEmailAccountForDialog.value.id, payload.currentPage, payload.pageSize);
  }
};

</script>

<template>
  <div class="email-account-list-view">
    <el-card class="box-card">
      <template #header>
        <div class="card-header">
          <span class="card-title">邮箱账户管理</span>
          <el-button type="primary" :icon="Plus" @click="handleAddEmailAccount">
             添加邮箱账户
          </el-button>
        </div>
      </template>

      <!-- Filters -->
      <div class="filters-section">
        <el-row :gutter="10" class="filter-row">
          <el-col :xs="24" :sm="12" :md="8" :lg="6">
            <el-input
              v-model="emailAccountStore.filters.emailAddressSearch"
              placeholder="按邮箱地址搜索"
              clearable
              @keyup.enter="handleEmailAddressSearchChange(emailAccountStore.filters.emailAddressSearch)"
              @clear="handleEmailAddressSearchChange('')"
            />
          </el-col>
          <el-col :xs="24" :sm="12" :md="8" :lg="6">
            <el-select
              v-model="emailAccountStore.filters.provider"
              placeholder="按服务商筛选"
              clearable
              filterable
              @change="handleProviderFilterChange"
            >
              <el-option
                v-for="item in emailAccountStore.uniqueProviders"
                :key="item"
                :label="item"
                :value="item"
              />
            </el-select>
          </el-col>
          <el-col :xs="24" :sm="12" :md="8" :lg="6">
            <el-button type="primary" @click="triggerApplyAllFilters" :loading="isQuerying">查询</el-button>
            <el-button @click="triggerClearAllFilters" :loading="isResetting">重置所有</el-button>
          </el-col>
        </el-row>
      </div>

      <div class="table-container" style="flex-grow: 1; overflow-y: auto;">
        <el-table
          :data="emailAccountStore.emailAccounts"
          v-loading="emailAccountStore.loading"
          style="width: 100%;"
          height="100%"
          @sort-change="handleSortChange"
          :default-sort="{ prop: emailAccountStore.sort.orderBy, order: emailAccountStore.sort.sortDirection === 'desc' ? 'descending' : 'ascending' }"
          border
          stripe
          resizable
        >
        <el-table-column prop="email_address" label="邮箱地址" min-width="200" sortable="custom" />
        <!-- 服务商列已移除，服务商信息由后端自动提取和管理 -->
        <!-- <el-table-column prop="id" label="ID" width="80" /> -->
        <el-table-column label="关联平台" width="120" :sortable="false">
          <template #default="scope">
            <span>{{ scope.row.platform_count }}</span>
            <el-button
              v-if="scope.row.platform_count > 0"
              type="primary"
              link
              size="small"
              :icon="ViewIcon"
              style="margin-left: 5px;"
              @click="showAssociatedPlatforms(scope.row)"
            />
          </template>
        </el-table-column>
        <el-table-column prop="notes" label="备注" min-width="200" sortable="custom" show-overflow-tooltip />
        <el-table-column prop="created_at" label="创建时间" width="200" sortable="custom" />
        <el-table-column prop="updated_at" label="更新时间" width="200" sortable="custom" />
        <el-table-column label="操作" width="140" fixed="right" align="center">
          <template #default="scope">
            <el-button link type="primary" :icon="Edit" @click="handleEdit(scope.row)">
               编辑
            </el-button>
            <el-button link type="danger" :icon="Delete" @click="confirmDeleteEmailAccount(scope.row.id)">
              删除
            </el-button>
          </template>
        </el-table-column>
        </el-table>
      </div>

      <el-pagination
        class="mt-4"
        background
        layout="total, sizes, prev, pager, next, jumper"
        :total="emailAccountStore.pagination.totalItems"
        :page-sizes="[10, 20, 50, 100]"
        :current-page="emailAccountStore.pagination.currentPage"
        :page-size="emailAccountStore.pagination.pageSize"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
      />
    </el-card>

    <AssociatedInfoDialog
      v-model:visible="associatedInfoDialog.visible"
      :title="associatedInfoDialog.title"
      :items="associatedInfoDialog.items"
      :item-layout="associatedInfoDialog.layout"
      :pagination="associatedInfoDialog.pagination"
      :loading="associatedInfoDialog.loading"
      @page-change="handleAssociatedPageChange"
    />

    <!-- Email Account Add/Edit Modal -->
    <ModalDialog
      v-model:visible="emailAccountFormDialog.visible"
      :title="emailAccountFormDialog.title"
      @confirm="() => { emailAccountFormRef?.triggerSubmit(); }"
      @cancel="handleFormCancel"
      :confirm-button-text="emailAccountFormDialog.isEditMode ? '保存更新' : '立即创建'"
    >
      <EmailAccountForm
        ref="emailAccountFormRef"
        :is-edit="emailAccountFormDialog.isEditMode"
        :email-account="emailAccountFormDialog.currentAccount"
        @submit-form="handleFormSubmit"
        @cancel="handleFormCancel"
      />
      <!-- Removed empty #footer template to use ModalDialog's default buttons -->
    </ModalDialog>
    <!-- End Email Account Add/Edit Modal -->

  </div>
</template>

<style scoped>
.email-account-list-view {
  padding: 20px; /* This padding might need to be on the card or a content wrapper inside */
  background-color: #f0f2f5; /* Light grey background for the whole view */
  display: flex;
  flex-direction: column;
  height: 100%; /* Fill parent (.main-content's router-view) */
  box-sizing: border-box; /* Include padding in height calculation */
}

/* .el-card styling is fine, we need to ensure .box-card within the flex context behaves */
.box-card {
  border-radius: 12px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
  display: flex;
  flex-direction: column;
  flex-grow: 1; /* Allow card to grow and fill space */
  overflow: hidden; /* Prevent card itself from scrolling */
  /* The padding from .email-account-list-view might be better here if that div is just a flex container */
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 0;
}

.card-title {
  font-size: 20px;
  font-weight: bold;
  color: #333;
}

/* 统一搜索栏背景样式 */
.filters-section {
  margin-bottom: 20px;
  padding: 15px;
  background-color: #f9f9f9;
  border-radius: 8px;
  border: 1px solid #ebeef5;
}

.filter-row {
  display: flex;
  flex-wrap: wrap;
  gap: 10px; /* Gap between filter items */
  align-items: center;
}

.filter-row .el-col {
  margin-bottom: 10px; /* Add some bottom margin for columns on smaller screens */
}

.filter-row .el-input,
.filter-row .el-select {
  width: 100%; /* Ensure inputs/selects take full width of their column */
}

/* 表格样式 - 移除竖线和修复多余竖线 */
:deep(.el-table) {
  margin-top: 20px;
  border-radius: 8px;
  overflow: hidden;
  border: none;
}

:deep(.el-table::before) {
  height: 0; /* Remove default bottom border */
}

:deep(.el-table .el-table__row:hover > td) {
  background: linear-gradient(135deg, var(--color-primary-50), rgba(59, 130, 246, 0.05));
}

/* 统一表格行高 - 更紧凑样式 */
:deep(.el-table td.el-table__cell) {
  padding: 4px 8px; /* 更紧凑的内边距 */
  border-bottom: 1px solid var(--color-gray-100);
  border-right: none; /* 移除竖线 */
  line-height: 1.4; /* 紧凑的行高 */
  /* 移除全局的 white-space: nowrap */
}

:deep(.el-table th.el-table__cell) {
  padding: 4px 8px; /* 更紧凑的内边距 */
  background: linear-gradient(135deg, var(--color-gray-50), var(--color-gray-100));
  color: var(--color-gray-700);
  font-weight: var(--font-semibold);
  border-bottom: 2px solid var(--color-gray-200);
  border-right: none; /* 移除竖线 */
  line-height: 1.4; /* 紧凑的行高 */
  /* 移除全局的 white-space: nowrap */
}

/* 为邮箱地址列设置不换行 */
:deep(.el-table td.el-table__cell:nth-child(1)) {
  white-space: nowrap;
}

/* 为备注列设置允许换行并显示省略号 */
:deep(.el-table td.el-table__cell:nth-child(4)) {
  white-space: normal;
  word-break: break-word;
}

/* 为时间列单独设置不换行样式，确保时间不会换行 */
:deep(.el-table td.el-table__cell:nth-child(5), 
       .el-table td.el-table__cell:nth-child(6)) {
  white-space: nowrap;
  min-width: 180px; /* 确保时间列有足够宽度 */
}

/* 操作按钮样式优化 */
:deep(.el-table td.el-table__cell .el-button) {
  margin-right: 0; /* 按钮间距 */
  margin-bottom: 0; /* 移除底部边距 */
  padding: 4px 8px; /* 按钮内边距 */
  font-size: 12px; /* 字体大小 */
  height: 28px; /* 按钮高度 */
  line-height: 1.2; /* 行高 */
  display: inline-block; /* 确保按钮水平排列 */
}

:deep(.el-table td.el-table__cell .el-button:last-child) {
  margin-right: 0; /* 最后一个按钮不需要右边距 */
}

/* 专门为操作列优化布局 */
:deep(.el-table td.el-table__cell:last-child) {
  padding: 8px 4px; /* 操作列的内边距 */
  white-space: nowrap; /* 确保按钮不换行 */
}

/* 操作列的按钮容器 */
:deep(.el-table .el-table__cell:last-child .cell) {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 2px; /* 按钮间隙 */
}

/* Pagination styles moved to utilities.css */

/* Responsive adjustments */
@media (max-width: 768px) {
  .card-header {
    flex-direction: column;
    align-items: flex-start;
  }
  .card-header span {
    margin-bottom: 10px;
  }
  .filter-row .el-col {
    flex-basis: 100%; /* Stack columns on very small screens */
  }
}
</style>