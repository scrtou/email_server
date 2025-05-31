<template>
  <el-card class="platform-registration-form-card">
    <template #header>
      <div class="card-header">
        <span class="card-title">{{ isEditMode ? '编辑平台注册信息' : '添加平台注册信息' }}</span>
      </div>
    </template>
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
              :disabled="isEditMode"
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
              :disabled="isEditMode"
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
              :placeholder="isEditMode ? '留空则不修改密码' : '请输入登录密码'"
              show-password
            />
          </el-form-item>
        </el-col>
      </el-row>
      
      <el-row :gutter="20">
        <el-col :xs="24" :sm="24" :md="12">
          <el-form-item label="备注" prop="notes">
            <el-input type="textarea" v-model="form.notes" :rows="3" placeholder="填写备注信息" />
          </el-form-item>
        </el-col>
        <el-col :xs="24" :sm="24" :md="12">
          <!--  确认密码字段已移除，此列保留用于布局对齐，如果不需要可以移除整个el-col -->
        </el-col>
      </el-row>

      <el-form-item class="form-actions">
        <el-button type="primary" @click="handleSubmit">
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
import { usePlatformRegistrationStore } from '@/stores/platformRegistration';
import { useEmailAccountStore } from '@/stores/emailAccount';
import { usePlatformStore } from '@/stores/platform';
import { ElMessage } from 'element-plus';

// eslint-disable-next-line no-undef
const props = defineProps({
  id: {
    type: [String, Number],
    default: null,
  },
});

const router = useRouter();
const route = useRoute();
const platformRegistrationStore = usePlatformRegistrationStore();
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

const isEditMode = computed(() => !!props.id || !!route.params.id);
const currentId = computed(() => props.id || route.params.id);

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

  if (isEditMode.value && currentId.value) {
    const regData = await platformRegistrationStore.fetchPlatformRegistrationById(currentId.value);
    if (regData) {
      form.value.email_account_id = regData.email_account_id;
      form.value.platform_id = regData.platform_id;
      form.value.login_username = regData.login_username;
      form.value.notes = regData.notes;
      // platformRegistrationStore.currentPlatformRegistration is already set by fetchPlatformRegistrationById
    } else {
      ElMessage.error('无法加载平台注册信息，可能ID无效');
      router.push({ name: 'PlatformRegistrationList' });
    }
  }
  loading.value = false;
});

const handleSubmit = async () => {
  if (!formRef.value) return;
  await formRef.value.validate(async (valid) => {
    if (valid) {
      loading.value = true;
      let success = false;
      if (isEditMode.value) {
        // 编辑模式逻辑保持不变，通常不允许更改 email_account_id 和 platform_id
        const payload = {
          login_username: form.value.login_username,
          notes: form.value.notes,
        };
        if (form.value.login_password) {
          payload.login_password = form.value.login_password;
        }
        // 注意：编辑模式下，email_account_id 和 platform_id 通常不应更改。
        // 如果需要更改，则应引导用户删除并重新创建。
        // 这里我们假设编辑模式不涉及 email_account_id 和 platform_id 的更改。
        // 如果表单允许更改它们，并且它们变成了字符串（新名称），则编辑逻辑也需要适配。
        // 为简单起见，此处假设编辑时不更改这两个字段。
         payload.email_account_id = form.value.email_account_id; // 确保传递，即使它们被禁用
         payload.platform_id = form.value.platform_id;       // 确保传递，即使它们被禁用


        success = await platformRegistrationStore.updatePlatformRegistration(currentId.value, payload);
      } else {
        // 创建模式逻辑
        const isEmailNew = typeof form.value.email_account_id === 'string';
        const isPlatformNew = typeof form.value.platform_id === 'string';

        let payload = {
          login_username: form.value.login_username,
          notes: form.value.notes,
        };
        if (form.value.login_password) {
          payload.login_password = form.value.login_password;
        }

        if (isEmailNew || isPlatformNew) { // One or both are new, use by-name API
            if (isEmailNew) {
                if (!form.value.email_account_id || String(form.value.email_account_id).trim() === '') {
                    ElMessage.error('新邮箱地址不能为空');
                    loading.value = false;
                    return;
                }
                payload.email_address = String(form.value.email_account_id).trim();
            } else { // Existing email
                const selectedEmail = emailAccountStore.emailAccounts.find(e => e.id === form.value.email_account_id);
                if (!selectedEmail) {
                    ElMessage.error('选择的邮箱账户无效');
                    loading.value = false;
                    return;
                }
                payload.email_address = selectedEmail.email_address; // Use address for by-name API
            }

            if (isPlatformNew) {
                if (!form.value.platform_id || String(form.value.platform_id).trim() === '') {
                    ElMessage.error('新平台名称不能为空');
                    loading.value = false;
                    return;
                }
                payload.platform_name = String(form.value.platform_id).trim();
            } else { // Existing platform
                const selectedPlatform = platformStore.platforms.find(p => p.id === form.value.platform_id);
                if (!selectedPlatform) {
                    ElMessage.error('选择的平台无效');
                    loading.value = false;
                    return;
                }
                payload.platform_name = selectedPlatform.name; // Use name for by-name API
            }
            success = await platformRegistrationStore.createPlatformRegistrationByName(payload);
        } else { // Both are existing, use by-id API
            payload.email_account_id = form.value.email_account_id;
            payload.platform_id = form.value.platform_id;
            if (!payload.email_account_id || !payload.platform_id) { // Ensure IDs are present
                 ElMessage.error('请选择有效的邮箱账户和平台');
                 loading.value = false;
                 return;
            }
            success = await platformRegistrationStore.createPlatformRegistration(payload);
        }
      }
      loading.value = false;
      if (success) {
        router.push({ name: 'PlatformRegistrationList' });
      }
    } else {
      ElMessage.error('请检查表单输入');
      return false;
    }
  });
};

const handleCancel = () => {
  router.push({ name: 'PlatformRegistrationList' });
};


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