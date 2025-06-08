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
            <el-input
              type="password"
              v-model="form.password"
              :placeholder="isEditMode ? '留空则不修改密码' : '密码可选，留空则不设置密码'"
              show-password
            />
          </el-form-item>
        </el-col>
       <el-col :span="24">
         <el-form-item label="IMAP 服务器 / 端口">
           <el-row :gutter="10">
             <el-col :span="16">
               <el-input v-model="form.imap_server" placeholder="例如：imap.gmail.com" />
             </el-col>
             <el-col :span="8">
               <el-input v-model.number="form.imap_port" placeholder="端口" />
             </el-col>
           </el-row>
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

const isEditMode = computed(() => props.isEdit);

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
</style>