<template>
  <div class="service-subscription-list-view">
    <el-card class="box-card">
      <template #header>
        <div class="card-header">
          <span class="card-title">服务订阅管理</span>
  <el-button type="primary" @click="handleAdd">
            <el-icon><Plus /></el-icon> 添加订阅
          </el-button>
        </div>
      </template>

      <el-form :inline="true" class="filter-form">
        <el-form-item label="平台名称">
          <el-select
            v-model="filters.filterPlatformName"
            placeholder="搜索平台名称"
            clearable
            filterable
            @change="triggerFetchWithCurrentFilters"
            @clear="triggerFetchWithCurrentFilters"
            style="width: 180px;"
          >
            <el-option
              v-for="item in platformNameOptions"
              :key="item"
              :label="item"
              :value="item"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="邮箱">
          <el-select
            v-model="filters.filterEmail"
            placeholder="搜索邮箱"
            clearable
            filterable
            @change="triggerFetchWithCurrentFilters"
            @clear="triggerFetchWithCurrentFilters"
            style="width: 200px;"
          >
            <el-option
              v-for="item in emailOptions"
              :key="item"
              :label="item"
              :value="item"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="用户名">
          <el-select
            v-model="filters.filterUsername"
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
        <el-form-item>
          <el-button type="primary" @click="triggerFetchWithCurrentFilters" :loading="isQuerying">查询</el-button>
          <el-button @click="triggerClearFilters" :loading="isResetting">重置</el-button>
        </el-form-item>
         <el-form-item label="订阅状态">
          <el-select
            v-model="filters.status"
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
            v-model="filters.billing_cycle"
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
            v-model="renewalDateRangeFilter"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            clearable
            value-format="YYYY-MM-DD"
            style="width: 280px;"
          />
        </el-form-item>
      </el-form>

      <div class="table-container">
        <el-table
          :data="serviceSubscriptionStore.serviceSubscriptions"
          v-loading="serviceSubscriptionStore.loading"
          style="width: 100%; height: 100%;"
          @sort-change="handleSortChange"
          :default-sort="{ prop: serviceSubscriptionStore.sort.orderBy, order: serviceSubscriptionStore.sort.sortDirection === 'desc' ? 'descending' : 'ascending' }"
          border
          stripe
          resizable
        >
        <el-table-column prop="service_name" label="服务名称" min-width="180" sortable="custom" />
        <!-- <el-table-column prop="id" label="ID" width="60" /> -->
        <el-table-column prop="platform_name" label="所属平台" min-width="150" sortable="custom" />
        <el-table-column prop="email_address" label="邮箱" min-width="180" sortable="custom">
          <template #default="scope">
            <span>{{ scope.row.email_address }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="login_username" label="用户名" min-width="150" sortable="custom">
          <template #default="scope">
            <span>{{ scope.row.login_username }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100" sortable="custom" />
        <el-table-column prop="cost" label="费用" width="100" sortable="custom" />
        <el-table-column prop="billing_cycle" label="计费周期" width="140" sortable="custom" />
        <el-table-column prop="next_renewal_date" label="下次续费日" width="140" sortable="custom" />
        <el-table-column prop="description" label="订阅信息" min-width="150" :sortable="false" show-overflow-tooltip />
        <el-table-column prop="created_at" label="创建时间" width="200" sortable="custom" />
        <el-table-column label="操作" width="140" fixed="right" :sortable="false">
          <template #default="scope">
            <el-button link type="primary" :icon="Edit" @click="handleEdit(scope.row)">
               编辑
            </el-button>
            <el-popconfirm
              title="确定要删除此订阅吗？"
              confirm-button-text="确定"
              cancel-button-text="取消"
              @confirm="handleDelete(scope.row.id)"
            >
              <template #reference>
                <el-button link type="danger" :icon="Delete" :loading="serviceSubscriptionStore.loading">
                  删除
                </el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
        </el-table>
      </div>

      <div class="pagination-container">
        <el-pagination
          v-if="serviceSubscriptionStore.pagination.totalItems > 0"
          layout="total, sizes, prev, pager, next, jumper"
          :total="serviceSubscriptionStore.pagination.totalItems"
          :current-page="serviceSubscriptionStore.pagination.currentPage"
          :page-size="serviceSubscriptionStore.pagination.pageSize"
          :page-sizes="settingsStore.getPageSizeOptions('serviceSubscriptions')"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </el-card>

    <ModalDialog
      :visible="isModalVisible"
      :title="modalTitle"
      @confirm="() => serviceSubscriptionFormRef?.triggerSubmit()"
      @cancel="closeModal"
      width="60%"
      :confirm-button-text="isEditMode ? '保存更新' : '立即创建'"
      :show-confirm-button="false"
      :show-cancel-button="false"
    >
      <!-- Form content remains the same -->
      <ServiceSubscriptionForm
        ref="serviceSubscriptionFormRef"
        v-show="isModalVisible"
        :id="isEditMode && currentSubscription ? currentSubscription.id : null"
        :initial-data="currentSubscription"
        @submit-form="handleFormSubmit"
        @cancel="handleFormCancel"
      />
      <!-- Custom Footer with loading state -->
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="closeModal">取消</el-button>
          <el-button
            type="primary"
            @click="handleDirectSubmit"
            :loading="serviceSubscriptionStore.loading"
            :disabled="serviceSubscriptionStore.loading"
          >
            {{ isEditMode ? '保存更新' : '立即创建' }}
          </el-button>
        </div>
      </template>
    </ModalDialog>
  </div>
</template>

<script setup>
import { onMounted, ref, computed, onUnmounted, watch } from 'vue'; // Import onUnmounted
// import { useRouter } from 'vue-router'; // Removed useRouter
const MIN_LOADING_TIME = 300; // 最小加载时间，单位毫秒
import { useServiceSubscriptionStore } from '@/stores/serviceSubscription';
import { usePlatformStore } from '@/stores/platform';
import { useEmailAccountStore } from '@/stores/emailAccount';
import { usePlatformRegistrationStore } from '@/stores/platformRegistration';
import { useSettingsStore } from '@/stores/settings';
import { ElIcon, ElMessage } from 'element-plus';
import { Plus, Edit, Delete } from '@element-plus/icons-vue';
import ModalDialog from '@/components/ui/ModalDialog.vue';
import ServiceSubscriptionForm from '@/components/forms/ServiceSubscriptionForm.vue';

// const router = useRouter(); // Removed router
const serviceSubscriptionStore = useServiceSubscriptionStore();
const platformStore = usePlatformStore();
const emailAccountStore = useEmailAccountStore();
const platformRegistrationStore = usePlatformRegistrationStore();
const settingsStore = useSettingsStore();

const isModalVisible = ref(false);
const modalTitle = ref('');
const currentSubscription = ref(null);
const formMode = ref('add'); // 'add' or 'edit' // This determines isEditMode
const serviceSubscriptionFormRef = ref(null); // Ref for the form
const isQuerying = ref(false); // 用于“查询”按钮的 loading 状态
const isResetting = ref(false); // 用于“重置”按钮的 loading 状态

let fetchController = null; // Variable to hold the AbortController

const isEditMode = computed(() => formMode.value === 'edit');

// Computed properties for select options
const platformNameOptions = computed(() => {
  const names = platformStore.platforms.map(p => p.name);
  return [...new Set(names)].sort();
});
const emailOptions = computed(() => {
  const emails = emailAccountStore.emailAccounts.map(e => e.email_address);
  return [...new Set(emails)].sort();
});
const usernameOptions = computed(() => {
  let filteredSubscriptions = serviceSubscriptionStore.serviceSubscriptions;

  // Filter by selected platform name (interpreted as "service" filter)
  if (filters.filterPlatformName) {
    filteredSubscriptions = filteredSubscriptions.filter(
      sub => sub.platform_name === filters.filterPlatformName
    );
  }

  // Filter by selected email account
  if (filters.filterEmail) {
    filteredSubscriptions = filteredSubscriptions.filter(
      sub => sub.email_address === filters.filterEmail
    );
  }

  const usernames = filteredSubscriptions
    .map(sub => sub.login_username) // Extract username from filtered subscriptions
    .filter(username => username && username.trim() !== ''); // Filter out empty or null usernames
  
  return [...new Set(usernames)].sort(); // Get unique, sorted usernames
});

// Reactive reference for filters, mirroring the store's structure for v-model binding
// This allows local changes to be easily managed and then committed to the store if needed,
// or used directly if the store's filters are already reactive and suitable for v-model.
// For simplicity and directness with Pinia, we can bind v-model directly to store state
// if the store state (filters object) is reactive.
// However, the original code binds to serviceSubscriptionStore.filters.filterPlatformName etc.
// Let's make a local reactive 'filters' object that syncs with the store's filters.
// This approach is often preferred to avoid direct mutation of store state from the template
// if not using `storeToRefs` or similar patterns.
// Given the current structure, it's simpler to continue direct binding or use computed setters if needed.
// For this task, we'll assume direct binding to store is fine as per existing pattern,
// but we'll use a local 'filters' ref for clarity if we were to manage it locally before committing.
// Let's stick to the provided pattern of using serviceSubscriptionStore.filters directly in v-model for now,
// as the request is about v-for and computed for options, not changing filter handling.
// The v-model bindings in the template were already `serviceSubscriptionStore.filters.filterPlatformName` etc.
// The change is to use computed properties for the `v-for` source.

// To make v-model work with the store's filters object directly, ensure it's reactive.
// Pinia state is reactive by default.
// Let's alias serviceSubscriptionStore.filters for easier use in template if preferred,
// or continue using the full path. The original template uses the full path.
// The request implies changing v-model to `filters.filterPlatformName`.
// This means we need a local `filters` ref that is two-way bound or reflects store state.

// Let's use a computed ref for filters that reflects the store,
// and ensure changes are propagated back if necessary (e.g. via @change handlers).
// Or, more simply, just use a direct alias if mutations are handled by store actions.
const filters = serviceSubscriptionStore.filters; // This is a direct reference, mutations will affect the store.
 
const renewalDateRangeFilter = computed({
  get: () => {
    const startDate = filters.renewal_date_start || null;
    const endDate = filters.renewal_date_end || null;
    return [startDate, endDate];
  },
  set: (newVal) => {
    let start = '';
    let end = '';

    if (newVal && newVal.length === 2) {
      start = newVal[0] || '';
      end = newVal[1] || '';
    }
    
    if (filters.renewal_date_start !== start || filters.renewal_date_end !== end) {
      filters.renewal_date_start = start;
      filters.renewal_date_end = end;
      serviceSubscriptionStore.pagination.currentPage = 1;
      fetchData(
        1, // Reset to page 1
        serviceSubscriptionStore.pagination.pageSize,
        serviceSubscriptionStore.sort,
        filters
      );
    }
  }
});
 
// Centralized function to fetch data with cancellation
const fetchData = async (page, size, sort = serviceSubscriptionStore.sort, currentFilters = filters, loadingStateRef = isQuerying) => {
  if (loadingStateRef.value) return; // 如果正在加载，则不执行新的操作

  loadingStateRef.value = true;
  const startTime = Date.now();

  if (fetchController) {
    fetchController.abort(); // Abort previous request if exists
  }
  fetchController = new AbortController(); // Create a new controller for the new request
  
  try {
    await serviceSubscriptionStore.fetchServiceSubscriptions(
      page,
      size,
      sort,
      currentFilters, // Use the local/aliased filters
      fetchController.signal // Pass the signal
    );
  } finally {
    const elapsedTime = Date.now() - startTime;
    if (elapsedTime < MIN_LOADING_TIME) {
      setTimeout(() => {
        loadingStateRef.value = false;
      }, MIN_LOADING_TIME - elapsedTime);
    } else {
      loadingStateRef.value = false;
    }
  }
};

// Abort request on component unmount
onUnmounted(() => {
  if (fetchController) {
    fetchController.abort();
  }
});

// Removed handlePlatformRegistrationFilterChange

const handleStatusFilterChange = (value) => {
  // Assuming direct fetch call here for demonstration:
  filters.status = value; // Use local/aliased filters
  serviceSubscriptionStore.pagination.currentPage = 1;
  fetchData(1, serviceSubscriptionStore.pagination.pageSize);
};

const handleBillingCycleFilterChange = (value) => {
   // Assuming direct fetch call here for demonstration:
  filters.billing_cycle = value; // Use local/aliased filters
  serviceSubscriptionStore.pagination.currentPage = 1;
  fetchData(1, serviceSubscriptionStore.pagination.pageSize);
};

const triggerFetchWithCurrentFilters = () => {
  // When triggering fetch, ensure it uses the current state of 'filters'
  fetchData(1, serviceSubscriptionStore.pagination.pageSize, serviceSubscriptionStore.sort, filters, isQuerying);
};

const triggerClearFilters = async () => {
  if (isResetting.value) return;

  // Set resetting state
  isResetting.value = true;
  const startTime = Date.now();

  try {
    // Clear local filters (which are reactive references to store's filters)
    filters.status = '';
    filters.billing_cycle = '';
    filters.renewal_date_start = '';
    filters.renewal_date_end = '';
    filters.filterPlatformName = '';
    filters.filterEmail = '';
    filters.filterUsername = '';
    serviceSubscriptionStore.pagination.currentPage = 1;

    // Call the store's clearFiltersAndFetch action or directly fetch
    // For consistency with PlatformRegistrationListView, we can call the store's action
    // if it exists and handles fetching. If not, we fetch directly.
    // Assuming serviceSubscriptionStore.clearFilters() exists and fetches:
    // await serviceSubscriptionStore.clearFilters();
    // OR, if we need to manage loading state here:
    await serviceSubscriptionStore.fetchServiceSubscriptions(
      1,
      serviceSubscriptionStore.pagination.pageSize,
      serviceSubscriptionStore.sort, // Use default sort or reset sort as needed
      filters // Pass the now-cleared filters
    );
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
  const newSort = { orderBy: prop, sortDirection };
  fetchData(1, serviceSubscriptionStore.pagination.pageSize, newSort, filters); // Use fetchData
};

const handleAdd = () => {
  formMode.value = 'add';
  modalTitle.value = '添加服务订阅';
  currentSubscription.value = null; // Or some default object if your form expects it
  isModalVisible.value = true;
};

const handleEdit = (row) => {
  formMode.value = 'edit';
  modalTitle.value = '编辑服务订阅';
  currentSubscription.value = { ...row }; // Pass a copy to avoid direct mutation
  isModalVisible.value = true;
};

const handleDelete = async (id) => {
  await serviceSubscriptionStore.deleteServiceSubscription(id);
  // Data is re-fetched by the store action on success
};

const handleSizeChange = (newSize) => {
  // 保存服务订阅管理页面专用的分页设置
  settingsStore.setPageSize('serviceSubscriptions', newSize);
  fetchData(1, newSize); // Use fetchData
};

const handleCurrentChange = (newPage) => {
  fetchData(newPage, serviceSubscriptionStore.pagination.pageSize); // Use fetchData
};

const closeModal = () => {
  isModalVisible.value = false;
  currentSubscription.value = null;
};

// 直接提交表单的方法
const handleDirectSubmit = async () => {
  console.log('服务订阅直接提交按钮被点击');

  // 防止重复提交
  if (serviceSubscriptionStore.loading) {
    console.log('表单正在提交中，忽略重复点击');
    return;
  }

  // 检查表单组件是否存在
  if (!serviceSubscriptionFormRef?.value) {
    ElMessage.error('表单组件未加载，请重试');
    return;
  }

  // 直接调用表单的triggerSubmit方法
  try {
    console.log('调用服务订阅表单的triggerSubmit方法');
    await serviceSubscriptionFormRef.value.triggerSubmit();
    console.log('服务订阅表单提交完成');
  } catch (error) {
    console.error('服务订阅表单提交失败:', error);
    ElMessage.error('提交失败，请重试');
  }
};

const handleFormSubmit = async (eventData) => {
  // 防止重复提交
  if (serviceSubscriptionStore.loading) {
    console.log('表单正在提交中，忽略重复点击');
    return;
  }

  // eventData is { payload, id, isEdit }
  const { payload, id, isEdit: formIsEdit } = eventData;
  let success = false;

  console.log('开始提交服务订阅表单:', { payload, id, formIsEdit });

  try {
    if (formIsEdit) { // This should align with isEditMode.value
      if (!id) {
        ElMessage.error('编辑错误：缺少订阅ID');
        return;
      }
      console.log('调用服务订阅更新方法，ID:', id);
      success = await serviceSubscriptionStore.updateServiceSubscription(id, payload);
    } else { // Create mode
      console.log('调用服务订阅创建方法');
      success = await serviceSubscriptionStore.createServiceSubscription(payload);
    }

    console.log('服务订阅表单提交结果:', success);
    if (success) {
      closeModal();
      // ElMessage is handled by store actions
      // Store actions should also handle re-fetching the list.
      // If not, call fetch here:
      // serviceSubscriptionStore.fetchServiceSubscriptions(
      //   serviceSubscriptionStore.pagination.currentPage,
      //   pageSize.value,
      //   serviceSubscriptionStore.sort
      // );
    } else {
      console.error('服务订阅表单提交失败，success为false');
    }
  } catch (error) {
    console.error('服务订阅表单提交异常:', error);
    ElMessage.error('操作失败，请重试');
  }
};

const handleFormCancel = () => {
  closeModal();
};
 
// Call fetch options on mount
onMounted(async () => {
  // 加载设置
  settingsStore.loadSettings();

  // 同步 store 的 pageSize 与 settings store（使用服务订阅管理页面专用设置）
  serviceSubscriptionStore.pagination.pageSize = settingsStore.getPageSize('serviceSubscriptions');

  // Removed platformRegistrationStore.fetchPlatformRegistrations
  fetchData(serviceSubscriptionStore.pagination.currentPage, serviceSubscriptionStore.pagination.pageSize);
  // Fetch all options for select dropdowns
  // Assuming these fetch actions retrieve all necessary items for the dropdowns.
  // Adjust pagination parameters (e.g., page size to a large number) if these stores'
  // fetch actions are paginated and you need all items.
  // For example: platformStore.fetchPlatforms(1, 10000)
  if (platformStore.platforms.length === 0) {
    await platformStore.fetchPlatforms(1, 10000, { orderBy: 'name', sortDirection: 'asc' });
  }
  if (emailAccountStore.emailAccounts.length === 0) {
    await emailAccountStore.fetchEmailAccounts(1, 10000, { orderBy: 'email_address', sortDirection: 'asc' });
  }
  if (platformRegistrationStore.platformRegistrations.length === 0) {
    // Fetching all registrations might be too much if the list is very large.
    // Consider if this is appropriate or if the username filter should be more dynamic
    // or based on a smaller, more relevant subset.
    // For now, following the pattern of fetching a large list for dropdowns.
    await platformRegistrationStore.fetchPlatformRegistrations(1, 10000, { orderBy: 'login_username', sortDirection: 'asc' });
  }

  // Logging for verification (optional, can be removed after testing)
  console.log('FROM COMPONENT (onMounted): platformStore.platforms count:', platformStore.platforms.length);
  console.log('FROM COMPONENT (onMounted): Computed platformNameOptions.value:', platformNameOptions.value.slice(0, 5)); // Log first 5
  console.log('FROM COMPONENT (onMounted): emailAccountStore.emailAccounts count:', emailAccountStore.emailAccounts.length);
  console.log('FROM COMPONENT (onMounted): Computed emailOptions.value:', emailOptions.value.slice(0, 5)); // Log first 5
  console.log('FROM COMPONENT (onMounted): platformRegistrationStore.platformRegistrations count:', platformRegistrationStore.platformRegistrations.length);
  console.log('FROM COMPONENT (onMounted): Computed usernameOptions.value:', usernameOptions.value.slice(0, 5)); // Log first 5
});

// 监听服务订阅管理页面专用的 pageSize 变化，同步到当前 store
watch(() => settingsStore.getPageSize('serviceSubscriptions'), (newPageSize) => {
  if (serviceSubscriptionStore.pagination.pageSize !== newPageSize) {
    serviceSubscriptionStore.pagination.pageSize = newPageSize;
    // 重新获取数据
    fetchData(1, newPageSize, serviceSubscriptionStore.sort, filters);
  }
});
 
</script>
 
<style scoped>
.service-subscription-list-view {
  padding: 20px;
  background-color: #f0f2f5;
  height: 100vh;
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
}

.box-card {
  border-radius: 12px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
  display: flex;
  flex-direction: column;
  height: calc(100vh - 40px);
  overflow: hidden;
}

.box-card :deep(.el-card__body) {
  display: flex;
  flex-direction: column;
  height: 100%;
  padding: 20px;
  overflow: hidden;
}

.table-container {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
  overflow: hidden;
}

.pagination-container {
   flex-shrink: 0;
  padding: 10px 0;
  display: flex;
  justify-content: center;
  align-items: center;
  max-height: 30px;
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
.filter-form {
  margin-bottom: 20px;
  display: flex;
  flex-wrap: wrap;
  gap: 15px;
  padding: 15px;
  background-color: #f9f9f9;
  border-radius: 8px;
  border: 1px solid #ebeef5;
}

.filter-form .el-form-item {
  margin-right: 0;
  margin-bottom: 0;
}

/* 表格样式 - 关键修复 */
:deep(.el-table) {
  height: 100% !important;
  border-radius: 8px;
  border: none;
}

:deep(.el-table .el-table__body-wrapper) {
  height: calc(100% - 40px) !important;
  overflow-y: auto !important;
}

:deep(.el-table .el-table__header-wrapper) {
  flex-shrink: 0;
}

:deep(.el-table::before) {
  height: 0;
}



:deep(.el-table .el-table__row:hover > td) {
  background: linear-gradient(135deg, var(--color-primary-50), rgba(59, 130, 246, 0.05));
}

/* 统一表格行高 - 更紧凑样式 */
:deep(.el-table td.el-table__cell) {
  padding: 4px 8px; /* 增加垂直内边距 */
  border-bottom: 1px solid var(--color-gray-100);
  border-right: none; /* 移除竖线 */
  line-height: 1.4; /* 紧凑的行高 */
  /* 移除全局的 white-space: nowrap */
}

:deep(.el-table th.el-table__cell) {
  padding: 4px 8px; /* 增加垂直内边距 */
  background: linear-gradient(135deg, var(--color-gray-50), var(--color-gray-100));
  color: var(--color-gray-700);
  font-weight: var(--font-semibold);
  border-bottom: 2px solid var(--color-gray-200);
  border-right: none; /* 移除竖线 */
  white-space: nowrap; /* 防止文字换行 */
  line-height: 1.4; /* 紧凑的行高 */
}

/* 操作按钮样式优化 */
/* 用于操作列单元格内的按钮容器 (div.cell) */
/* 按钮容器 */
:deep(.el-table .el-table__cell:last-child .cell) {
  display: flex !important;
  justify-content: center !important;
  align-items: center !important;
  gap: 2px !important;
}

/* 按钮本身 */
:deep(.el-table td.el-table__cell .el-button) {
  margin-right: 0px !important;
  margin-bottom: 0 !important;
  padding: 4px 8px !important;
  font-size: 12px !important;
  height: 28px !important;
  line-height: 1.2 !important;
  display: inline-block !important;
}

/* 最后一个按钮 */
:deep(.el-table td.el-table__cell .el-button:last-child) {
  margin-right: 0 !important;
}

/* 操作列单元格 */
:deep(.el-table td.el-table__cell:last-child) {
  padding: 8px 4px !important;
  white-space: nowrap !important;
}
/* Pagination styles moved to utilities.css */
</style>