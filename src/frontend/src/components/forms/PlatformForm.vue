<template>
  <el-card class="platform-form-card">
    <template #header>
      <div class="card-header">
        <span class="card-title">{{ isEditMode ? '编辑平台' : '添加平台' }}</span>
      </div>
    </template>
    <el-form
      ref="platformFormRef"
      :model="form"
      :rules="rules"
      label-width="120px"
      v-loading="loading"
      class="platform-form"
    >
      <el-form-item label="平台名称" prop="name">
        <el-input v-model="form.name" placeholder="例如：Google, GitHub, Steam" clearable />
      </el-form-item>
      <el-form-item label="平台网址" prop="website_url">
        <el-input v-model="form.website_url" placeholder="例如：https://www.google.com" clearable />
      </el-form-item>
      <el-form-item label="备注" prop="notes">
        <el-input type="textarea" v-model="form.notes" :rows="4" resize="vertical" show-word-limit maxlength="500" />
      </el-form-item>
      <el-form-item class="form-actions">
        <el-button type="primary" @click="handleSubmit" :loading="loading">
          {{ isEditMode ? '保存更新' : '立即创建' }}
        </el-button>
        <el-button @click="handleCancel">取消</el-button>
      </el-form-item>
    </el-form>

    <el-divider v-if="isEditMode && form.name" content-position="left">
      <span class="divider-text">关联的邮箱账户注册信息</span>
    </el-divider>
    <div v-if="isEditMode && form.name" class="associated-info-section">
      <el-button
        type="info"
        plain
        :icon="ViewIcon"
        @click="showAssociatedEmailsDialog"
        :disabled="associatedInfoDialog.loading"
        class="view-associated-button"
      >
        查看在此平台上注册的邮箱 ({{ form.email_account_count || 0 }})
      </el-button>
    </div>
  </el-card>

  <AssociatedInfoDialog
    v-if="isEditMode"
    v-model:visible="associatedInfoDialog.visible"
    :title="associatedInfoDialog.title"
    :items="associatedInfoDialog.items"
    :item-layout="associatedInfoDialog.layout"
    :pagination="associatedInfoDialog.pagination"
    :loading="associatedInfoDialog.loading"
    @page-change="handleAssociatedPageChange"
  />
</template>

<script setup>
import { ref, onMounted, computed, reactive } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { usePlatformStore } from '@/stores/platform';
import { ElMessage, ElDivider, ElButton, ElCard, ElForm, ElFormItem, ElInput } from 'element-plus';
import { View as ViewIcon } from '@element-plus/icons-vue';
import AssociatedInfoDialog from '@/components/AssociatedInfoDialog.vue';

// eslint-disable-next-line no-undef
const props = defineProps({
  id: {
    type: [String, Number],
    default: null,
  },
});

const router = useRouter();
const route = useRoute();
const platformStore = usePlatformStore();

const platformFormRef = ref(null);
const form = ref({
  name: '',
  website_url: '',
  notes: '',
  email_account_count: 0, // To store the count from fetched platform data
});
const loading = ref(false);

const associatedInfoDialog = reactive({
  visible: false,
  title: '',
  items: [],
  layout: [ // Define layout for displaying email registrations
    { label: '邮箱地址', prop: 'email_address', minWidth: '200px' },
    { label: '登录用户名', prop: 'login_username', minWidth: '150px' },
    { label: '注册备注', prop: 'registration_notes', minWidth: '200px', showOverflowTooltip: true },
  ],
  pagination: {
    currentPage: 1,
    pageSize: 5,
    totalItems: 0,
  },
  loading: false,
});

const isEditMode = computed(() => !!props.id || !!route.params.id);
const currentId = computed(() => props.id || route.params.id);

const rules = ref({
  name: [
    { required: true, message: '请输入平台名称', trigger: 'blur' },
    { min: 2, max: 100, message: '名称长度应为 2 到 100 个字符', trigger: 'blur' },
  ],
  website_url: [
    { type: 'url', message: '请输入有效的网址', trigger: ['blur', 'change'] },
  ],
});

