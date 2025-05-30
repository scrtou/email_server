<template>
  <el-card class="platform-registration-form-card">
    <template #header>
      <span>{{ isEditMode ? '编辑平台注册信息' : '添加平台注册信息' }}</span>
    </template>
    <el-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-width="140px"
      v-loading="loading"
    >
      <el-form-item label="邮箱账户" prop="email_account_id">
        <el-select 
          v-model="form.email_account_id" 
          placeholder="选择或输入邮箱账户"
          filterable
          allow-create
          default-first-option
          :disabled="isEditMode"
          style="width: 100%;"
        >
          <el-option
            v-for="item in emailAccountStore.emailAccounts"
            :key="item.id"
            :label="item.email_address"
            :value="item.id"
          />
        </el-select>
      </el-form-item>

      <el-form-item label="平台" prop="platform_id">
        <el-select 
          v-model="form.platform_id" 
          placeholder="选择或输入平台名称"
          filterable
          allow-create
          default-first-option
          :disabled="isEditMode"
          style="width: 100%;"
        >
          <el-option
            v-for="item in platformStore.platforms"
            :key="item.id"
            :label="item.name"
            :value="item.id"
          />
        </el-select>
      </el-form-item>

      <el-form-item label="登录用户名/ID" prop="login_username">
        <el-input v-model="form.login_username" placeholder="在该平台的登录名或ID" />
      </el-form-item>

      <el-form-item label="登录密码" prop="login_password">
        <el-input
          type="password"
          v-model="form.login_password"
          :placeholder="isEditMode ? '留空则不修改密码' : '请输入登录密码'"
          show-password
        />
      </el-form-item>
      <el-form-item v-if="isEditMode && form.login_password" label="确认新密码" prop="confirm_password">
        <el-input
          type="password"
          v-model="form.confirm_password"
          placeholder="再次输入新密码"
          show-password
        />
      </el-form-item>

      <el-form-item label="备注" prop="notes">
        <el-input type="textarea" v-model="form.notes" :rows="3" />
      </el-form-item>

      <el-form-item>
        <el-button type="primary" @click="handleSubmit">
          {{ isEditMode ? '保存更新' : '立即创建' }}
        </el-button>
        <el-button @click="handleCancel">取消</el-button>
      </el-form-item>
    </el-form>
  </el-card>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'; // Removed watch
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
  confirm_password: '',
  notes: '',
});
const loading = ref(false);

const isEditMode = computed(() => !!props.id || !!route.params.id);
const currentId = computed(() => props.id || route.params.id);

const validatePassConfirm = (rule, value, callback) => {
  if (form.value.login_password) {
    if (value === '') {
      callback(new Error('请再次输入新密码'));
    } else if (value !== form.value.login_password) {
      callback(new Error('两次输入的新密码不一致'));
    } else {
      callback();
    }
  } else {
    callback();
  }
};

const rules = ref({
  email_account_id: [{ required: true, message: '请选择邮箱账户', trigger: 'change' }],
  platform_id: [{ required: true, message: '请选择平台', trigger: 'change' }],
  login_username: [{ max: 255, message: '登录用户名过长', trigger: 'blur' }],
  login_password: [
    // Password is no longer required
    { required: false, message: '请输入登录密码', trigger: 'blur' },
    { min: 6, message: '密码长度至少为6位', trigger: 'blur' },
  ],
  confirm_password: [
    { validator: validatePassConfirm, trigger: 'blur' }
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
    emailAccountStore.fetchEmailAccounts(1, 10000), // Fetch all for dropdown
    platformStore.fetchPlatforms(1, 10000) // Fetch all for dropdown
  ]);

  if (isEditMode.value && currentId.value) {
    const regData = await platformRegistrationStore.fetchPlatformRegistrationById(currentId.value);
    if (regData) {
      form.value.email_account_id = regData.email_account_id;
      form.value.platform_id = regData.platform_id;
      form.value.login_username = regData.login_username;
      form.value.notes = regData.notes;
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

        if (isEmailNew || isPlatformNew) {
          // 调用新的 by-name 接口
          if (isEmailNew) {
            payload.email_address = form.value.email_account_id;
          } else {
            payload.email_account_id = form.value.email_account_id;
          }

          if (isPlatformNew) {
            payload.platform_name = form.value.platform_id;
          } else {
            payload.platform_id = form.value.platform_id;
          }
          // 确保至少有一个是 name，或者两个都是 name
          if (!isEmailNew && !payload.email_account_id) {
            ElMessage.error('请选择或输入邮箱账户');
            loading.value = false;
            return;
          }
           if (!isPlatformNew && !payload.platform_id) {
            ElMessage.error('请选择或输入平台');
            loading.value = false;
            return;
          }


          success = await platformRegistrationStore.createPlatformRegistrationByName(payload);
        } else {
          // 调用旧的 by-id 接口
          payload.email_account_id = form.value.email_account_id;
          payload.platform_id = form.value.platform_id;
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
  max-width: 700px;
  margin: 20px auto;
}
</style>