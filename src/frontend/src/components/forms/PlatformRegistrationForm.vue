<template>
  <div class="platform-registration-form-container">
    <!-- Card header content can be moved here if needed, or handled by ModalDialog -->
    <el-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-width="140px"
      v-loading="loading"
      class="platform-registration-form"
    >
      <el-row :gutter="20">
        <el-col :xs="24" :sm="24" :md="12">
          <el-form-item label="邮箱账户" prop="email_account_id">
            <el-select
              v-model="form.email_account_id"
              placeholder="选择或输入邮箱账户"
              filterable
              allow-create
              default-first-option
              :disabled="props.isEdit"
              class="full-width-select"
            >
              <el-option
                v-for="item in emailAccountStore.emailAccounts"
                :key="item.id"
                :label="item.email_address"
                :value="item.id"
              />
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :xs="24" :sm="24" :md="12">
          <el-form-item label="平台" prop="platform_id">
            <el-select
              v-model="form.platform_id"
              placeholder="选择或输入平台名称"
              filterable
              allow-create
              default-first-option
              :disabled="props.isEdit"
              class="full-width-select"
            >
              <el-option
                v-for="item in platformStore.platforms"
                :key="item.id"
                :label="item.name"
                :value="item.id"
              />
            </el-select>
          </el-form-item>
        </el-col>
      </el-row>

      <el-row :gutter="20">
        <el-col :xs="24" :sm="24" :md="12">
          <el-form-item label="登录用户名/ID" prop="login_username">
            <el-input v-model="form.login_username" placeholder="在该平台的登录名或ID" />
          </el-form-item>
        </el-col>
        <el-col :xs="24" :sm="24" :md="12">
          <el-form-item label="登录密码" prop="login_password">
            <el-input
              type="password"
              v-model="form.login_password"
              :placeholder="props.isEdit ? '留空则不修改密码' : '请输入登录密码'"
              show-password
            />
          </el-form-item>
        </el-col>
      </el-row>
      
      <el-row :gutter="20">
        <el-col :xs="24" :sm="24" :md="24">
          <el-form-item label="备注" prop="notes">
            <el-input type="textarea" v-model="form.notes" :rows="3" placeholder="填写备注信息" />
          </el-form-item>
        </el-col>
      </el-row>

    </el-form>
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue';
// import { useRouter, useRoute } from 'vue-router'; // Removed
import { useEmailAccountStore } from '@/stores/emailAccount';
import { usePlatformStore } from '@/stores/platform';
import { ElMessage } from 'element-plus';

// eslint-disable-next-line no-undef
const props = defineProps({
  platformRegistration: {
    type: Object,
    default: null,
  },
  isEdit: { // Renamed from isEditMode for clarity, and to receive from parent
    type: Boolean,
    default: false,
  }
});

// eslint-disable-next-line no-undef
const emit = defineEmits(['submit-form', 'cancel']);

// const router = useRouter(); // Removed
// const route = useRoute(); // Removed
const emailAccountStore = useEmailAccountStore();
const platformStore = usePlatformStore();

const formRef = ref(null);
const form = ref({
  email_account_id: null,
  platform_id: null,
  login_username: '',
  login_password: '',
  notes: '',
});
const loading = ref(false);

// const isEditMode = computed(() => !!props.id || !!route.params.id); // Replaced by props.isEdit
// const currentId = computed(() => props.id || route.params.id); // ID will come from props.platformRegistration.id

const rules = ref({
  email_account_id: [{ required: true, message: '请选择邮箱账户', trigger: 'change' }],
  platform_id: [{ required: true, message: '请选择平台', trigger: 'change' }],
  login_username: [{ max: 255, message: '登录用户名过长', trigger: 'blur' }],
  login_password: [
    // Password is no longer required
    { required: false, message: '请输入登录密码', trigger: 'blur' },
    { min: 6, message: '密码长度至少为6位', trigger: 'blur' },
  ],
});

// watch(() => form.value.login_username, (newVal) => {
//   if (!isEditMode.value) {
//     rules.value.login_password[0].required = newVal !== ''; // This logic is removed
//   }
// });
// watch(isEditMode, (newVal) => {
//   if (!newVal) { // Create mode
//     rules.value.login_password[0].required = form.value.login_username !== ''; // This logic is removed
//   } else { // Edit mode
//      rules.value.login_password[0].required = false; // Password not required for edit unless changing
//   }
// }, { immediate: true });


onMounted(async () => {
  loading.value = true;
  await Promise.all([
    emailAccountStore.fetchEmailAccounts(1, 10000, { orderBy: 'email_address', sortDirection: 'asc' }, { provider: '', emailAddressSearch: '' }), // Fetch all for dropdown, clear filters
    platformStore.fetchPlatforms(1, 10000, { orderBy: 'name', sortDirection: 'asc' }, { nameSearch: '' }) // Fetch all for dropdown, clear filters
  ]);

  // Populate form if in edit mode and data is provided
  if (props.isEdit && props.platformRegistration) {
    form.value.email_account_id = props.platformRegistration.email_account_id;
    form.value.platform_id = props.platformRegistration.platform_id;
    form.value.login_username = props.platformRegistration.login_username;
    form.value.notes = props.platformRegistration.notes;
    // Password is not pre-filled for editing
  } else {
    // Reset form for create mode or if no data
    form.value.email_account_id = null;
    form.value.platform_id = null;
    form.value.login_username = '';
    form.value.login_password = '';
    form.value.notes = '';
  }
  loading.value = false;
});

