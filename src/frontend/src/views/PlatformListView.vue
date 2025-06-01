<template>
  <div class="platform-list-view">
    <el-card class="box-card">
      <template #header>
        <div class="card-header">
          <span class="card-title">平台管理</span>
          <el-button type="primary" :icon="Plus" @click="handleAddPlatform">
            添加平台
          </el-button>
        </div>
      </template>

      <!-- Filters -->
      <div class="filters-section">
        <el-input
          v-model="platformStore.filters.nameSearch"
          placeholder="按平台名称搜索"
          clearable
          @keyup.enter="handleNameSearchChange(platformStore.filters.nameSearch)"
          @clear="handleNameSearchChange('')"
          class="filter-input"
        />
        <el-button type="primary" @click="triggerApplyAllFilters" :loading="isQuerying">查询</el-button>
        <el-button @click="triggerClearAllFilters" :loading="isResetting">重置</el-button>
      </div>

      <div class="table-container" style="flex-grow: 1; overflow-y: auto;">
        <el-table
          :data="platformStore.platforms"
          v-loading="platformStore.loading"
          style="width: 100%"
          height="100%"
          @sort-change="handleSortChange"
          :default-sort="{ prop: platformStore.sort.orderBy, order: platformStore.sort.sortDirection === 'desc' ? 'descending' : 'ascending' }"
          border
          stripe
          resizable
        >
        <el-table-column prop="name" label="平台名称" min-width="180" sortable="custom" show-overflow-tooltip />
        <el-table-column prop="website_url" label="平台网址" min-width="220" sortable="custom" show-overflow-tooltip>
          <template #default="scope">
            <el-link :href="scope.row.website_url" target="_blank" type="primary">{{ scope.row.website_url }}</el-link>
          </template>
        </el-table-column>
        <el-table-column label="关联邮箱" width="120" align="center">
          <template #default="scope">
            <span>{{ scope.row.email_account_count }}</span>  
            <el-button
              v-if="scope.row.email_account_count > 0"
              type="primary"
              link
              size="small"
              :icon="ViewIcon"
              style="margin-left: 5px;"
              @click="showAssociatedEmails(scope.row)"
            />
          </template>
        </el-table-column>
        <el-table-column prop="notes" label="备注" min-width="180" sortable="custom" show-overflow-tooltip />
        <el-table-column prop="created_at" label="创建时间" width="200" sortable="custom" />
        <el-table-column prop="updated_at" label="更新时间" width="200" sortable="custom" />
        <el-table-column label="操作" width="140" fixed="right" align="center">
          <template #default="scope">
            <el-button link type="primary" :icon="Edit" @click="handleEdit(scope.row)">
              编辑
            </el-button>
            <el-button link type="danger"  :icon="Delete" @click="confirmDeletePlatform(scope.row.id)">
              删除
            </el-button>
          </template>
        </el-table-column>
        </el-table>
      </div>

      <el-pagination
        v-if="platformStore.pagination.totalItems > 0"
        class="pagination-container"
        background
        layout="total, sizes, prev, pager, next, jumper"
        :total="platformStore.pagination.totalItems"
        :page-sizes="[10, 20, 50, 100]"
        :current-page="platformStore.pagination.currentPage"
        :page-size="pageSize.value"
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

    <!-- 新增/编辑平台弹框 -->
    <ModalDialog
      :visible="modalVisible"
      :title="modalTitle"
      @update:visible="modalVisible = $event"
      @confirm="() => platformFormRef?.triggerSubmit()"
      @cancel="handleFormCancel"
      :confirm-button-text="modalMode === 'edit' ? '保存更新' : '立即创建'"
    >
      <PlatformForm
        ref="platformFormRef"
        v-if="modalVisible"
        :id="modalMode === 'edit' && currentEditItem ? currentEditItem.id : null"
        :initial-data="modalMode === 'edit' && currentEditItem ? currentEditItem : {}"
        @submit-form="handleFormSubmit"
        @cancel="handleFormCancel"
      />
    </ModalDialog>
  </div>
