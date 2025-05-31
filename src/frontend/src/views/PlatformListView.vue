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
        <el-button type="primary" @click="triggerApplyAllFilters">查询</el-button>
        <el-button @click="triggerClearAllFilters">重置</el-button>
      </div>

      <el-table
        :data="platformStore.platforms"
        v-loading="platformStore.loading"
        style="width: 100%"
        @sort-change="handleSortChange"
        :default-sort="{ prop: platformStore.sort.orderBy, order: platformStore.sort.sortDirection === 'desc' ? 'descending' : 'ascending' }"
        border
        stripe
      >
        <el-table-column prop="name" label="平台名称" min-width="180" sortable="custom" show-overflow-tooltip />
        <el-table-column prop="website_url" label="平台网址" min-width="220" sortable="custom" show-overflow-tooltip>
          <template #default="scope">
            <el-link :href="scope.row.website_url" target="_blank" type="primary">{{ scope.row.website_url }}</el-link>
          </template>
        </el-table-column>
        <el-table-column label="关联邮箱" width="120" align="center">
          <template #default="scope">
            <el-tag v-if="scope.row.email_account_count > 0" type="info" size="small">
              {{ scope.row.email_account_count }}
            </el-tag>
            <span v-else>0</span>
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
        <el-table-column prop="created_at" label="创建时间" width="160" sortable="custom" />
        <el-table-column prop="updated_at" label="更新时间" width="160" sortable="custom" />
        <el-table-column label="操作" width="160" fixed="right" align="center">
          <template #default="scope">
            <el-button size="small" :icon="Edit" @click="handleEdit(scope.row)">
              编辑
            </el-button>
            <el-button size="small" type="danger" :icon="Delete" @click="confirmDeletePlatform(scope.row.id)">
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-if="platformStore.pagination.totalItems > 0"
        class="pagination-container"
        background
        layout="total, sizes, prev, pager, next, jumper"
        :total="platformStore.pagination.totalItems"
        :page-sizes="[10, 20, 50, 100]"
        :current-page="platformStore.pagination.currentPage"
        :page-size="platformStore.pagination.pageSize"
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
import { usePlatformStore } from '@/stores/platform';
import { ElCard, ElButton, ElInput, ElTable, ElTableColumn, ElPagination, ElMessageBox, ElLink, ElTag } from 'element-plus';
import { Plus, Edit, Delete, View as ViewIcon } from '@element-plus/icons-vue';
import AssociatedInfoDialog from '@/components/AssociatedInfoDialog.vue';

const router = useRouter();
const platformStore = usePlatformStore();

const currentPlatformForDialog = ref(null); // To store the platform context for pagination

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

const fetchData = (
  page = platformStore.pagination.currentPage,
  pageSize = platformStore.pagination.pageSize,
  sortOptions = { orderBy: platformStore.sort.orderBy, sortDirection: platformStore.sort.sortDirection },
  filterOptions = { nameSearch: platformStore.filters.nameSearch } // Pass current nameSearch filter
) => {
  platformStore.fetchPlatforms(page, pageSize, sortOptions, filterOptions);
};

const handleNameSearchChange = (value) => {
  platformStore.setFilter('nameSearch', value || '');
  fetchData(1); // Reset to page 1 and fetch with new filter
};

const triggerApplyAllFilters = () => {
  fetchData(1); // Fetch with all current filters from store, reset to page 1
};

const triggerClearAllFilters = () => {
  platformStore.clearFilters(); // Clears nameSearch in store
  fetchData(1); // Fetch with cleared filters, reset to page 1
};

const handleSortChange = ({ prop, order }) => {
  const sortDirection = order === 'descending' ? 'desc' : 'asc';
  fetchData(1, platformStore.pagination.pageSize, { orderBy: prop, sortDirection });
};

const handleAddPlatform = () => {
  router.push({ name: 'PlatformCreate' }); 
};

const handleEdit = (row) => {
  router.push({ name: 'PlatformEdit', params: { id: row.id } });
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
  fetchData(1, newSize); 
};

const handleCurrentChange = (newPage) => {
  fetchData(newPage, platformStore.pagination.pageSize);
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
  min-height: calc(100vh - 80px); /* Adjust based on your AppLayout header/footer */
}

.box-card {
  border-radius: 12px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
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

/* Table specific styles */
.el-table {
  margin-top: 20px; /* Space above the table */
  border-radius: 8px;
  overflow: hidden; /* Ensures border-radius applies to table corners */
}

/* Pagination styles */
.pagination-container {
  margin-top: 25px; /* More space above pagination */
  display: flex;
  justify-content: flex-end; /* Align pagination to the right */
  padding: 10px 0;
}

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