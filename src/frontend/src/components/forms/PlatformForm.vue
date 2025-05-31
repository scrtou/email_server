<template>
  <div class="platform-form-container">
    <!-- Card header content can be moved here if needed, or handled by ModalDialog -->
    <el-form
      ref="platformFormRef"
      :model="form"
      :rules="rules"
      label-width="120px"
      v-loading="loading"
      class="platform-form"
    >
      <el-form-item label="平台名称" prop="name">
        <el-input v-model="form.name" placeholder="例如：Google, GitHub, Steam" clearable :disabled="isEditMode" />
      </el-form-item>
      <el-form-item label="平台网址" prop="website_url">
        <el-input v-model="form.website_url" placeholder="例如：https://www.google.com" clearable />
      </el-form-item>
      <el-form-item label="备注" prop="notes">
        <el-input type="textarea" v-model="form.notes" :rows="4" resize="vertical" show-word-limit maxlength="500" />
      </el-form-item>
    </el-form>
  </div>
</template>

<script setup>
import { ref, onMounted, computed, watch } from 'vue'; // 引入 watch
// import { useRouter, useRoute } from 'vue-router'; // 移除 useRouter 和 useRoute
import { usePlatformStore } from '@/stores/platform';
import { ElMessage, ElForm, ElFormItem, ElInput } from 'element-plus';

// eslint-disable-next-line no-undef
const props = defineProps({
  id: { // 用于编辑模式，传入平台ID
    type: [String, Number],
    default: null,
  },
  initialData: { // 用于接收初始数据，编辑模式下为平台对象，新增模式下可为空对象
    type: Object,
    default: () => ({}),
  }
});

// eslint-disable-next-line no-undef
const emit = defineEmits(['submit-form', 'cancel']); // 定义 emit

// const router = useRouter(); // 移除 router
// const route = useRoute(); // 移除 route
const platformStore = usePlatformStore();

const platformFormRef = ref(null);
const form = ref({
  name: '',
  website_url: '',
  notes: '',
});
const loading = ref(false);

const isEditMode = computed(() => !!props.id); // 编辑模式判断简化为只依赖 props.id
// const currentId = computed(() => props.id || route.params.id); // currentId 直接使用 props.id

const rules = ref({
  name: [
    { required: true, message: '请输入平台名称', trigger: 'blur' },
    { min: 2, max: 100, message: '名称长度应为 2 到 100 个字符', trigger: 'blur' },
  ],
  website_url: [
    { type: 'url', message: '请输入有效的网址', trigger: ['blur', 'change'] },
  ],
});

const resetForm = () => {
  form.value = {
    name: '',
    website_url: '',
    notes: '',
  };
  if (platformFormRef.value) {
    platformFormRef.value.clearValidate();
  }
};

watch(() => props.initialData, (newData) => {
  resetForm(); // 每次 initialData 变化时重置表单
  if (newData && Object.keys(newData).length > 0) {
    form.value.name = newData.name || '';
    form.value.website_url = newData.website_url || '';
    form.value.notes = newData.notes || '';
  }
}, { immediate: true, deep: true });


onMounted(async () => {
  // onMounted 中不再需要根据 route.params.id 加载数据，依赖 props.initialData
  if (isEditMode.value && props.id) {
    // 如果 initialData 为空但有 id，理论上父组件应该已经获取并传入
    // 但作为备用，可以考虑是否需要再次 fetch，当前假设 initialData 会被正确填充
    if (!props.initialData || Object.keys(props.initialData).length === 0) {
        loading.value = true;
        const platformData = await platformStore.fetchPlatformById(props.id);
        if (platformData) {
          form.value.name = platformData.name;
          form.value.website_url = platformData.website_url;
          form.value.notes = platformData.notes;
        } else {
          ElMessage.error('无法加载平台数据，ID可能无效');
          // 此处不再进行路由跳转，由父组件处理弹窗关闭
        }
        loading.value = false;
    }
  } else {
    resetForm(); // 新增模式，确保表单清空
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
      } else {
        payload.website_url = ''; // 确保如果为空字符串也提交
      }
      emit('submit-form', payload);
      loading.value = false;
    } else {
      ElMessage.error('请检查表单输入');
      return false;
    }
  });
};


// eslint-disable-next-line no-undef
defineExpose({
  triggerSubmit: handleSubmit,
  resetForm,
  platformFormRef
});
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