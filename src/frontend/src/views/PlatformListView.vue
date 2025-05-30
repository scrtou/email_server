<template>
  <div class="platform-list-view">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>平台管理</span>
          <el-button type="primary" @click="handleAddPlatform">
            <el-icon><Plus /></el-icon> 添加平台
          </el-button>
        </div>
      </template>

      <el-table :data="platformStore.platforms" v-loading="platformStore.loading" style="width: 100%">
        <el-table-column prop="name" label="平台名称" min-width="200" />
        <!-- <el-table-column prop="id" label="ID" width="80" /> -->
        <el-table-column prop="website_url" label="平台网址" min-width="250">
          <template #default="scope">
            <a :href="scope.row.website_url" target="_blank" rel="noopener noreferrer">{{ scope.row.website_url }}</a>
          </template>
        </el-table-column>
        <el-table-column label="关联邮箱" width="120">
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
        <el-table-column prop="notes" label="备注" min-width="200" show-overflow-tooltip />
        <el-table-column prop="created_at" label="创建时间" width="180" />
        <el-table-column prop="updated_at" label="更新时间" width="180" />
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="scope">
            <el-button size="small" @click="handleEdit(scope.row)">
              <el-icon><Edit /></el-icon> 编辑
            </el-button>
            <el-button size="small" type="danger" @click="confirmDeletePlatform(scope.row.id)">
              <el-icon><Delete /></el-icon> 删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-if="platformStore.pagination.totalItems > 0"
        class="mt-4"
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
import { onMounted, ref, reactive } from 'vue'; // Added reactive
import { useRouter } from 'vue-router';
import { usePlatformStore } from '@/stores/platform';
import { ElIcon, ElButton, ElMessageBox } from 'element-plus';
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

const fetchData = (page = platformStore.pagination.currentPage, pageSize = platformStore.pagination.pageSize) => {
  platformStore.fetchPlatforms(page, pageSize);
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