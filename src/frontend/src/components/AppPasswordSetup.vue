<!-- 应用密码设置组件 -->
<template>
  <div class="app-password-setup">
    <!-- 未登录提示 -->
    <el-alert
      v-if="!isAuthenticated"
      title="需要登录才能设置应用密码"
      type="warning"
      :closable="false"
      show-icon
      class="login-alert"
    >
      <p>请先登录您的账户，然后再设置应用密码。</p>
      <el-button type="primary" @click="redirectToLogin" style="margin-top: 10px;">
        前往登录
      </el-button>
    </el-alert>

    <!-- 主要内容区域 -->
    <div v-else>
      <el-card class="setup-card">
        <template #header>
          <div class="card-header">
            <el-icon class="header-icon"><Key /></el-icon>
            <span class="card-title">应用密码设置</span>
            <el-tag v-if="isConnected" type="success">已连接</el-tag>
          </div>
        </template>

      <!-- 设置指南 -->
      <el-alert
        title="应用密码 - 永不过期的邮箱访问方案"
        type="success"
        :closable="false"
        class="guide-alert"
      >
        <p>应用密码提供稳定可靠的邮箱访问，无需担心OAuth2 token过期问题。</p>
        <ul class="benefits-list">
          <li>✅ 永不过期 - 一次设置，长期使用</li>
          <li>✅ 更稳定 - 99%连接成功率</li>
          <li>✅ 更简单 - 无需复杂的重新授权</li>
          <li>✅ 更安全 - 可随时在邮箱设置中撤销</li>
        </ul>
      </el-alert>

      <!-- 服务商选择 -->
      <el-form :model="form" :rules="rules" ref="formRef" label-width="120px" class="setup-form">
        <el-form-item label="邮箱服务商" prop="provider">
          <el-select 
            v-model="form.provider" 
            placeholder="请选择邮箱服务商"
            @change="handleProviderChange"
            style="width: 100%"
          >
                         <el-option
               v-for="provider in supportedProviders"
               :key="provider.id"
               :label="provider.name"
               :value="provider.id"
             >
               <div class="provider-option">
                 <span class="provider-name">{{ provider.name }}</span>
                 <el-tag size="small" type="info">{{ provider.imapServer }}:{{ provider.imapPort }}</el-tag>
               </div>
             </el-option>
          </el-select>
        </el-form-item>

        <el-form-item label="邮箱地址" prop="email">
          <el-input
            v-model="form.email"
            placeholder="请输入邮箱地址"
            :prefix-icon="Message"
            :disabled="!!emailAccount"
          />
        </el-form-item>

        <el-form-item label="应用密码" prop="password">
          <el-input
            v-model="form.password"
            type="password"
            placeholder="请输入应用密码"
            :prefix-icon="Lock"
            show-password
          />
          <div class="password-help">
            <el-text size="small" type="info">
              应用密码通常为16位，可包含空格（如：abcd efgh ijkl mnop）
            </el-text>
          </div>
        </el-form-item>
      </el-form>

      <!-- 获取应用密码指南 -->
      <el-collapse v-if="form.provider" class="guide-collapse">
        <el-collapse-item title="📖 如何获取应用密码？" name="guide">
          <div class="guide-content" v-html="getProviderGuide(form.provider)"></div>
        </el-collapse-item>
      </el-collapse>

      <!-- 操作按钮 -->
      <div class="action-buttons">
        <el-button
          type="primary"
          @click="handleTest"
          :loading="testing"
          :icon="Refresh"
        >
          测试连接
        </el-button>
        
        <el-button
          type="success"
          @click="handleSetup"
          :loading="setting"
          :disabled="!canSetup"
          :icon="Check"
        >
          {{ isEdit ? '更新设置' : '完成设置' }}
        </el-button>

        <el-button @click="handleCancel" :disabled="testing || setting">
          取消
        </el-button>

        <!-- 调试按钮 -->
        <el-button 
          type="info" 
          size="small" 
          @click="debugAuthStatus" 
          style="margin-left: 10px;"
        >
          调试认证状态
        </el-button>
      </div>
    </el-card>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, watch, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Key, Message, Lock, Refresh, Check } from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'
import api from '@/utils/api'