onMounted(async () => {
  if (isEditMode.value && currentId.value) {
    loading.value = true;
    const platformData = await platformStore.fetchPlatformById(currentId.value);
    if (platformData) {
      form.value.name = platformData.name;
      form.value.website_url = platformData.website_url;
      form.value.notes = platformData.notes;
      form.value.email_account_count = platformData.email_account_count || 0;
      platformStore.setCurrentPlatform(platformData);
    } else {
      ElMessage.error('无法加载平台数据，可能ID无效');
      router.push({ name: 'PlatformList' });
    }
    loading.value = false;
  }
});

const handleSubmit = async () => {
  if (!platformFormRef.value) return;
  await platformFormRef.value.validate(async (valid) => {
    if (valid) {
      loading.value = true;
      const payload = {
        name: form.value.name,
        notes: form.value.notes,
      };
      if (form.value.website_url && form.value.website_url.trim() !== '') {
        payload.website_url = form.value.website_url;
      }

      let success = false;
      if (isEditMode.value) {
        success = await platformStore.updatePlatform(currentId.value, payload);
      } else {
        success = await platformStore.createPlatform(payload);
      }
      loading.value = false;
      if (success) {
        router.push({ name: 'PlatformList' });
      }
    } else {
      ElMessage.error('请检查表单输入');
      return false;
    }
  });
};

const handleCancel = () => {
  router.push({ name: 'PlatformList' });
};

const fetchAssociatedEmailsData = async (page = 1, pageSize = 5) => {
  if (!currentId.value) return;
  associatedInfoDialog.loading = true;
  try {
    // Assuming platformStore has a method like fetchAssociatedEmailRegistrations
    const result = await platformStore.fetchAssociatedEmailRegistrations(currentId.value, { page, pageSize });
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

const showAssociatedEmailsDialog = async () => {
  if (!form.value.name || !currentId.value) return;
  associatedInfoDialog.title = `平台 "${form.value.name}" 关联的邮箱注册信息`;
  associatedInfoDialog.pagination.currentPage = 1;
  await fetchAssociatedEmailsData(1, associatedInfoDialog.pagination.pageSize);
  associatedInfoDialog.visible = true;
};

const handleAssociatedPageChange = (payload) => {
  fetchAssociatedEmailsData(payload.currentPage, payload.pageSize);
};

</script>

<style scoped>
.platform-form-card {
  max-width: 700px;
  margin: 40px auto; /* Increased margin for better visual separation */
  padding: 20px; /* Add some internal padding */
  border-radius: 12px;
  box-shadow: 0 6px 20px rgba(0, 0, 0, 0.08); /* More pronounced shadow */
  background-color: #ffffff;
}

.card-header {
  display: flex;
  justify-content: center; /* Center the title */
  align-items: center;
  padding-bottom: 15px;
  border-bottom: 1px solid #ebeef5; /* Subtle separator */
  margin-bottom: 20px;
}

.card-title {
  font-size: 24px; /* Larger title */
  font-weight: bold;
  color: #303133;
}

.platform-form {
  padding: 0 20px; /* Padding for the form content */
}

.el-form-item {
  margin-bottom: 22px; /* Standardized spacing between form items */
}

.form-actions {
  margin-top: 30px; /* Space above action buttons */
  text-align: right; /* Align buttons to the right */
}

.el-button + .el-button {
  margin-left: 15px; /* Space between buttons */
}

.el-divider {
  margin: 30px 0; /* More vertical space for divider */
}

.divider-text {
  font-size: 16px;
  color: #606266;
  font-weight: bold;
}

.associated-info-section {
  text-align: center; /* Center the button */
  padding-bottom: 10px;
}

.view-associated-button {
  width: 80%; /* Make button wider */
  max-width: 350px; /* Limit max width */
  padding: 12px 20px; /* Larger padding for a bigger button */
  font-size: 16px;
  border-radius: 8px;
}

/* Responsive adjustments */
@media (max-width: 768px) {
  .platform-form-card {
    margin: 20px 10px; /* Smaller margins on small screens */
    padding: 15px;
  }

  .platform-form {
    padding: 0;
  }

  .el-form-item {
    margin-bottom: 18px;
  }

  .form-actions {
    text-align: center; /* Center buttons on small screens */
  }

  .el-button {
    width: 100%; /* Full width buttons */
    margin-left: 0 !important; /* Remove left margin for stacked buttons */
    margin-bottom: 10px; /* Add bottom margin for stacked buttons */
  }

  .view-associated-button {
    width: 100%;
  }
}
</style>