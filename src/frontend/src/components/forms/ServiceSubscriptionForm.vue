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
      <el-form-item v-if="!isEditMode" label="平台名称" prop="platform_name">
        <el-select
          v-model="form.platform_name"
          placeholder="选择或输入平台名称"
          filterable
          allow-create
          default-first-option
          :reserve-keyword="false"
          style="width: 100%;"
        >
          <el-option
            v-for="platform in platformStore.platforms"
            :key="platform.id"
            :label="platform.name"
            :value="platform.name"
          />
        </el-select>
      </el-form-item>
      <el-form-item v-if="isEditMode" label="平台名称" prop="platform_name">
        <el-input v-model="form.platform_name" disabled />
      </el-form-item>

      <!-- Add Username Field -->
      <el-form-item label="用户名/ID" prop="login_username_input">
        <el-select
          v-model="form.login_username_input"
          placeholder="选择或输入用户名/ID"
          filterable
          allow-create
          clearable
          default-first-option
          :reserve-keyword="false"
          style="width: 100%;"
          :loading="loadingUsernames"
          @change="handleUsernameChange"
        >
          <el-option
            v-for="user in platformUsernames"
            :key="user.id"
            :label="user.login_username ? user.login_username : `[无用户名]${user.email_address ? ` (邮箱: ${user.email_address})` : ''}`"
            :value="user.id"
          />
          <!-- If allow-create is used with el-select, when user types something not in options and hits enter,
               the input value itself becomes the model value.
               We need to ensure `handleUsernameChange` correctly interprets this.
               The `value` in `@change` will be the `id` if an option is selected,
               or the typed string if a new item is "created".
               It's important that `form.login_username_input` (the v-model) correctly reflects the text input
               when a new item is created, and `form.selected_username_registration_id` is set when an item is selected.
               The `el-select` with `allow-create` might bind the input text to `form.login_username_input` directly when creating.
               If an option is selected, `form.login_username_input` (v-model) will be bound to the `value` of el-option (which is `user.id`).
               This means we need to adjust `handleUsernameChange` or how v-model is used.

               Alternative: Use `v-model="form.selected_username_registration_id"` for selection,
               and a separate input for manual entry, or a more complex combobox.

               Given the current setup with `v-model="form.login_username_input"` and `allow-create`,
               when an item is selected, `form.login_username_input` becomes the `id`.
               We need `form.login_username_input` to hold the *text* of the username.
               Let's change `v-model` to `form.selected_username_registration_id` when an item is *selected*,
               and use `form.login_username_input` for the *textual input/creation*.
               This is tricky with a single `el-select`.

               A common pattern for `el-select` with `allow-create` and wanting to distinguish selected ID vs created text:
               The `v-model` will hold the ID if selected, or the text if created.
               We can check the type of the `v-model`'s value in `handleUsernameChange`.
            -->
        </el-select>
        <div v-if="loadingUsernames" class="el-form-item__error">正在加载用户名...</div>
      </el-form-item>

      <el-form-item v-if="!isEditMode" label="邮箱地址" prop="email_address">
        <el-select
          v-model="form.email_address"
          placeholder="选择或输入邮箱地址"
          filterable
          allow-create
          default-first-option
          :reserve-keyword="false"
          style="width: 100%;"
        >
          <el-option
            v-for="account in emailAccountStore.emailAccounts"
            :key="account.id"
            :label="account.email_address"
            :value="account.email_address"
          />
        </el-select>
      </el-form-item>
      <el-form-item v-if="isEditMode" label="邮箱地址" prop="email_address">
        <el-input v-model="form.email_address" :disabled="isEditMode" />
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
import { usePlatformRegistrationStore } from '@/stores/platformRegistration'; // Added
import { usePlatformStore } from '@/stores/platform';
import { useEmailAccountStore } from '@/stores/emailAccount';
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
const platformRegistrationStore = usePlatformRegistrationStore(); // Added
const platformStore = usePlatformStore();
const emailAccountStore = useEmailAccountStore();

const formRef = ref(null);
const form = ref({
  platform_name: '',
  email_address: '',
  login_username_input: '', // Renamed from login_username, for el-select v-model
  selected_username_registration_id: null, // To store ID if selected from dropdown
  service_name: '',
  description: '',
  status: 'active',
  cost: 0.00,
  billing_cycle: 'monthly',
  next_renewal_date: null,
  payment_method_notes: '',
});
const loading = ref(false);
const platformUsernames = ref([]); // For storing fetched usernames { id, login_username }
const loadingUsernames = ref(false); // Loading state for usernames

const isEditMode = computed(() => !!props.id);