</template>

<script setup>
import { onMounted, ref, reactive } from 'vue';
// import { useRouter } from 'vue-router'; // 移除 useRouter
import { usePlatformStore } from '@/stores/platform';
import { ElMessage, ElCard, ElButton, ElInput, ElTable, ElTableColumn, ElPagination, ElMessageBox, ElLink} from 'element-plus';
import { Plus, Edit, Delete, View as ViewIcon } from '@element-plus/icons-vue';
import AssociatedInfoDialog from '@/components/AssociatedInfoDialog.vue';
import ModalDialog from '@/components/ui/ModalDialog.vue'; // 引入 ModalDialog
import PlatformForm from '@/components/forms/PlatformForm.vue'; // 引入 PlatformForm

// const router = useRouter(); // 移除 router
const platformStore = usePlatformStore();
const pageSize = ref(platformStore.pagination.pageSize || 10); // Initialize with store's pageSize or a default
const MIN_LOADING_TIME = 300; // Minimum loading time in milliseconds
const isQuerying = ref(false);
const isResetting = ref(false);

const currentPlatformForDialog = ref(null); // To store the platform context for pagination

// 弹框相关状态
const platformFormRef = ref(null); // Ref for PlatformForm
const modalVisible = ref(false);
const modalTitle = ref('');
const modalMode = ref('add'); // 'add' or 'edit'
const currentEditItem = ref(null);

