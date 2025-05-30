<template>
  <el-card class="email-account-form-card">
    <template #header>
      <span>{{ isEditMode ? '编辑邮箱账户' : '添加邮箱账户' }}</span>
    </template>
    <el-form
      ref="emailAccountFormRef"
      :model="form"
      :rules="rules"
      label-width="120px"
      v-loading="loading"
    >
      <el-form-item label="邮箱地址" prop="email_address">
        <el-input v-model="form.email_address" placeholder="例如：user@example.com" />
      </el-form-item>
      <el-form-item label="密码" prop="password">
        <el-input
          type="password"
          v-model="form.password"
          :placeholder="isEditMode ? '留空则不修改密码' : '请输入密码'"
          show-password
        />
      </el-form-item>
      <el-form-item v-if="isEditMode" label="确认新密码" prop="confirm_password">
        <el-input
          type="password"
          v-model="form.confirm_password"
          placeholder="再次输入新密码"
          show-password
        />
      </el-form-item>
      <!-- 服务商字段已移除，将由后端自动从邮箱地址提取 -->
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
import { useEmailAccountStore } from '@/stores/emailAccount';
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
const emailAccountStore = useEmailAccountStore();

const emailAccountFormRef = ref(null);
const form = ref({
  email_address: '',
  password: '',
  confirm_password: '', // Only for edit mode password change confirmation
  // provider: '', // 已移除
  notes: '',
});
const loading = ref(false);

const isEditMode = computed(() => !!props.id || !!route.params.id);
const currentId = computed(() => props.id || route.params.id);


const validatePassConfirm = (rule, value, callback) => {
  if (form.value.password) { // Only validate if new password is set
    if (value === '') {
      callback(new Error('请再次输入新密码'));
    } else if (value !== form.value.password) {
      callback(new Error('两次输入的新密码不一致'));
    } else {
      callback();
    }
  } else {
    callback(); // If no new password, confirmation is not needed
  }
};

const rules = ref({
  email_address: [
    { required: true, message: '请输入邮箱地址', trigger: 'blur' },
    { type: 'email', message: '请输入有效的邮箱地址', trigger: 'blur' },
  ],
  password: [
    // Password is no longer required
    { required: false, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码长度至少为6位', trigger: 'blur' },
  ],
  confirm_password: [
    // Required only if password field has a value in edit mode
    { validator: validatePassConfirm, trigger: 'blur' }
  ],
  // provider: [{ max: 100, message: '服务商名称过长', trigger: 'blur' }], // 已移除
});


// watch(isEditMode, (newVal) => {
//   rules.value.password[0].required = !newVal; // This logic is removed as password is no longer strictly required by frontend
// }, { immediate: true });


onMounted(async () => {
  if (isEditMode.value && currentId.value) {
    loading.value = true;
    const accountData = await emailAccountStore.fetchEmailAccountById(currentId.value);
    if (accountData) {
      form.value.email_address = accountData.email_address;
      // form.value.provider = accountData.provider; // 已移除，provider 由后端处理
      form.value.notes = accountData.notes;
      // Password is not pre-filled for security
    } else {
      ElMessage.error('无法加载邮箱账户数据，可能ID无效');
      router.push({ name: 'EmailAccountList' }); // Or some error page
    }
    loading.value = false;
  }
});

const handleSubmit = async () => {
  if (!emailAccountFormRef.value) return;
  await emailAccountFormRef.value.validate(async (valid) => {
    if (valid) {
      loading.value = true;
      const payload = {
        email_address: form.value.email_address,
        // provider: form.value.provider, // 已移除，provider 由后端处理
        notes: form.value.notes,
      };
      if (isEditMode.value) {
        // In edit mode, only include password if it's provided (for changing password)
        if (form.value.password) {
          payload.password = form.value.password;
        }
      } else {
        // In create mode, password is required by backend validation based on rules
        // Frontend validation should ensure it's not empty.
        payload.password = form.value.password;
      }

      let success = false;
      if (isEditMode.value) {
        success = await emailAccountStore.updateEmailAccount(currentId.value, payload);
      } else {
        success = await emailAccountStore.createEmailAccount(payload);
      }
      loading.value = false;
      if (success) {
        router.push({ name: 'EmailAccountList' }); // Navigate back to list on success
      }
    } else {
      ElMessage.error('请检查表单输入');
      return false;
    }
  });
};

const handleCancel = () => {
  router.push({ name: 'EmailAccountList' });
};
</script>

<style scoped>
.email-account-form-card {
  max-width: 700px;
  margin: 20px auto;
}
</style>