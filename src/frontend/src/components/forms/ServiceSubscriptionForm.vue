<template>
  <div class="service-subscription-form-container">
    <!-- Card header content can be moved here if needed, or handled by ModalDialog -->
    <el-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-width="160px"
      v-loading="loading"
    >
      <el-form-item label="平台注册信息" prop="platform_registration_id">
        <el-select
          v-model="form.platform_registration_id"
          placeholder="选择关联的平台注册信息"
          filterable
          :disabled="isEditMode"
          style="width: 100%;"
        >
          <el-option
            v-for="item in platformRegistrationStore.platformRegistrations"
            :key="item.id"
            :label="`${item.platform_name} - ${item.email_address} (${item.login_username || '无用户名'})`"
            :value="item.id"
          />
        </el-select>
      </el-form-item>

      <el-form-item label="服务名称" prop="service_name">
        <el-input v-model="form.service_name" placeholder="例如：Google Workspace, Netflix Premium" />
      </el-form-item>

      <el-form-item label="描述" prop="description">
        <el-input type="textarea" v-model="form.description" :rows="2" placeholder="服务描述信息"/>
      </el-form-item>

      <el-form-item label="订阅状态" prop="status">
        <el-select v-model="form.status" placeholder="选择订阅状态" filterable style="width: 100%;">
          <el-option label="活跃 (active)" value="active" />
          <el-option label="已取消 (cancelled)" value="cancelled" />
          <el-option label="试用 (free_trial)" value="free_trial" />
          <el-option label="已过期 (expired)" value="expired" />
          <el-option label="其他 (other)" value="other" />
        </el-select>
      </el-form-item>

      <el-form-item label="费用金额" prop="cost">
        <el-input-number v-model="form.cost" :precision="2" :step="1" :min="0" style="width: 100%;" controls-position="right"/>
      </el-form-item>

      <el-form-item label="计费周期" prop="billing_cycle">
        <el-select v-model="form.billing_cycle" placeholder="选择计费周期" filterable style="width: 100%;">
          <el-option label="每月 (monthly)" value="monthly" />
          <el-option label="每年 (yearly)" value="yearly" />
          <el-option label="一次性 (onetime)" value="onetime" />
          <el-option label="免费 (free)" value="free" />
          <el-option label="其他 (other)" value="other" />
        </el-select>
      </el-form-item>

      <el-form-item label="下次续费日期" prop="next_renewal_date">
        <el-date-picker
          v-model="form.next_renewal_date"
          type="date"
          placeholder="选择日期"
          format="YYYY-MM-DD"
          value-format="YYYY-MM-DD"
          style="width: 100%;"
          clearable
        />
      </el-form-item>

      <el-form-item label="支付方式备注" prop="payment_method_notes">
        <el-input type="textarea" v-model="form.payment_method_notes" :rows="2" placeholder="例如：Visa **** 1234, PayPal"/>
      </el-form-item>

    </el-form>
  </div>
</template>

<script setup>
import { ref, onMounted, computed, watch } from 'vue';
// import { useRouter, useRoute } from 'vue-router'; // Removed
import { useServiceSubscriptionStore } from '@/stores/serviceSubscription';
import { usePlatformRegistrationStore } from '@/stores/platformRegistration';
import { ElMessage } from 'element-plus';

// eslint-disable-next-line no-undef
const props = defineProps({
  id: { // Used to determine if it's edit mode and to fetch/update
    type: [String, Number],
    default: null,
  },
  initialData: { // Used to populate form in edit mode
    type: Object,
    default: null,
  },
});

// eslint-disable-next-line no-undef
const emit = defineEmits(['submit-form', 'cancel']);

// const router = useRouter(); // Removed
// const route = useRoute(); // Removed
const serviceSubscriptionStore = useServiceSubscriptionStore();
const platformRegistrationStore = usePlatformRegistrationStore();

const formRef = ref(null);
const form = ref({
  platform_registration_id: null,
  service_name: '',
  description: '',
  status: 'active',
  cost: 0.00,
  billing_cycle: 'monthly',
  next_renewal_date: null,
  payment_method_notes: '',
});
const loading = ref(false);

const isEditMode = computed(() => !!props.id);

const rules = ref({
  platform_registration_id: [{ required: true, message: '请选择平台注册信息', trigger: 'change' }],
  service_name: [
    { required: true, message: '请输入服务名称', trigger: 'blur' },
    { max: 255, message: '服务名称过长', trigger: 'blur' },
  ],
  status: [{ required: true, message: '请选择订阅状态', trigger: 'change' }],
  cost: [{ type: 'number', message: '费用必须是数字', trigger: 'blur' }],
  billing_cycle: [{ required: true, message: '请选择计费周期', trigger: 'change' }],
});

const populateForm = (data) => {
  if (data) {
    form.value.platform_registration_id = data.platform_registration_id;
    form.value.service_name = data.service_name;
    form.value.description = data.description;
    form.value.status = data.status;
    form.value.cost = data.cost;
    form.value.billing_cycle = data.billing_cycle;
    form.value.next_renewal_date = data.next_renewal_date || null;
    form.value.payment_method_notes = data.payment_method_notes;
  } else {
    // Reset form for add mode or if data is null
    formRef.value?.resetFields(); // Reset validation and fields
    form.value = { // Explicitly reset data
      platform_registration_id: null,
      service_name: '',
      description: '',
      status: 'active',
      cost: 0.00,
      billing_cycle: 'monthly',
      next_renewal_date: null,
      payment_method_notes: '',
    };
  }
};


watch(() => props.initialData, (newData) => {
  populateForm(newData);
}, { immediate: true, deep: true });


onMounted(async () => {
  loading.value = true;
  if (platformRegistrationStore.platformRegistrations.length === 0) {
    await platformRegistrationStore.fetchPlatformRegistrations(
      1,
      10000, // Fetch all for dropdown
      { orderBy: 'platform_name', sortDirection: 'asc' },
      { email_account_id: null, platform_id: null }
    );
  }

  // If in edit mode and initialData is not yet populated (e.g. direct navigation for dev, though not typical for modal)
  // Or if initialData was passed but we want to ensure it's the freshest.
  // For modal usage, initialData from parent is preferred.
  // This explicit fetch might be redundant if ListView always passes fresh `initialData`.
  if (isEditMode.value && props.id && !props.initialData) {
    const subData = await serviceSubscriptionStore.fetchServiceSubscriptionById(props.id);
    if (subData) {
      populateForm(subData);
    } else {
      ElMessage.error('无法加载服务订阅数据，可能ID无效');
      // router.push({ name: 'ServiceSubscriptionList' }); // Cannot use router
      emit('cancel'); // Close modal if data load fails
    }
  } else if (props.initialData) {
      populateForm(props.initialData);
  } else {
      populateForm(null); // Ensure form is reset for add mode
  }
  loading.value = false;
});

const handleSubmit = async () => {
  if (!formRef.value) return;
  await formRef.value.validate(async (valid) => {
    if (valid) {
      loading.value = true;
      const payload = { ...form.value };
      if (!payload.next_renewal_date) {
        payload.next_renewal_date = null; // Ensure null is sent if date is cleared
      }
      // The actual store call will be handled by the parent component
      emit('submit-form', { payload, id: props.id, isEdit: isEditMode.value });
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
  resetForm: () => populateForm(null), // Use existing populateForm for reset logic
  formRef
});
</script>

<style scoped>
.service-subscription-form-card {
  max-width: 800px; /* Increased max-width for a more spacious layout */
  margin: 20px auto;
}
</style>