watch(() => props.platformRegistration, (newVal) => {
  if (props.isEdit && newVal) {
    form.value.email_account_id = newVal.email_account_id;
    form.value.platform_id = newVal.platform_id;
    form.value.login_username = newVal.login_username;
    form.value.notes = newVal.notes;
    form.value.login_password = ''; // Clear password on edit
  } else if (!props.isEdit) {
    formRef.value?.resetFields(); // Reset form for create mode
    form.value.email_account_id = null;
    form.value.platform_id = null;
    form.value.login_username = '';
    form.value.login_password = '';
    form.value.notes = '';
  }
}, { immediate: true, deep: true });


const handleSubmit = async () => {
  if (!formRef.value) return;
  await formRef.value.validate(async (valid) => {
    if (valid) {
      loading.value = true;
      const currentIdToUpdate = props.isEdit && props.platformRegistration ? props.platformRegistration.id : null;

      let payload = {
        login_username: form.value.login_username,
        notes: form.value.notes,
      };
      if (form.value.login_password) {
        payload.login_password = form.value.login_password;
      }

      if (props.isEdit) {
        if (!currentIdToUpdate) {
          ElMessage.error('编辑错误：缺少注册信息ID');
          loading.value = false;
          return;
        }
        // For update, include IDs as they are part of the form and might be expected by backend
        // even if not directly editable in the UI for this specific form's edit mode.
        payload.email_account_id = form.value.email_account_id;
        payload.platform_id = form.value.platform_id;
        // The store action will be called by the parent.
        emit('submit-form', { payload, id: currentIdToUpdate, isEdit: true });

      } else { // Create mode
        const isEmailNew = typeof form.value.email_account_id === 'string' && form.value.email_account_id.trim() !== '';
        const isPlatformNew = typeof form.value.platform_id === 'string' && form.value.platform_id.trim() !== '';

        if (isEmailNew || isPlatformNew) { // One or both are new
          if (isEmailNew) {
            payload.email_address = String(form.value.email_account_id).trim();
          } else {
            if (!form.value.email_account_id) {
                ElMessage.error('请选择邮箱账户');
                loading.value = false; return;
            }
            const selectedEmail = emailAccountStore.emailAccounts.find(e => e.id === form.value.email_account_id);
            if (!selectedEmail) { ElMessage.error('选择的邮箱账户无效'); loading.value = false; return; }
            payload.email_address = selectedEmail.email_address;
          }

          if (isPlatformNew) {
            payload.platform_name = String(form.value.platform_id).trim();
          } else {
             if (!form.value.platform_id) {
                ElMessage.error('请选择平台');
                loading.value = false; return;
            }
            const selectedPlatform = platformStore.platforms.find(p => p.id === form.value.platform_id);
            if (!selectedPlatform) { ElMessage.error('选择的平台无效'); loading.value = false; return; }
            payload.platform_name = selectedPlatform.name;
          }
          emit('submit-form', { payload, useByNameApi: true, isEdit: false });
        } else { // Both are existing
          if (!form.value.email_account_id || !form.value.platform_id) {
            ElMessage.error('请选择有效的邮箱账户和平台');
            loading.value = false; return;
          }
          payload.email_account_id = form.value.email_account_id;
          payload.platform_id = form.value.platform_id;
          emit('submit-form', { payload, useByNameApi: false, isEdit: false });
        }
      }
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
  resetForm: () => { // Expose a resetForm that also clears fields
    if (formRef.value) {
      formRef.value.resetFields();
    }
    form.value.email_account_id = null;
    form.value.platform_id = null;
    form.value.login_username = '';
    form.value.login_password = '';
    form.value.notes = '';
  },
  formRef
});
</script>

<style scoped>
.platform-registration-form-card {
  max-width: 800px; /* 增加最大宽度 */
  margin: 30px auto; /* 增加上下边距 */
  padding: 20px; /* 增加内边距 */
  border-radius: 8px; /* 圆角 */
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.08); /* 更明显的阴影 */
}

.card-header {
  padding-bottom: 15px;
  border-bottom: 1px solid #ebeef5;
  margin-bottom: 20px;
}

.card-title {
  font-size: 22px; /* 标题字体大小 */
  font-weight: bold;
  color: #303133;
}

.platform-registration-form .el-form-item {
  margin-bottom: 22px; /* 增加表单项间距 */
}

.full-width-select {
  width: 100%; /* 确保选择器占满宽度 */
}

.form-actions {
  margin-top: 30px; /* 按钮组顶部间距 */
  text-align: right; /* 按钮右对齐 */
}

.associated-details-section {
  margin-top: 30px; /* 关联信息部分顶部间距 */
  padding-top: 20px;
  border-top: 1px dashed #ebeef5; /* 虚线分隔 */
}

.associated-descriptions {
  margin-top: 15px; /* 描述列表顶部间距 */
}

.associated-info-section {
  margin-top: 20px;
  padding-top: 15px;
}

.view-associated-button {
  width: 100%; /* 按钮占满宽度 */
  padding: 12px 20px; /* 增加按钮内边距 */
  font-size: 16px; /* 按钮字体大小 */
}

/* 响应式调整 */
@media (max-width: 768px) {
  .platform-registration-form-card {
    margin: 15px; /* 移动端左右边距 */
    padding: 15px;
  }
  .card-title {
    font-size: 20px;
  }
  .platform-registration-form .el-form-item {
    margin-bottom: 18px;
  }
  .form-actions {
    text-align: center; /* 移动端按钮居中 */
  }
  .form-actions .el-button {
    width: 100%;
    margin-bottom: 10px;
    margin-left: 0 !important; /* 覆盖默认左边距 */
    margin-right: 0;
  }
}
</style>