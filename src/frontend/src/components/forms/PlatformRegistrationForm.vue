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
      <!-- 第一行：平台和邮箱账户 -->
      <el-row :gutter="20">
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
        <el-col :xs="24" :sm="24" :md="12">
          <el-form-item label="邮箱账户" prop="email_account_id">
            <el-select
              v-model="form.email_account_id"
              placeholder="选择或输入邮箱账户"
              filterable
              allow-create
              default-first-option
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
      </el-row>

      <!-- 第二行：用户名和手机号，对齐上一行 -->
      <el-row :gutter="20">
        <el-col :xs="24" :sm="24" :md="12">
          <el-form-item label="用户名/ID" prop="login_username">
            <el-input
              v-model="form.login_username"
              placeholder="请输入用户名/ID"
              clearable
            />
          </el-form-item>
        </el-col>
        <el-col :xs="24" :sm="24" :md="12">
          <el-form-item label="手机号" prop="phone_number">
            <el-input v-model="form.phone_number" placeholder="请输入手机号" clearable />
          </el-form-item>
        </el-col>
      </el-row>

      <!-- 第三行：登录密码 -->
      <el-row :gutter="20">
        <el-col :xs="24" :sm="24" :md="24">
          <el-form-item label="登录密码" prop="login_password">
            <!-- 密码输入区域和状态指示器水平对齐 -->
            <div class="password-main-container">
              <div class="password-input-container">
                <el-input
                  type="password"
                  v-model="form.login_password"
                  :placeholder="props.isEdit ? '留空则不修改密码' : '请输入登录密码'"
                  show-password
                  class="password-input"
                />
                <!-- 查看现有密码按钮 -->
                <el-button
                  v-if="props.isEdit && hasExistingPassword"
                  type="primary"
                  size="small"
                  :icon="View"
                  @click="handleViewPassword"
                  :loading="viewPasswordLoading"
                  class="view-password-btn"
                  title="查看当前密码"
                >
                  查看
                </el-button>
              </div>

              <!-- 密码状态指示器 - 与输入框和按钮水平对齐 -->
              <div v-if="props.isEdit" class="password-status-inline">
                <el-tag
                  v-if="hasExistingPassword"
                  type="success"
                  size="small"
                  effect="plain"
                >
                  <el-icon><Lock /></el-icon>
                  已设置密码
                </el-tag>
                <el-tag
                  v-else
                  type="info"
                  size="small"
                  effect="plain"
                >
                  <el-icon><Unlock /></el-icon>
                  未设置密码
                </el-tag>
              </div>
            </div>

            <!-- 显示查看到的密码 -->
            <div v-if="viewedPassword" class="viewed-password">
              <el-alert
                type="info"
                :closable="true"
                @close="viewedPassword = ''"
                show-icon
              >
                <template #default>
                  <div class="password-display">
                    <span class="password-label">当前密码：</span>
                    <span class="password-text">{{ viewedPassword }}</span>
                    <el-button
                      type="text"
                      :icon="CopyDocument"
                      @click="copyPassword"
                      size="small"
                    >
                      复制
                    </el-button>
                  </div>
                </template>
              </el-alert>
            </div>
          </el-form-item>
        </el-col>
      </el-row>

      <!-- 第四行：备注字段，占据全宽，左侧对齐 -->
      <el-row :gutter="20">
        <el-col :span="24">
          <el-form-item label="备注" prop="notes" class="notes-form-item">
            <el-input type="textarea" v-model="form.notes" :rows="3" placeholder="填写备注信息" />
          </el-form-item>
        </el-col>
      </el-row>

    </el-form>
  </div>
</template>


<script setup>
import { ref, onMounted, watch, computed } from 'vue';
import { useEmailAccountStore } from '@/stores/emailAccount';
import { usePlatformStore } from '@/stores/platform';
import { ElMessage } from 'element-plus';
import { Lock, Unlock, View, CopyDocument } from '@element-plus/icons-vue';
import { usePlatformRegistrationStore } from '@/stores/platformRegistration';

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

