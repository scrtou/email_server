<template>
  <div class="email-account-form-container">
    <!-- Card header content can be moved here if needed, or handled by ModalDialog -->
    <!-- For now, we remove the header as ModalDialog provides its own title mechanism -->
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
            <el-input
              type="password"
              v-model="form.password"
              :placeholder="isEditMode ? '留空则不修改密码' : '密码可选，留空则不设置密码'"
              show-password
            />
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
// import { useRouter, useRoute } from 'vue-router'; // No longer needed
import { ElMessage } from 'element-plus';
// import { View as ViewIcon } from '@element-plus/icons-vue';

// eslint-disable-next-line no-undef
const props = defineProps({
  // id: { // Replaced by emailAccount prop
  //   type: [String, Number],
  //   default: null,
  // },
  emailAccount: { // Used for editing, null for creating
    type: Object,
    default: null,
  },
  isEdit: { // Explicitly pass if it's edit mode
    type: Boolean,
    default: false,
  }
});

// eslint-disable-next-line no-undef
const emit = defineEmits(['submit-form', 'cancel']);

// const router = useRouter(); // No longer needed
// const route = useRoute(); // No longer needed

const emailAccountFormRef = ref(null);
const form = ref({
  email_address: '',
  password: '',
phone_number: '',
  // provider: '', // 已移除
  notes: '',
});
const loading = ref(false);

const isEditMode = computed(() => props.isEdit); // Use the new prop
// const currentId = computed(() => props.id || route.params.id); // Replaced by props.emailAccount.id


const rules = ref({
  email_address: [
    { required: true, message: '请输入邮箱地址', trigger: 'blur' },
    { type: 'email', message: '请输入有效的邮箱地址', trigger: 'blur' },
  ],
phone_number: [
    {
      // Basic validation for phone number (e.g., not empty, or a simple regex)
      // This can be enhanced with a more specific regex if needed.
      // For now, let's keep it simple or make it optional.
      // required: true, message: '请输入手机号', trigger: 'blur'
      // Example: Allow empty or match a simple pattern
      validator: (rule, value, callback) => {
        if (value && !/^\+?[0-9\s-]{7,20}$/.test(value)) { // Allows digits, spaces, hyphens, optional leading +
          callback(new Error('请输入有效的手机号'));
        } else {
          callback();
        }
      },
      trigger: 'blur'
    }
  ],
  password: [
    // Password is optional in both create and edit modes
    {
      validator: (rule, value, callback) => {
        // 密码在创建和编辑模式下都是可选的
        // 如果提供了密码，检查长度
        if (value && value.length < 6) {
          callback(new Error('密码长度至少为6位'));
          return;
        }
        callback();
      },
      trigger: 'blur'
    },
  ],
  // provider: [{ max: 100, message: '服务商名称过长', trigger: 'blur' }], // 已移除
});


// watch(isEditMode, (newVal) => {
//   rules.value.password[0].required = !newVal; // This logic is removed as password is no longer strictly required by frontend
// }, { immediate: true });


const resetForm = () => {
  if (emailAccountFormRef.value) {
    emailAccountFormRef.value.resetFields();
  }
  form.value = {
    email_address: '',
    password: '',
phone_number: '',
    notes: '',
  };
};

watch(() => props.emailAccount, (newAccount) => {
  resetForm();
  if (newAccount && isEditMode.value) {
    form.value.email_address = newAccount.email_address || '';
form.value.phone_number = newAccount.phone_number || '';
    form.value.notes = newAccount.notes || '';
    // Password and confirm_password remain blank for editing
  }
}, { immediate: true, deep: true });


onMounted(() => {
  // Data loading is now handled by the watcher for props.emailAccount
  // If it's not edit mode, the form will be blank by default due to resetForm.
  if (!isEditMode.value) {
    resetForm();
  }
});

const handleSubmit = async () => {
  console.log('[EmailAccountForm] handleSubmit called');
  if (!emailAccountFormRef.value) {
    console.log('[EmailAccountForm] emailAccountFormRef is null');
    return;
  }
  await emailAccountFormRef.value.validate(async (valid) => {
    console.log('[EmailAccountForm] Form validation result:', valid);
    if (valid) {
      loading.value = true; // Keep loading state for visual feedback if needed
      const payload = {
        email_address: form.value.email_address,
phone_number: form.value.phone_number,
        notes: form.value.notes,
      };

      if (form.value.password) {
        payload.password = form.value.password;
      }
      console.log('[EmailAccountForm] Emitting submit-form with payload:', payload);
      // The actual store call will be handled by the parent component
      emit('submit-form', payload);
      loading.value = false; // Reset loading state after emitting
    } else {
      console.log('[EmailAccountForm] Form validation failed');
      ElMessage.error('请检查表单输入');
      // Optionally emit an error event or let the parent handle UI feedback
      return false;
    }
  });
};


// eslint-disable-next-line no-undef
defineExpose({
  // handleSubmit is no longer exposed as it's an internal handler emitting an event
  // If parent needs to trigger validation, a new method can be exposed.
  // For now, ModalDialog's confirm button will trigger the form's submit logic
  // which in turn calls handleSubmit.
  // Expose resetForm if parent needs to call it directly, though it's also called on prop change.
  resetForm,
  // Expose the form reference if direct manipulation or validation check is needed from parent
  emailAccountFormRef,
  // Expose a method to trigger submission, which will then call internal handleSubmit
  triggerSubmit: handleSubmit
});
</script>

<style scoped>
.email-account-form-card {
  max-width: 800px; /* Slightly wider card */
  margin: 30px auto; /* More vertical margin */
  border-radius: 10px; /* Rounded corners */
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08); /* More pronounced shadow */
  background-color: #ffffff;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-bottom: 15px;
  border-bottom: 1px solid #ebeef5;
  margin-bottom: 20px;
}

.card-header span {
  font-size: 1.5rem; /* Larger title font */
  font-weight: bold;
  color: #303133;
}

.email-account-form {
  padding: 0 20px; /* Add some horizontal padding to the form */
}

.el-form-item {
  margin-bottom: 22px; /* Standardize item spacing */
}

.el-input,
.el-textarea {
  width: 100%; /* Ensure full width within form item */
}

.form-actions {
  margin-top: 30px; /* More space above action buttons */
  text-align: right; /* Align buttons to the right */
}

.form-actions .el-button {
  min-width: 100px; /* Ensure buttons have a consistent minimum width */
  font-size: 1rem;
  padding: 10px 20px;
  border-radius: 5px;
}

.el-divider {
  margin: 40px 0; /* More space around divider */
}

.divider-text {
  font-size: 1.1rem;
  font-weight: bold;
  color: #606266;
}

/* Responsive adjustments */
@media (max-width: 768px) {
  .email-account-form-card {
    margin: 15px; /* Reduce margin on smaller screens */
    padding: 10px;
  }

  .card-header span {
    font-size: 1.2rem; /* Adjust title font size */
  }

  .email-account-form {
    padding: 0 10px;
  }

  .form-actions {
    text-align: center; /* Center buttons on small screens */
  }

  .form-actions .el-button {
    width: 100%; /* Full width buttons on small screens */
    margin-bottom: 10px; /* Space between stacked buttons */
  }
}
</style>