const associatedInfoDialog = reactive({
  visible: false,
  title: '',
  items: [],
  layout: [
    { label: '邮箱地址', prop: 'email_address', minWidth: '200px' },
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
  fetchData();
});

const fetchData = async (
  page = platformStore.pagination.currentPage,
  size = pageSize.value, // Use the new ref here
  sortOptions = { orderBy: platformStore.sort.orderBy, sortDirection: platformStore.sort.sortDirection },
  filterOptions = { nameSearch: platformStore.filters.nameSearch } // Pass current nameSearch filter
) => {
  // platformStore.loading is handled by the store itself.
  // The individual button loading states (isQuerying, isResetting) are separate.
  await platformStore.fetchPlatforms(page, size, sortOptions, filterOptions);
};

const handleNameSearchChange = (value) => {
  platformStore.setFilter('nameSearch', value || '');
  fetchData(1); // Reset to page 1 and fetch with new filter
};

const triggerApplyAllFilters = async () => {
  isQuerying.value = true;
  const startTime = Date.now();
  try {
    // fetchData will use platformStore.filters.nameSearch internally
    await fetchData(1); // Fetch with all current filters from store, reset to page 1
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
    platformStore.clearFilters(); // Clears nameSearch in store
    await fetchData(1); // Fetch with cleared filters, reset to page 1
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
  fetchData(1, pageSize.value, { orderBy: prop, sortDirection });
};

const handleAddPlatform = () => {
  modalMode.value = 'add';
  modalTitle.value = '添加平台';
  associatedInfoDialog.visible = false; // 确保关联信息弹窗已关闭
  currentEditItem.value = null;
  modalVisible.value = true;
};

const handleEdit = (row) => {
  associatedInfoDialog.visible = false; // 确保关联信息弹窗已关闭
  modalMode.value = 'edit';
  modalTitle.value = '编辑平台';
  currentEditItem.value = { ...row };
  modalVisible.value = true;
};

const handleFormSubmit = async (payloadFromForm) => {
  // payloadFromForm is the object emitted by PlatformForm's submit-form event
  // The PlatformForm itself handles the store interaction (create/update)
  // This handler is now primarily for closing the dialog and refreshing the list.
  
  // The actual API call is now inside PlatformForm, triggered by its internal handleSubmit,
  // which then emits 'submit-form' with the payload.
  // The parent (this view) will then call the store.

  let success = false;
  if (modalMode.value === 'edit' && currentEditItem.value && currentEditItem.value.id) {
    success = await platformStore.updatePlatform(currentEditItem.value.id, payloadFromForm);
  } else if (modalMode.value === 'add') {
    success = await platformStore.createPlatform(payloadFromForm);
  } else {
    ElMessage.error('操作失败：无法确定模式或ID丢失。');
    return;
  }

  if (success) {
    modalVisible.value = false;
    fetchData(); // Refresh list data
  }
  // Error messages are handled by the store actions
};

const handleFormCancel = () => {
  modalVisible.value = false;
  // Optionally call reset on the form if needed, though v-if should re-initialize
  // platformFormRef.value?.resetForm();
};

const confirmDeletePlatform = (id) => {
  ElMessageBox.confirm(
    '删除此平台将同时删除其下所有关联的平台注册信息以及这些平台注册信息下的服务订阅数据。是否确认删除？',
    '警告',
    {
      confirmButtonText: '确定删除',
      cancelButtonText: '取消',
      type: 'warning',
    }
  )
    .then(async () => {
      const success = await platformStore.deletePlatform(id);
      if (success) {
        if (platformStore.platforms.length === 0 && platformStore.pagination.currentPage > 1) {
          fetchData(platformStore.pagination.currentPage - 1);
        }
      }
    })
    .catch(() => {
      // User cancelled
    });
};

const handleSizeChange = (newSize) => {
  pageSize.value = newSize;
  fetchData(1, pageSize.value);
};

const handleCurrentChange = (newPage) => {
  fetchData(newPage, pageSize.value);
};

const fetchAssociatedEmailsData = async (platformId, page = 1, pageSize = 5) => {
  associatedInfoDialog.loading = true;
  try {
    const result = await platformStore.fetchAssociatedEmailRegistrations(platformId, { page, pageSize });
    associatedInfoDialog.items = result.data;
    associatedInfoDialog.pagination.currentPage = result.meta.current_page;
    associatedInfoDialog.pagination.pageSize = result.meta.page_size;
    associatedInfoDialog.pagination.totalItems = result.meta.total_records;
  } catch (error) {
    associatedInfoDialog.items = [];
    associatedInfoDialog.pagination.totalItems = 0;
  } finally {
    associatedInfoDialog.loading = false;
  }
};

const showAssociatedEmails = async (platform) => {
  currentPlatformForDialog.value = platform; // Store context
  associatedInfoDialog.title = `平台 "${platform.name}" 关联的邮箱账户`;
  associatedInfoDialog.pagination.currentPage = 1; // Reset to first page
  modalVisible.value = false; // 确保编辑/新增弹窗已关闭
  await fetchAssociatedEmailsData(platform.id, 1, associatedInfoDialog.pagination.pageSize);
  associatedInfoDialog.visible = true;
};

const handleAssociatedPageChange = (payload) => {
  if (currentPlatformForDialog.value) {
    fetchAssociatedEmailsData(currentPlatformForDialog.value.id, payload.currentPage, payload.pageSize);
  }
};
</script>

<style scoped>
.platform-list-view {
  padding: 20px;
  background-color: #f0f2f5; /* Light grey background for the whole view */
  display: flex;
  flex-direction: column;
  height: 100%;
  box-sizing: border-box;
}

.box-card {
  border-radius: 12px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
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
  font-size: 20px;
  font-weight: bold;
  color: #333;
}

.filters-section {
  display: flex;
  flex-wrap: wrap; /* Allow wrapping on smaller screens */
  gap: 15px; /* Increased gap for better spacing */
  align-items: center;
  margin-bottom: 20px;
  padding: 15px;
  background-color: #f9f9f9;
  border-radius: 8px;
  border: 1px solid #ebeef5;
}

.filter-input {
  width: 220px; /* Slightly wider input */
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

/* Pagination styles moved to utilities.css */
/* .pagination-container class is still used in the template, but styled globally now */

/* Responsive adjustments */
@media (max-width: 768px) {
  .filters-section {
    flex-direction: column;
    align-items: stretch;
  }

  .filter-input {
    width: 100%; /* Full width on small screens */
  }

  .el-button {
    width: 100%; /* Full width buttons */
  }
}
</style>