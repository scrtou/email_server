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

  </el-card>
</template>

<script setup>
import { ref, onMounted, computed, reactive } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { usePlatformStore } from '@/stores/platform';
import { ElMessage, ElButton, ElCard, ElForm, ElFormItem, ElInput } from 'element-plus';

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
});
const loading = ref(false);

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
}
</style>