const rules = ref({
  platform_name: [{ required: true, message: '请输入或选择平台名称', trigger: 'change' }],
  login_username_input: [{ required: false, message: '请输入或选择用户名/ID', trigger: 'blur' }], // Changed from login_username
  email_address: [
    // { required: true, message: '请输入或选择邮箱地址', trigger: 'change' }, // Removed required rule
    { type: 'email', message: '请输入有效的邮箱地址', trigger: ['blur', 'change'] },
  ],
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
    // For edit mode, platform_name and email_address should come from the fetched subscription data
    // Assuming the backend API for fetching a subscription now includes platform_name, email_address, and login_username
    form.value.platform_name = data.platform_name || '';
    form.value.email_address = data.email_address || '';
    // In edit mode, login_username might come as part of the subscription details if it was manually entered.
    // If selected_username_registration_id was used, login_username might be empty.
    // We prioritize displaying the existing login_username if available.
    // The dropdown will be populated based on platform/email if applicable.
    form.value.login_username_input = data.login_username || '';
    form.value.selected_username_registration_id = data.selected_username_registration_id || null;
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
      platform_name: '',
      email_address: '',
      login_username_input: '', // Reset username input
      selected_username_registration_id: null, // Reset selected ID
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
  // Fetch platforms and email accounts for dropdowns
  // Fetching all items for dropdowns, assuming the lists are not excessively large.
  // Consider pagination/search for dropdowns if lists become very long.
  if (platformStore.platforms.length === 0) {
    await platformStore.fetchPlatforms(1, 10000, { orderBy: 'name', sortDirection: 'asc' });
  }
  if (emailAccountStore.emailAccounts.length === 0) {
    await emailAccountStore.fetchEmailAccounts(1, 10000, { orderBy: 'email_address', sortDirection: 'asc' });
  }

  // If in edit mode and initialData is not yet populated (e.g. direct navigation for dev, though not typical for modal)
  // Or if initialData was passed but we want to ensure it's the freshest.
  // For modal usage, initialData from parent is preferred.
  // This explicit fetch might be redundant if ListView always passes fresh `initialData`.
  if (isEditMode.value && props.id && !props.initialData) {
    const subData = await serviceSubscriptionStore.fetchServiceSubscriptionById(props.id);
    if (subData) {
      // Ensure subData contains platform_name and email_address for edit mode
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

const fetchUsernamesForSelect = async () => {
  if (isEditMode.value) {
    // In edit mode, we don't automatically fetch or change the username based on platform/email changes
    // as the initial username is already set. The user can manually change it if needed.
    // However, we can still populate the dropdown for selection if they clear the input.
    // For now, let's keep it simple: only fetch in create mode or if username is empty in edit mode.
    // This behavior might need refinement based on UX preferences.
    if (!form.value.login_username_input && !form.value.selected_username_registration_id) {
       // If username is cleared in edit mode, allow fetching.
    } else {
      // If there's an existing username in edit mode, don't auto-fetch.
      // We could still populate platformUsernames if platform/email changes,
      // to allow user to pick a different one.
    }
  }

  const platformName = form.value.platform_name;
  const emailAddress = form.value.email_address;

  let platformId = null;
  let emailAccountId = null;

  if (platformName) {
    const foundPlatform = platformStore.platforms.find(p => p.name === platformName);
    if (foundPlatform) {
      platformId = foundPlatform.id;
    } else {
      // If platform name is custom, we can't get an ID for filtering registrations yet.
      // Depending on requirements, we might clear usernames or allow manual input only.
      // For now, if no ID, we won't filter by platform.
    }
  }

  if (emailAddress) {
    const foundAccount = emailAccountStore.emailAccounts.find(e => e.email_address === emailAddress);
    if (foundAccount) {
      emailAccountId = foundAccount.id;
    } else {
      // Similar to platform, if email is custom, no ID for filtering.
    }
  }

  // Only fetch if at least one identifier (platformId or emailAccountId) is available,
  // or if we want to fetch all usernames (though less ideal without context).
  // For this feature, it's better to fetch when context (platform/email) is known.
  if (!platformId && !emailAccountId) {
    platformUsernames.value = []; // Clear previous options if context is lost
    // form.value.login_username_input = ''; // Optionally clear username input
    // form.value.selected_username_registration_id = null;
    return;
  }

  loadingUsernames.value = true;
  try {
    // Fetch all matching registrations, not paginated for a dropdown
    await platformRegistrationStore.fetchPlatformRegistrations(1, 10000, {}, {
      platform_id: platformId,
      email_account_id: emailAccountId,
    });
    // The store action updates `platformRegistrationStore.platformRegistrations`
    // We need to adapt this to get the raw list for the select, or add a new store action
    // For now, let's assume fetchPlatformRegistrations can return the list or we access it.
    // A dedicated getter or a modified action in the store would be cleaner.
    // Let's assume the action updates a list that we can then use or it returns the list directly.
    // Given the current store, it updates `this.platformRegistrations`.
    // We'll use a temporary workaround if the action doesn't return the list directly.
    // A better approach: modify store action to return the fetched items for such use cases,
    // or add a new action that doesn't affect the main paginated list in the store.

    // Workaround: use the store's state after fetch.
    // This is okay if fetchPlatformRegistrations is called specifically for this dropdown
    // and doesn't interfere with a main list view using the same store.
    // The current `fetchPlatformRegistrations` updates `this.platformRegistrations` in the store.
    // We will use that.
    
    // If `fetchPlatformRegistrations` returned the data directly:
    // platformUsernames.value = result.data.map(reg => ({ id: reg.id, login_username: reg.login_username }));

    // Using the store's state (assuming it's updated by the call above):
    platformUsernames.value = platformRegistrationStore.platformRegistrations.map(reg => ({
      id: reg.id, // Ensure your platform registration model has 'id'
      login_username: reg.login_username, // And 'login_username'
      email_address: reg.email_address // Assuming reg object has email_address
    }));

  } catch (error) {
    ElMessage.error('获取用户名列表失败');
    platformUsernames.value = [];
  } finally {
    loadingUsernames.value = false;
  }
};

// Watch for changes in platform or email to re-fetch usernames
// Debounce this if it becomes too frequent or performance is an issue
watch(() => form.value.platform_name, (newName, oldName) => {
  if (newName !== oldName) {
    // When platform changes, reset selected username ID and potentially the input
    // if the new platform might not have the old username.
    form.value.selected_username_registration_id = null;
    // form.value.login_username_input = ''; // Optional: clear input
    fetchUsernamesForSelect();
  }
});

watch(() => form.value.email_address, (newEmail, oldEmail) => {
  if (newEmail !== oldEmail) {
    form.value.selected_username_registration_id = null;
    // form.value.login_username_input = ''; // Optional: clear input
    fetchUsernamesForSelect();
  }
});


// Handle selection or creation in the username combobox
const handleUsernameChange = (value) => {
  // 'value' is the item's 'value' prop if selected, or the input string if created.
  // Our el-option value is the registration 'id'.
  if (typeof value === 'number' || (typeof value === 'string' && !isNaN(Number(value)) && platformUsernames.value.some(u => u.id === Number(value)) ) ) {
    // User selected an existing item from the dropdown
    // The 'value' from el-option is the registration_id
    const selectedId = Number(value);
    form.value.selected_username_registration_id = selectedId;
    const selectedUser = platformUsernames.value.find(u => u.id === selectedId);
    if (selectedUser) {
      form.value.login_username_input = selectedUser.login_username; // Update input field to show selected username
    }
  } else if (typeof value === 'string') {
    // User typed a new username (or selected a "created" one if el-select behaves that way)
    // The 'value' is the manually typed string.
    form.value.selected_username_registration_id = null;
    form.value.login_username_input = value; // This should already be set by v-model, but good for clarity
  }
  // If value is undefined (e.g. field cleared), reset both
  if (value === undefined || value === null || value === '') {
    form.value.selected_username_registration_id = null;
    form.value.login_username_input = '';
  }
};


const handleSubmit = async () => {
  if (!formRef.value) return;
  await formRef.value.validate(async (valid) => {
    if (valid) {
      loading.value = true;
      const baseFormData = { ...form.value };
      
      let payload = {};

      // Common fields
      payload.service_name = baseFormData.service_name;
      payload.description = baseFormData.description;
      payload.status = baseFormData.status;
      payload.cost = baseFormData.cost;
      payload.billing_cycle = baseFormData.billing_cycle;
      payload.next_renewal_date = baseFormData.next_renewal_date === '' ? null : baseFormData.next_renewal_date;
      payload.payment_method_notes = baseFormData.payment_method_notes;

      if (!isEditMode.value) {
        // These are only set at creation time for the subscription
        payload.platform_name = baseFormData.platform_name;
        payload.email_address = baseFormData.email_address;
      }

      // Handle username/ID linkage - applies to both create and edit
      // If a username is selected from the dropdown, its ID is stored in selected_username_registration_id.
      // If a username is manually typed, it's stored in login_username_input.
      // The handleUsernameChange function ensures login_username_input reflects the text of the selection,
      // and selected_username_registration_id is set if an item is chosen from the list.
      if (baseFormData.selected_username_registration_id) {
        payload.selected_username_registration_id = baseFormData.selected_username_registration_id;
        // login_username should not be in payload if selected_id is sent
      } else if (baseFormData.login_username_input && baseFormData.login_username_input.trim() !== '') {
        payload.login_username = baseFormData.login_username_input.trim();
        // selected_username_registration_id should not be in payload if login_username is sent
      }
      // If both selected_id and login_username_input are effectively empty/null after form interaction,
      // then neither 'selected_username_registration_id' nor 'login_username' keys will be added to the payload.
      // This fulfills the requirement "应为空或不包含在此次提交中".
      
      // The actual store call will be handled by the parent component
      emit('submit-form', { payload, id: props.id, isEdit: isEditMode.value });
      loading.value = false; // Moved here to ensure it's set after emit
    } else {
      ElMessage.error('请检查表单输入');
      // loading.value = false; // Ensure loading is false on validation error
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