const emailAccountStore = useEmailAccountStore();
const platformStore = usePlatformStore();
const platformRegistrationStore = usePlatformRegistrationStore();

const formRef = ref(null);
const form = ref({
  email_account_id: null,
  platform_id: null,
  login_username: '',
  login_password: '',
  phone_number: '', // Added phone_number
  notes: '',
});
const loading = ref(false); // General form loading

// 查看密码相关状态
const viewPasswordLoading = ref(false);
const viewedPassword = ref('');

// 计算属性：检查是否有现有密码
const hasExistingPassword = computed(() => {
  return props.isEdit && props.platformRegistration && props.platformRegistration.has_password;
});

const rules = ref({
  email_account_id: [{ required: false, message: '请选择邮箱账户', trigger: 'change' }], // Removed required validation
  platform_id: [{ required: true, message: '请选择平台', trigger: 'change' }],
  login_username: [{ max: 255, message: '登录用户名过长', trigger: 'blur' }],
  phone_number: [ // Added phone_number validation
    {
      validator: (rule, value, callback) => {
        if (value && !/^\+?[0-9\s-]{7,20}$/.test(value)) {
          callback(new Error('请输入有效的手机号'));
        } else {
          callback();
        }
      },
      trigger: 'blur'
    }
  ],
  login_password: [
    { required: false, message: '请输入登录密码', trigger: 'blur' },
    { min: 6, message: '密码长度至少为6位', trigger: 'blur' },
  ],
});



onMounted(async () => {
  loading.value = true;
  await Promise.all([
    emailAccountStore.fetchEmailAccounts(1, 10000, { orderBy: 'email_address', sortDirection: 'asc' }, { provider: '', emailAddressSearch: '' }),
    platformStore.fetchPlatforms(1, 10000, { orderBy: 'name', sortDirection: 'asc' }, { nameSearch: '' })
  ]);

  // Populate form if in edit mode and data is provided
  if (props.isEdit && props.platformRegistration) {
    // 处理 email_account_id：如果为 0 则设置为 null（表示未关联邮箱）
    form.value.email_account_id = props.platformRegistration.email_account_id === 0 ? null : props.platformRegistration.email_account_id;
    form.value.platform_id = props.platformRegistration.platform_id;
    form.value.login_username = props.platformRegistration.login_username;
    form.value.phone_number = props.platformRegistration.phone_number || ''; // Populate phone_number
    form.value.notes = props.platformRegistration.notes;
  } else {
    // Reset form for create mode or if no data
    form.value.email_account_id = null;
    form.value.platform_id = null;
    form.value.login_username = '';
    form.value.login_password = '';
    form.value.phone_number = ''; // Reset phone_number
    form.value.notes = '';
  }
  loading.value = false;
});

watch(() => props.platformRegistration, (newVal) => {
  if (props.isEdit && newVal) {
    // 处理 email_account_id：如果为 0 则设置为 null（表示未关联邮箱）
    form.value.email_account_id = newVal.email_account_id === 0 ? null : newVal.email_account_id;
    form.value.platform_id = newVal.platform_id;
    form.value.login_username = newVal.login_username;
    form.value.phone_number = newVal.phone_number || ''; // Populate phone_number
    form.value.notes = newVal.notes;
    form.value.login_password = ''; // Clear password on edit
  } else if (!props.isEdit) {
    formRef.value?.resetFields(); // Reset form for create mode
    form.value.email_account_id = null;
    form.value.platform_id = null;
    form.value.login_username = '';
    form.value.login_password = '';
    form.value.phone_number = ''; // Reset phone_number
    form.value.notes = '';
  }
}, { immediate: true, deep: true });

