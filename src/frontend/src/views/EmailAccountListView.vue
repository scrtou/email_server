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
  fetchData();
});

const fetchData = (
  page = emailAccountStore.pagination.currentPage,
  pageSize = emailAccountStore.pagination.pageSize,
  sortOptions = { orderBy: emailAccountStore.sort.orderBy, sortDirection: emailAccountStore.sort.sortDirection }
) => {
  emailAccountStore.fetchEmailAccounts(page, pageSize, sortOptions);
};

const handleSortChange = ({ prop, order }) => {
  const sortDirection = order === 'descending' ? 'desc' : 'asc';
  fetchData(1, emailAccountStore.pagination.pageSize, { orderBy: prop, sortDirection });
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
        // Data is re-fetched by the store action on success
        // Check if the current page becomes empty after deletion
        if (emailAccountStore.emailAccounts.length === 0 && emailAccountStore.pagination.currentPage > 1) {
          fetchData(emailAccountStore.pagination.currentPage - 1);
        }
      }
    })
    .catch(() => {
      // User cancelled
    });
};

const handleSizeChange = (newSize) => {
  fetchData(1, newSize); // Reset to page 1 when size changes
};

const handleCurrentChange = (newPage) => {
  fetchData(newPage, emailAccountStore.pagination.pageSize);
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