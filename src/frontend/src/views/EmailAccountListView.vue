<template>
  <div class="email-account-list-view">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>邮箱账户管理</span>
          <el-button type="primary" @click="handleAddEmailAccount">
            <el-icon><Plus /></el-icon> 添加邮箱账户
          </el-button>
        </div>
      </template>

      <!-- Filters -->
      <div class="filters-section" style="margin-bottom: 20px; display: flex; flex-wrap: wrap; gap: 10px; align-items: center;">
        <el-input
          v-model="emailAccountStore.filters.emailAddressSearch"
          placeholder="按邮箱地址搜索"
          clearable
          @keyup.enter="handleEmailAddressSearchChange(emailAccountStore.filters.emailAddressSearch)"
          @clear="handleEmailAddressSearchChange('')"
          style="width: 240px;"
        />
        <el-select
          v-model="emailAccountStore.filters.provider"
          placeholder="按服务商筛选"
          clearable
          filterable
          @change="handleProviderFilterChange"
          style="width: 240px;"
        >
          <el-option
            v-for="item in emailAccountStore.uniqueProviders"
            :key="item"
            :label="item"
            :value="item"
          />
        </el-select>
        <el-button type="primary" @click="triggerApplyAllFilters">查询</el-button>
        <el-button @click="triggerClearAllFilters">重置所有</el-button>
      </div>

      <el-table
        :data="emailAccountStore.emailAccounts"
        v-loading="emailAccountStore.loading"
        style="width: 100%"
        @sort-change="handleSortChange"
        :default-sort="{ prop: emailAccountStore.sort.orderBy, order: emailAccountStore.sort.sortDirection === 'desc' ? 'descending' : 'ascending' }"
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
        <el-table-column prop="created_at" label="创建时间" width="180" sortable="custom" />
        <el-table-column prop="updated_at" label="更新时间" width="180" sortable="custom" />
        <el-table-column label="操作" width="180" fixed="right" :sortable="false">
          <template #default="scope">
            <el-button size="small" @click="handleEdit(scope.row)">
              <el-icon><Edit /></el-icon> 编辑
            </el-button>
            <el-button size="small" type="danger" @click="confirmDeleteEmailAccount(scope.row.id)">
              <el-icon><Delete /></el-icon> 删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-if="emailAccountStore.pagination.totalItems > 0"
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
  </div>
</template>

<script setup>
import { onMounted, ref, reactive } from 'vue';
import { useRouter } from 'vue-router';
import { useEmailAccountStore } from '@/stores/emailAccount';
import { ElIcon, ElButton, ElMessageBox } from 'element-plus';
import { Plus, Edit, Delete, View as ViewIcon } from '@element-plus/icons-vue';
import AssociatedInfoDialog from '@/components/AssociatedInfoDialog.vue';

const router = useRouter();
const emailAccountStore = useEmailAccountStore();

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

const triggerApplyAllFilters = () => {
  // This is for a dedicated "Query/Search" button if present
  // It ensures that the current state of all filters in the store is used to fetch.
  // The individual filter handlers already call setFilter which then calls fetch.
  // So this button might be redundant if instant filtering on change is preferred.
  // If used, it should fetch with all current filters.
  emailAccountStore.fetchEmailAccounts(1, emailAccountStore.pagination.pageSize, emailAccountStore.sort, emailAccountStore.filters);
};

const triggerClearAllFilters = () => {
  emailAccountStore.clearFilters(); // This will clear all filters and fetch
  // After clearing, explicitly call fetch to ensure UI updates if clearFilters doesn't auto-fetch
  // (though the current store's clearFilters doesn't auto-fetch, setFilter does)
  emailAccountStore.fetchEmailAccounts(1, emailAccountStore.pagination.pageSize, emailAccountStore.sort, emailAccountStore.filters);
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
  // Navigate to a form view for creating a new email account
  // This route will be defined later
  router.push({ name: 'EmailAccountCreate' });
};

const handleEdit = (row) => {
  // Navigate to a form view for editing, passing the id
  // This route will be defined later
  router.push({ name: 'EmailAccountEdit', params: { id: row.id } });
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
  emailAccountStore.fetchEmailAccounts(1, newSize); // Reset to page 1 when size changes
};

const handleCurrentChange = (newPage) => {
  // Store's fetchEmailAccounts will use current filters and sort
  emailAccountStore.fetchEmailAccounts(newPage, emailAccountStore.pagination.pageSize);
};

const fetchAssociatedPlatformsData = async (emailAccountId, page = 1, pageSize = 5) => {
  associatedInfoDialog.loading = true;
  try {
    const result = await emailAccountStore.fetchAssociatedPlatformRegistrations(emailAccountId, { page, pageSize });
    associatedInfoDialog.items = result.data;
    associatedInfoDialog.pagination.currentPage = result.meta.current_page;
    associatedInfoDialog.pagination.pageSize = result.meta.page_size;
    associatedInfoDialog.pagination.totalItems = result.meta.total_records;
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
  await fetchAssociatedPlatformsData(emailAccount.id, 1, associatedInfoDialog.pagination.pageSize);
  associatedInfoDialog.visible = true;
};

const handleAssociatedPageChange = (payload) => {
  if (currentEmailAccountForDialog.value) {
    fetchAssociatedPlatformsData(currentEmailAccountForDialog.value.id, payload.currentPage, payload.pageSize);
  }
};

</script>

<style scoped>
.email-account-list-view {
  padding: 20px;
}
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.mt-4 {
  margin-top: 1.5rem;
}
</style>