const handleSubmit = async () => {
  if (!formRef.value) return;
  await formRef.value.validate(async (valid) => {
    // Add custom validation: Username and Email cannot both be empty
    if (!form.value.login_username && !form.value.email_account_id) {
        ElMessage.error('用户名/ID 和 邮箱账户 不能同时为空');
        return false; // Prevent submission
    }

    if (valid) {
      loading.value = true;
      const currentIdToUpdate = props.isEdit && props.platformRegistration ? props.platformRegistration.id : null;

      let payload = {
        login_username: form.value.login_username,
        phone_number: form.value.phone_number, // Add phone_number to payload
        notes: form.value.notes,
      };
      if (form.value.login_password) {
        payload.login_password = form.value.login_password;
      }

      try {
        if (props.isEdit) {
          if (!currentIdToUpdate) {
            ElMessage.error('编辑错误：缺少注册信息ID');
            return;
          }
          // For update, include IDs as they are part of the form and might be expected by backend
          // even if not directly editable in the UI for this specific form's edit mode.
          // 如果 email_account_id 为 null，则不包含在 payload 中，让后端处理为 NULL
          if (form.value.email_account_id !== null) {
            payload.email_account_id = form.value.email_account_id;
          }
          payload.platform_id = form.value.platform_id;
          // The store action will be called by the parent.
          emit('submit-form', { payload, id: currentIdToUpdate, isEdit: true });

        } else { // Create mode
          const isEmailNew = typeof form.value.email_account_id === 'string' && form.value.email_account_id.trim() !== '';
          const isPlatformNew = typeof form.value.platform_id === 'string' && form.value.platform_id.trim() !== '';

          if (isEmailNew || isPlatformNew) { // One or both are new
            if (isEmailNew) {
              payload.email_address = String(form.value.email_account_id).trim();
            } else if (form.value.email_account_id) { // Only process if an existing email IS selected
              // If email_account_id is null/undefined, we skip this block, allowing username-only submission
              const selectedEmail = emailAccountStore.emailAccounts.find(e => e.id === form.value.email_account_id);
              if (!selectedEmail) {
                ElMessage.error('选择的邮箱账户无效');
                return;
              }
              payload.email_address = selectedEmail.email_address;
            }

            if (isPlatformNew) {
              payload.platform_name = String(form.value.platform_id).trim();
            } else {
               if (!form.value.platform_id) {
                  ElMessage.error('请选择平台');
                  return;
              }
              const selectedPlatform = platformStore.platforms.find(p => p.id === form.value.platform_id);
              if (!selectedPlatform) {
                ElMessage.error('选择的平台无效');
                return;
              }
              payload.platform_name = selectedPlatform.name;
            }
            emit('submit-form', { payload, useByNameApi: true, isEdit: false });
          } else { // Both are existing (selected from dropdown)
            // Removed the check for email_account_id.
            // platform_id is already validated by formRef.value.validate based on rules.
            // We still need to assign them to the payload if they exist.
            if (form.value.email_account_id) { // Assign if selected
               payload.email_account_id = form.value.email_account_id;
            }
            // platform_id is required by rules, so it should exist here if validation passed.
            payload.platform_id = form.value.platform_id;
            emit('submit-form', { payload, useByNameApi: false, isEdit: false });
          }
        }
      } finally {
        // 重置加载状态，但不影响父组件的加载状态管理
        loading.value = false;
      }
    } else {
      ElMessage.error('请检查表单输入');
      return false;
    }
  });
};

// 查看密码方法
const handleViewPassword = async () => {
  if (!props.platformRegistration?.id) {
    ElMessage.error('无法获取密码：缺少注册信息ID');
    return;
  }

  viewPasswordLoading.value = true;
  try {
    const password = await platformRegistrationStore.getPassword(props.platformRegistration.id);
    if (password) {
      viewedPassword.value = password;
      ElMessage.success('密码获取成功');
    } else {
      ElMessage.warning('未找到密码信息');
    }
  } catch (error) {
    ElMessage.error('获取密码失败：' + (error.message || '未知错误'));
  } finally {
    viewPasswordLoading.value = false;
  }
};

