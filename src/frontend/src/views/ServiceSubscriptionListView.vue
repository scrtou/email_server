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

      <div class="table-container" style="flex-grow: 1; overflow-y: auto;">
        <el-table
          :data="serviceSubscriptionStore.serviceSubscriptions"
          v-loading="serviceSubscriptionStore.loading"
          style="width: 100%"
          height="100%"
          @sort-change="handleSortChange"
          :default-sort="{ prop: serviceSubscriptionStore.sort.orderBy, order: serviceSubscriptionStore.sort.sortDirection === 'desc' ? 'descending' : 'ascending' }"
          border
          stripe
          resizable
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
        <el-table-column prop="billing_cycle" label="计费周期" width="140" sortable="custom" />
        <el-table-column prop="next_renewal_date" label="下次续费日" width="140" sortable="custom" />
        <el-table-column prop="notes" label="备注" min-width="150" :sortable="false" show-overflow-tooltip />
        <el-table-column prop="created_at" label="创建时间" width="200" sortable="custom" />
        <el-table-column label="操作" width="120" fixed="right" :sortable="false">
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
                <el-button link type="danger" :icon="Delete">
                  删除
                </el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
        </el-table>
      </div>

      <el-pagination
        v-if="serviceSubscriptionStore.pagination.totalItems > 0"
        background
        layout="total, sizes, prev, pager, next, jumper"
        :total="serviceSubscriptionStore.pagination.totalItems"
        :page-sizes="[10, 20, 50, 100]"
        :current-page="serviceSubscriptionStore.pagination.currentPage"
        :page-size="pageSize.value"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
      />
    </el-card>

    <ModalDialog
      :visible="isModalVisible"
      :title="modalTitle"
      @confirm="() => serviceSubscriptionFormRef?.triggerSubmit()"
      @cancel="closeModal"
      width="60%"
      :confirm-button-text="isEditMode ? '保存更新' : '立即创建'"
    >
      <ServiceSubscriptionForm
        ref="serviceSubscriptionFormRef"
        v-if="isModalVisible"
        :id="isEditMode && currentSubscription ? currentSubscription.id : null"
        :initial-data="currentSubscription"
        @submit-form="handleFormSubmit"
        @cancel="handleFormCancel"
      />
    </ModalDialog>
  </div>
</template>

<script setup>
import { onMounted, ref, computed } from 'vue';
// import { useRouter } from 'vue-router'; // Removed useRouter
import { useServiceSubscriptionStore } from '@/stores/serviceSubscription';
import { usePlatformRegistrationStore } from '@/stores/platformRegistration';
import { ElIcon, ElMessage } from 'element-plus';
import { Plus, Edit, Delete } from '@element-plus/icons-vue';
import ModalDialog from '@/components/ui/ModalDialog.vue';
import ServiceSubscriptionForm from '@/components/forms/ServiceSubscriptionForm.vue';

// const router = useRouter(); // Removed router
const serviceSubscriptionStore = useServiceSubscriptionStore();
const platformRegistrationStore = usePlatformRegistrationStore();
const pageSize = ref(serviceSubscriptionStore.pagination.pageSize || 10);

const isModalVisible = ref(false);
const modalTitle = ref('');
const currentSubscription = ref(null);
const formMode = ref('add'); // 'add' or 'edit' // This determines isEditMode
const serviceSubscriptionFormRef = ref(null); // Ref for the form

const isEditMode = computed(() => formMode.value === 'edit');

onMounted(async () => {
  if (platformRegistrationStore.platformRegistrations.length === 0) {
    await platformRegistrationStore.fetchPlatformRegistrations({ page: 1, pageSize: 1000, orderBy: 'platform_name', sortDirection: 'asc' });
  }
  serviceSubscriptionStore.fetchServiceSubscriptions(
    serviceSubscriptionStore.pagination.currentPage,
    pageSize.value,
    serviceSubscriptionStore.sort
  );
});

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
  serviceSubscriptionStore.filters.renewal_date_start = start;
  serviceSubscriptionStore.filters.renewal_date_end = end;
  serviceSubscriptionStore.pagination.currentPage = 1;
  serviceSubscriptionStore.fetchServiceSubscriptions(
    serviceSubscriptionStore.pagination.currentPage,
    pageSize.value,
    serviceSubscriptionStore.sort
  );
};

const triggerFetchWithCurrentFilters = () => {
  serviceSubscriptionStore.fetchServiceSubscriptions(1, pageSize.value);
};

const triggerClearFilters = () => {
  serviceSubscriptionStore.clearFilters();
};

const handleSortChange = ({ prop, order }) => {
  const sortDirection = order === 'descending' ? 'desc' : 'asc';
  serviceSubscriptionStore.fetchServiceSubscriptions(
    1,
    pageSize.value,
    { orderBy: prop, sortDirection }
  );
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
  pageSize.value = newSize;
  serviceSubscriptionStore.fetchServiceSubscriptions(1, pageSize.value);
};

const handleCurrentChange = (newPage) => {
  serviceSubscriptionStore.fetchServiceSubscriptions(newPage, pageSize.value);
};

const closeModal = () => {
  isModalVisible.value = false;
  currentSubscription.value = null;
};

const handleFormSubmit = async (eventData) => {
  // eventData is { payload, id, isEdit }
  const { payload, id, isEdit: formIsEdit } = eventData;
  let success = false;

  if (formIsEdit) { // This should align with isEditMode.value
    if (!id) {
      ElMessage.error('编辑错误：缺少订阅ID');
      return;
    }
    success = await serviceSubscriptionStore.updateServiceSubscription(id, payload);
  } else { // Create mode
    success = await serviceSubscriptionStore.createServiceSubscription(payload);
  }

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
  }
};

const handleFormCancel = () => {
  closeModal();
};

</script>

<style scoped>
.service-subscription-list-view {
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