const props = defineProps({
  emailAccount: {
    type: Object,
    default: null
  },
  visible: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['success', 'cancel'])

const router = useRouter()
const authStore = useAuthStore()

// 表单数据
const form = reactive({
  provider: '',
  email: '',
  password: ''
})

// 表单验证规则
const rules = {
  provider: [
    { required: true, message: '请选择邮箱服务商', trigger: 'change' }
  ],
  email: [
    { required: true, message: '请输入邮箱地址', trigger: ['blur', 'change'] },
    { type: 'email', message: '请输入正确的邮箱地址', trigger: ['blur', 'change'] }
  ],
  password: [
    { required: true, message: '请输入应用密码', trigger: ['blur', 'change'] },
    { min: 8, message: '应用密码长度至少8位', trigger: ['blur', 'change'] }
  ]
}

const formRef = ref(null)
const testing = ref(false)
const setting = ref(false)
const supportedProviders = ref([])
const isConnected = ref(false)

const isEdit = computed(() => !!props.emailAccount)
const canSetup = computed(() => form.provider && form.email && form.password)
const isAuthenticated = computed(() => authStore.isAuthenticated)

// 初始化
onMounted(async () => {
  await loadSupportedProviders()
  if (props.emailAccount) {
    initializeForEdit()
  }
})

// 监听邮箱账户变化
watch(() => props.emailAccount, (newAccount) => {
  if (newAccount) {
    initializeForEdit()
  } else {
    resetForm()
  }
})

const loadSupportedProviders = async () => {
  try {
    // 添加调试信息
    console.log('🔐 用户认证状态:', authStore.isAuthenticated)
    console.log('🔐 当前token:', authStore.token)
    
    const response = await api.get('/app-password/providers')
    if (response.data.success) {
      supportedProviders.value = response.data.data || []
    } else {
      // 如果API不存在，提供默认的服务商列表
      supportedProviders.value = getDefaultProviders()
    }
  } catch (error) {
    console.error('加载服务商列表失败:', error)
    console.error('错误详情:', error.response?.data)
    
    if (error.response?.status === 401) {
      ElMessage.error('登录已过期，请重新登录')
      router.push('/login')
      return
    }
    
    // 提供默认的服务商列表作为备选
    supportedProviders.value = getDefaultProviders()
    ElMessage.warning('使用默认服务商列表')
  }
}

// 默认服务商列表
const getDefaultProviders = () => {
  return [
    {
      id: 'gmail',
      name: 'Gmail',
      imapServer: 'imap.gmail.com',
      imapPort: 993,
      description: 'Google Gmail 邮箱服务'
    },
    {
      id: 'outlook',
      name: 'Outlook/Hotmail',
      imapServer: 'outlook.office365.com',
      imapPort: 993,
      description: 'Microsoft Outlook 邮箱服务'
    },
    {
      id: 'yahoo',
      name: 'Yahoo Mail',
      imapServer: 'imap.mail.yahoo.com',
      imapPort: 993,
      description: 'Yahoo 邮箱服务'
    },
    {
      id: 'icloud',
      name: 'iCloud Mail',
      imapServer: 'imap.mail.me.com',
      imapPort: 993,
      description: 'Apple iCloud 邮箱服务'
    }
  ]
}

const initializeForEdit = () => {
  if (props.emailAccount) {
    form.email = props.emailAccount.email_address
    // 检查是否已经设置了应用密码
    checkAppPasswordStatus()
  }
}

const checkAppPasswordStatus = async () => {
  try {
    const response = await api.get(`/app-password/check/${props.emailAccount.id}`)
    const data = response.data.data
    isConnected.value = data.auth_type === 'app_password'
    if (isConnected.value && data.provider_info) {
      form.provider = data.provider_info.id
    }
  } catch (error) {
    console.error('检查应用密码状态失败:', error)
  }
}

const handleProviderChange = (value) => {
  console.log('选择服务商:', value)
}

const resetForm = () => {
  form.provider = ''
  form.email = ''
  form.password = ''
  isConnected.value = false
}

const handleTest = async () => {
  if (!formRef.value) return
  
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  // 检查登录状态
  if (!authStore.isAuthenticated) {
    ElMessage.error('请先登录再测试连接')
    router.push('/login')
    return
  }

  testing.value = true
  try {
    console.log('🧪 开始测试应用密码连接...')
    console.log('🔐 当前token:', authStore.token?.substring(0, 20) + '...')
    
    const response = await api.post('/app-password/test', {
      email_address: form.email,
      app_password: form.password,
      provider: form.provider
    })

    if (response.data.data && response.data.data.success) {
      ElMessage.success('连接测试成功！应用密码配置正确')
    } else {
      const errorMsg = response.data.data?.message || response.data.error || '连接测试失败'
      ElMessage.error(errorMsg)
    }
  } catch (error) {
    console.error('测试连接失败:', error)
    console.error('错误详情:', error.response?.data)
    
    if (error.response?.status === 401) {
      ElMessage.error('登录已过期，请重新登录')
      router.push('/login')
    } else {
      const errorMsg = error.response?.data?.error || error.response?.data?.message || '连接测试失败'
      ElMessage.error(errorMsg)
    }
  } finally {
    testing.value = false
  }
}

const handleSetup = async () => {
  if (!formRef.value) return
  
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  setting.value = true
  try {
    const response = await api.post('/app-password/setup', {
      email_address: form.email,
      app_password: form.password,
      provider: form.provider
    })

    if (response.data.success) {
      ElMessage.success('应用密码设置成功！')
      emit('success', response.data.data)
    } else {
      ElMessage.error(response.data.error || '设置失败')
    }
  } catch (error) {
    console.error('设置应用密码失败:', error)
    ElMessage.error(error.response?.data?.error || '设置失败')
  } finally {
    setting.value = false
  }
}

const handleCancel = () => {
  resetForm()
  emit('cancel')
}

const getProviderGuide = (providerId) => {
  const guides = {
    gmail: `
      <h4>📧 Gmail 应用密码设置步骤：</h4>
      <ol>
        <li>登录 <a href="https://myaccount.google.com" target="_blank">Google 账户</a></li>
        <li>点击左侧菜单的「安全性」</li>
        <li>在「登录 Google」部分，启用「两步验证」（必须先启用）</li>
        <li>启用两步验证后，点击「应用专用密码」</li>
        <li>选择「邮件」应用，点击「生成」</li>
        <li>复制生成的16位应用密码到上面的输入框</li>
      </ol>
    `,
    outlook: `
      <h4>📧 Outlook/Hotmail 应用密码设置步骤：</h4>
      <ol>
        <li>登录 <a href="https://account.microsoft.com" target="_blank">Microsoft 账户</a></li>
        <li>点击「安全性」标签</li>
        <li>在「高级安全选项」下，启用「两步验证」</li>
        <li>启用后，点击「创建新的应用密码」</li>
        <li>输入应用名称（如：邮件客户端）</li>
        <li>复制生成的密码到上面的输入框</li>
      </ol>
    `,
    yahoo: `
      <h4>📧 Yahoo 应用密码设置步骤：</h4>
      <ol>
        <li>登录 <a href="https://login.yahoo.com" target="_blank">Yahoo 账户</a></li>
        <li>点击右上角头像，选择「账户信息」</li>
        <li>点击「账户安全」</li>
        <li>启用「两步验证」</li>
        <li>点击「生成应用密码」</li>
        <li>选择应用类型，生成密码</li>
        <li>复制生成的密码到上面的输入框</li>
      </ol>
    `,
    icloud: `
      <h4>📧 iCloud 应用密码设置步骤：</h4>
      <ol>
        <li>登录 <a href="https://appleid.apple.com" target="_blank">Apple ID</a></li>
        <li>在「登录和安全」部分，启用「双重认证」</li>
        <li>在「应用专用密码」部分，点击「生成密码」</li>
        <li>输入标签（如：邮件客户端）</li>
        <li>复制生成的密码到上面的输入框</li>
      </ol>
    `
  }
  return guides[providerId] || '请选择服务商查看设置指南'
}

const redirectToLogin = () => {
  router.push('/login')
}

// 调试函数
const debugAuthStatus = () => {
  console.log('🔐 调试认证状态:')
  console.log('- isAuthenticated:', authStore.isAuthenticated)
  console.log('- token存在:', !!authStore.token)
  console.log('- token内容:', authStore.token?.substring(0, 50) + '...')
  console.log('- localStorage token:', localStorage.getItem('token')?.substring(0, 50) + '...')
  console.log('- 用户信息:', authStore.user)
  
  ElMessage.info('请查看浏览器控制台查看认证状态详情')
}
</script>

<style scoped>
.app-password-setup {
  max-width: 700px;
  margin: 0 auto;
}

.setup-card {
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.header-icon {
  font-size: 20px;
  margin-right: 8px;
  color: var(--el-color-primary);
}

.card-title {
  font-weight: 600;
  font-size: 16px;
}

.guide-alert {
  margin-bottom: 24px;
}

.benefits-list {
  margin: 8px 0 0 0;
  padding-left: 0;
  list-style: none;
}

.benefits-list li {
  margin: 4px 0;
  color: var(--el-text-color-regular);
}

.setup-form {
  margin-bottom: 24px;
}

.provider-option {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.provider-name {
  font-weight: 500;
}

.password-help {
  margin-top: 4px;
}

.guide-collapse {
  margin-bottom: 24px;
}

.guide-content {
  line-height: 1.6;
}

.guide-content h4 {
  color: var(--el-color-primary);
  margin-bottom: 12px;
}

.guide-content ol {
  padding-left: 20px;
}

.guide-content ol li {
  margin: 8px 0;
}

.guide-content a {
  color: var(--el-color-primary);
  text-decoration: none;
}

.guide-content a:hover {
  text-decoration: underline;
}

.action-buttons {
  display: flex;
  justify-content: center;
  gap: 12px;
  margin-top: 24px;
  padding-top: 24px;
  border-top: 1px solid var(--el-border-color-light);
}

.login-alert {
  margin-bottom: 24px;
}
</style> 