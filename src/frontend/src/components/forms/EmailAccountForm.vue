<template>
  <div class="email-account-form-container">
    <el-form
      ref="emailAccountFormRef"
      :model="form"
      :rules="rules"
      label-width="120px"
      v-loading="loading"
      class="email-account-form"
    >
      <el-row :gutter="20">
        <el-col :span="24">
          <el-form-item label="邮箱地址" prop="email_address">
            <el-input v-model="form.email_address" placeholder="例如：user@example.com" :disabled="isEditMode" />
          </el-form-item>
        </el-col>
        <el-col :span="24">
          <el-form-item label="密码" prop="password">
            <!-- 密码输入区域和状态指示器水平对齐 -->
            <div class="password-main-container">
              <div class="password-input-container">
                <el-input
                  type="password"
                  v-model="form.password"
                  :placeholder="isEditMode ? '留空则不修改密码' : '密码可选，留空则不设置密码'"
                  show-password
                  class="password-input"
                />
                <!-- 查看现有密码按钮 -->
                <el-button
                  v-if="isEditMode && hasExistingPassword"
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
              <div v-if="isEditMode" class="password-status-inline">
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
       <el-col :span="24">
         <el-form-item label="IMAP服务器">
           <div style="display: flex; gap: 10px; align-items: center;">
             <el-input
               v-model="form.imap_server"
               placeholder="例如：imap.gmail.com"
               style="flex: 1; min-width: 0;"
             />
             <el-input
               v-model.number="form.imap_port"
               placeholder="端口"
               style="width: 100px; flex-shrink: 0;"
             />
           </div>
         </el-form-item>
       </el-col>
 <el-col :span="24">
           <el-form-item label="手机号" prop="phone_number">
            <el-input v-model="form.phone_number" placeholder="请输入手机号" />
          </el-form-item>
        </el-col>
        <el-col :span="24">
          <el-form-item label="备注" prop="notes">
            <el-input type="textarea" v-model="form.notes" :rows="4" placeholder="添加任何相关备注" />
          </el-form-item>
        </el-col>
      </el-row>
    </el-form>
  </div>
</template>

<script setup>
import { ref, onMounted, computed, watch } from 'vue';
import { ElMessage } from 'element-plus';
import { View, Lock, Unlock, CopyDocument } from '@element-plus/icons-vue';
import { useEmailAccountStore } from '@/stores/emailAccount';

const props = defineProps({
  emailAccount: {
    type: Object,
    default: null,
  },
  isEdit: {
    type: Boolean,
    default: false,
  }
});

const emit = defineEmits(['submit-form', 'cancel']);

const emailAccountStore = useEmailAccountStore();

const emailAccountFormRef = ref(null);
const form = ref({
  email_address: '',
  password: '',
  phone_number: '',
  notes: '',
  imap_server: '',
  imap_port: null,
});
const loading = ref(false);

// 查看密码相关状态
const viewPasswordLoading = ref(false);
const viewedPassword = ref('');

const isEditMode = computed(() => props.isEdit);

// 计算属性：检查是否有现有密码
const hasExistingPassword = computed(() => {
  return isEditMode.value && props.emailAccount && props.emailAccount.has_password;
});

const rules = ref({
  email_address: [
    { required: true, message: '请输入邮箱地址', trigger: 'blur' },
    { type: 'email', message: '请输入有效的邮箱地址', trigger: 'blur' },
  ],
  imap_port: [
    { type: 'number', message: '端口必须是数字', trigger: 'blur' },
  ],
  phone_number: [
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
  password: [
    {
      validator: (rule, value, callback) => {
        if (value && value.length < 6) {
          callback(new Error('密码长度至少为6位'));
          return;
        }
        callback();
      },
      trigger: 'blur'
    },
  ],
});

const resetForm = () => {
  if (emailAccountFormRef.value) {
    emailAccountFormRef.value.resetFields();
  }
  form.value = {
    email_address: '',
    password: '',
    phone_number: '',
    notes: '',
    imap_server: '',
    imap_port: null,
  };
};

watch(() => props.emailAccount, (newAccount) => {
  resetForm();
  if (newAccount && isEditMode.value) {
    form.value.email_address = newAccount.email_address || '';
    form.value.imap_server = newAccount.imap_server || '';
    form.value.imap_port = newAccount.imap_port || null;
    form.value.phone_number = newAccount.phone_number || '';
    form.value.notes = newAccount.notes || '';
  }
}, { immediate: true, deep: true });

onMounted(() => {
  if (!isEditMode.value) {
    resetForm();
  }
});

const handleSubmit = async () => {
  if (!emailAccountFormRef.value) return;
  await emailAccountFormRef.value.validate(async (valid) => {
    if (valid) {
      loading.value = true;
      const payload = {
        email_address: form.value.email_address,
        imap_server: form.value.imap_server,
        imap_port: form.value.imap_port,
        phone_number: form.value.phone_number,
        notes: form.value.notes,
      };

      if (form.value.password) {
        payload.password = form.value.password;
      }
      // If imap_port is an empty string from the input, convert it to null
      // so the backend's JSON parser, which expects a number or null for a *int field, doesn't fail.
      if (payload.imap_port === '') {
        payload.imap_port = null;
      }
      emit('submit-form', payload);
      loading.value = false;
    } else {
      ElMessage.error('请检查表单输入');
      return false;
    }
  });
};

// 查看密码方法
const handleViewPassword = async () => {
  if (!props.emailAccount?.id) {
    ElMessage.error('无法获取密码：缺少邮箱账户ID');
    return;
  }

  viewPasswordLoading.value = true;
  try {
    const password = await emailAccountStore.getPassword(props.emailAccount.id);
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

defineExpose({
  resetForm,
  emailAccountFormRef,
  triggerSubmit: handleSubmit
});
</script>

<style scoped>
.email-account-form-container {
  padding: 0 20px;
}
.el-form-item {
  margin-bottom: 22px;
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
  font-family: monospace;
  background-color: #f5f7fa;
  padding: 2px 6px;
  border-radius: 4px;
  color: #303133;
}
</style>