// 复制密码方法
const copyPassword = async () => {
  if (!viewedPassword.value) {
    ElMessage.warning('没有可复制的密码');
    return;
  }

  try {
    await navigator.clipboard.writeText(viewedPassword.value);
    ElMessage.success('密码已复制到剪贴板');
  } catch (error) {
    // 降级方案：使用传统的复制方法
    const textArea = document.createElement('textarea');
    textArea.value = viewedPassword.value;
    document.body.appendChild(textArea);
    textArea.select();
    try {
      document.execCommand('copy');
      ElMessage.success('密码已复制到剪贴板');
    } catch (fallbackError) {
      ElMessage.error('复制失败，请手动复制');
    }
    document.body.removeChild(textArea);
  }
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
    form.value.phone_number = ''; // Reset phone_number
    form.value.notes = '';
    viewedPassword.value = ''; // 清空查看的密码
  },
  formRef
});
</script>

<style scoped>
/* 弹框内容统一边距 - 移除额外的padding，因为ModalDialog已经有20px的padding */
.platform-registration-form {
  padding: 0; /* 移除额外的左右内边距 */
  box-sizing: border-box;
}

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

/* 确保所有行都有一致的边距 */
.platform-registration-form .el-row {
  margin-left: 0 !important;
  margin-right: 0 !important;
}

.platform-registration-form .el-col {
  padding-left: 8px;
  padding-right: 8px;
}

/* 第一列和最后一列的特殊处理 */
.platform-registration-form .el-row .el-col:first-child {
  padding-left: 0;
}

.platform-registration-form .el-row .el-col:last-child {
  padding-right: 0;
}

/* 单列布局时不需要左右padding */
.platform-registration-form .el-row .el-col[class*="24"] {
  padding-left: 0 !important;
  padding-right: 0 !important;
}

.full-width-select {
  width: 100%; /* 确保选择器占满宽度 */
}

/* 备注字段左侧对齐 */
.notes-form-item :deep(.el-form-item__content) {
  text-align: left;
}

.notes-form-item :deep(.el-textarea__inner) {
  text-align: left;
}

/* 密码主容器 - 水平布局 */
.password-main-container {
  display: flex;
  gap: 12px;
  align-items: center;
  flex-wrap: wrap;
}

.password-input-container {
  display: flex;
  gap: 8px;
  align-items: center;
  flex: 1;
  min-width: 300px; /* 确保输入框有最小宽度 */
}

.password-input {
  flex: 1;
}

.view-password-btn {
  flex-shrink: 0;
  margin-top: 0;
}

/* 密码状态指示器 - 内联显示 */
.password-status-inline {
  flex-shrink: 0;
  display: flex;
  align-items: center;
}

.password-status-inline .el-tag {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

/* 响应式：小屏幕时垂直排列 */
@media (max-width: 768px) {
  .password-main-container {
    flex-direction: column;
    align-items: stretch;
    gap: 8px;
  }

  .password-input-container {
    min-width: auto;
  }

  .password-status-inline {
    justify-content: flex-start;
  }
}

.viewed-password {
  margin-top: 8px;
}

.password-display {
  display: flex;
  align-items: center;
  gap: 8px;
}

.password-label {
  font-weight: 500;
  color: #606266;
}

.password-text {
  font-family: 'Courier New', monospace;
  background-color: #f5f7fa;
  padding: 4px 8px;
  border-radius: 4px;
  border: 1px solid #dcdfe6;
  font-size: 14px;
  color: #303133;
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
  .platform-registration-form {
    padding: 0 12px; /* 移动端进一步减少左右内边距 */
  }

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

  /* 移动端列间距调整 */
  .platform-registration-form .el-col {
    padding-left: 0 !important;
    padding-right: 0 !important;
    margin-bottom: 16px;
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