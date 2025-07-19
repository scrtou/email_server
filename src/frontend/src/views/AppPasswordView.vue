<!-- 应用密码管理页面 -->
<template>
  <div class="app-password-view">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-content">
        <div class="header-info">
          <h1 class="page-title">
            <el-icon class="title-icon"><Key /></el-icon>
            应用密码管理
          </h1>
          <p class="page-description">
            管理您的邮箱应用密码，享受永不过期的稳定邮箱访问体验
          </p>
        </div>
        <div class="header-actions">
          <el-button type="primary" :icon="Plus" @click="showSetupDialog">
            添加应用密码
          </el-button>
        </div>
      </div>
    </div>

    <!-- 统计卡片 -->
    <div class="stats-cards">
      <el-row :gutter="20">
        <el-col :xs="24" :sm="12" :md="6">
          <el-card class="stat-card">
            <div class="stat-content">
              <div class="stat-icon success">
                <el-icon><Check /></el-icon>
              </div>
              <div class="stat-info">
                <div class="stat-value">{{ stats.total }}</div>
                <div class="stat-label">总账户数</div>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :xs="24" :sm="12" :md="6">
          <el-card class="stat-card">
            <div class="stat-content">
              <div class="stat-icon primary">
                <el-icon><Key /></el-icon>
              </div>
              <div class="stat-info">
                <div class="stat-value">{{ stats.appPassword }}</div>
                <div class="stat-label">使用应用密码</div>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :xs="24" :sm="12" :md="6">
          <el-card class="stat-card">
            <div class="stat-content">
              <div class="stat-icon warning">
                <el-icon><Connection /></el-icon>
              </div>
              <div class="stat-info">
                <div class="stat-value">{{ stats.oauth2 }}</div>
                <div class="stat-label">使用OAuth2</div>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :xs="24" :sm="12" :md="6">
          <el-card class="stat-card">
            <div class="stat-content">
              <div class="stat-icon success">
                <el-icon><TrendCharts /></el-icon>
              </div>
              <div class="stat-info">
                <div class="stat-value">99%</div>
                <div class="stat-label">应用密码可靠性</div>
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </div>

    <!-- 功能介绍 -->
    <el-card class="feature-card" v-if="showIntroduction">
      <template #header>
        <div class="card-header">
          <span>🚀 应用密码功能介绍</span>
          <el-button link @click="showIntroduction = false">
            <el-icon><Close /></el-icon>
          </el-button>
        </div>
      </template>
      
      <el-row :gutter="20">
        <el-col :xs="24" :md="12">
          <div class="feature-content">
            <h3>✨ 主要优势</h3>
            <ul class="feature-list">
              <li><strong>永不过期</strong> - 一次设置，长期使用</li>
              <li><strong>高稳定性</strong> - 99%连接成功率</li>
              <li><strong>简化维护</strong> - 无需处理token刷新</li>
              <li><strong>安全可控</strong> - 可随时撤销</li>
            </ul>
          </div>
        </el-col>
        <el-col :xs="24" :md="12">
          <div class="feature-content">
            <h3>📊 性能对比</h3>
            <div class="comparison-table">
              <div class="comparison-row header">
                <span>特性</span>
                <span>OAuth2</span>
                <span>应用密码</span>
              </div>
              <div class="comparison-row">
                <span>过期时间</span>
                <span>1小时</span>
                <span class="highlight">永不过期</span>
              </div>
              <div class="comparison-row">
                <span>稳定性</span>
                <span>85%</span>
                <span class="highlight">99%</span>
              </div>
            </div>
          </div>
        </el-col>
      </el-row>
    </el-card>

    <!-- 邮箱账户列表 -->
    <el-card class="accounts-card">
      <template #header>
        <div class="card-header">
          <span class="card-title">邮箱账户</span>
          <div class="header-actions">
            <el-input
              v-model="searchText"
              placeholder="搜索邮箱地址"
              :prefix-icon="Search"
              clearable
              style="width: 250px;"
              @input="handleSearch"
            />
            <el-button :icon="Refresh" @click="refreshAccounts" :loading="loading">
              刷新
            </el-button>
          </div>
        </div>
      </template>

      <el-table 
        :data="filteredAccounts" 
        v-loading="loading"
        style="width: 100%"
        @sort-change="handleSortChange"
      >
        <el-table-column prop="email_address" label="邮箱地址" min-width="200" sortable="custom">
          <template #default="scope">
            <div class="email-cell">
              <el-icon class="email-icon"><Message /></el-icon>
              {{ scope.row.email_address }}
            </div>
          </template>
        </el-table-column>
        
        <el-table-column label="认证类型" width="150" align="center">
          <template #default="scope">
            <el-tag 
              :type="scope.row.auth_type === 'app_password' ? 'success' : 'warning'"
              :icon="scope.row.auth_type === 'app_password' ? Key : Connection"
            >
              {{ scope.row.auth_type === 'app_password' ? '应用密码' : 'OAuth2' }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column label="连接状态" width="120" align="center">
          <template #default="scope">
            <el-tag 
              :type="scope.row.connection_status === 'connected' ? 'success' : 'danger'"
              size="small"
            >
              {{ scope.row.connection_status === 'connected' ? '已连接' : '未连接' }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column label="服务商" width="120" align="center">
          <template #default="scope">
            <span class="provider-name">{{ getProviderName(scope.row.provider) }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="updated_at" label="最后更新" width="160" sortable="custom">
          <template #default="scope">
            <el-tooltip :content="scope.row.updated_at">
              <span>{{ formatDate(scope.row.updated_at) }}</span>
            </el-tooltip>
          </template>
        </el-table-column>

        <el-table-column label="操作" width="240" fixed="right" align="center">
          <template #default="scope">
            <div class="action-buttons">
              <!-- 应用密码操作 -->
              <template v-if="scope.row.auth_type === 'app_password'">
                <el-button link type="primary" size="small" @click="editAppPassword(scope.row)">
                  <el-icon><Edit /></el-icon>
                  编辑
                </el-button>
                <el-button link type="success" size="small" @click="testConnection(scope.row)">
                  <el-icon><Connection /></el-icon>
                  测试
                </el-button>
              </template>
              
              <!-- OAuth2操作 -->
              <template v-else>
                <el-button link type="warning" size="small" @click="migrateToAppPassword(scope.row)">
                  <el-icon><Switch /></el-icon>
                  迁移
                </el-button>
              </template>
              
              <el-button link type="info" size="small" @click="viewEmails(scope.row)">
                <el-icon><View /></el-icon>
                查看邮件
              </el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-container">
        <el-pagination
          layout="total, sizes, prev, pager, next, jumper"
          :total="pagination.total"
          :current-page="pagination.currentPage"
          :page-size="pagination.pageSize"
          :page-sizes="[10, 20, 50, 100]"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </el-card>

    <!-- 应用密码设置对话框 -->
    <el-dialog
      v-model="setupDialog.visible"
      :title="setupDialog.title"
      width="800px"
      :close-on-click-modal="false"
      @close="handleSetupDialogClose"
    >
      <AppPasswordSetup
        :email-account="setupDialog.emailAccount"
        @success="handleSetupSuccess"
        @cancel="handleSetupCancel"
      />
    </el-dialog>

    <!-- 迁移确认对话框 -->
    <el-dialog
      v-model="migrateDialog.visible"
      title="迁移到应用密码"
      width="600px"
      :close-on-click-modal="false"
    >
      <div class="migrate-content">
        <el-alert
          title="确认迁移"
          type="warning"
          :closable="false"
          class="migrate-alert"
        >
          <p>您即将将以下账户从OAuth2迁移到应用密码认证：</p>
          <p><strong>{{ migrateDialog.emailAccount?.email_address }}</strong></p>
          <p>迁移后，该账户将使用应用密码进行邮箱访问。</p>
        </el-alert>

        <AppPasswordSetup
          :email-account="migrateDialog.emailAccount"
          @success="handleMigrateSuccess"
          @cancel="handleMigrateCancel"
        />
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import {
  Key, Plus, Check, Connection, TrendCharts, Close, Search, Refresh,
  Message, Edit, Switch, View
} from '@element-plus/icons-vue'
import AppPasswordSetup from '@/components/AppPasswordSetup.vue'
import api from '@/utils/api'

const router = useRouter()

// 响应式数据
const loading = ref(false)
const showIntroduction = ref(true)
const searchText = ref('')
const accounts = ref([])
const supportedProviders = ref([])

// 统计数据
const stats = reactive({
  total: 0,
  appPassword: 0,
  oauth2: 0
})

// 分页
const pagination = reactive({
  currentPage: 1,
  pageSize: 20,
  total: 0
})

// 对话框状态
const setupDialog = reactive({
  visible: false,
  title: '',
  emailAccount: null
})

const migrateDialog = reactive({
  visible: false,
  emailAccount: null
})

// 计算属性
const filteredAccounts = computed(() => {
  if (!searchText.value) return accounts.value
  
  const search = searchText.value.toLowerCase()
  return accounts.value.filter(account =>
    account.email_address.toLowerCase().includes(search)
  )
})

// 生命周期
onMounted(() => {
  loadSupportedProviders()
  loadAccounts()
})

// 方法
const loadSupportedProviders = async () => {
  try {
    const response = await api.get('/app-password/providers')
    if (response.data.success) {
      supportedProviders.value = response.data.data || []
    } else {
      supportedProviders.value = getDefaultProviders()
    }
  } catch (error) {
    console.error('加载服务商失败:', error)
    // 提供默认的服务商列表
    supportedProviders.value = getDefaultProviders()
  }
}

// 默认服务商列表
const getDefaultProviders = () => {
  return [
    {
      id: 'gmail',
      name: 'Gmail',
      imapServer: 'imap.gmail.com',
      imapPort: 993
    },
    {
      id: 'outlook',
      name: 'Outlook/Hotmail',
      imapServer: 'outlook.office365.com',
      imapPort: 993
    },
    {
      id: 'yahoo',
      name: 'Yahoo Mail',
      imapServer: 'imap.mail.yahoo.com',
      imapPort: 993
    },
    {
      id: 'icloud',
      name: 'iCloud Mail',
      imapServer: 'imap.mail.me.com',
      imapPort: 993
    }
  ]
}

const loadAccounts = async () => {
  loading.value = true
  try {
    // 加载邮箱账户列表
    const accountsResponse = await api.get('/email-accounts', {
      params: {
        page: pagination.currentPage,
        pageSize: pagination.pageSize
      }
    })
    
    if (accountsResponse.data.success) {
      accounts.value = accountsResponse.data.data || []
      pagination.total = accountsResponse.data.meta?.total_items || 0
    }

    // 检查每个账户的认证类型
    await checkAuthTypes()
    
    // 更新统计信息
    updateStats()
  } catch (error) {
    console.error('加载账户失败:', error)
    ElMessage.error('加载账户信息失败')
  } finally {
    loading.value = false
  }
}

const checkAuthTypes = async () => {
  const promises = accounts.value.map(async (account) => {
    try {
      const response = await api.get(`/app-password/check/${account.id}`)
      const data = response.data.data
      
      account.auth_type = data.auth_type || 'oauth2'
      account.connection_status = data.connected ? 'connected' : 'disconnected'
      account.provider = data.provider_info?.id || ''
    } catch (error) {
      account.auth_type = 'oauth2'
      account.connection_status = 'disconnected'
    }
  })
  
  await Promise.all(promises)
}

const updateStats = () => {
  stats.total = accounts.value.length
  stats.appPassword = accounts.value.filter(acc => acc.auth_type === 'app_password').length
  stats.oauth2 = stats.total - stats.appPassword
}

const refreshAccounts = () => {
  loadAccounts()
}

const handleSearch = () => {
  // 搜索功能已通过 computed 实现
}

const handleSortChange = ({ prop, order }) => {
  // TODO: 实现排序功能
  console.log('排序:', prop, order)
}

const handleSizeChange = (size) => {
  pagination.pageSize = size
  pagination.currentPage = 1
  loadAccounts()
}

const handleCurrentChange = (page) => {
  pagination.currentPage = page
  loadAccounts()
}

const showSetupDialog = () => {
  setupDialog.title = '添加应用密码'
  setupDialog.emailAccount = null
  setupDialog.visible = true
}

const editAppPassword = (account) => {
  setupDialog.title = '编辑应用密码'
  setupDialog.emailAccount = account
  setupDialog.visible = true
}

const migrateToAppPassword = (account) => {
  migrateDialog.emailAccount = account
  migrateDialog.visible = true
}

const testConnection = async (account) => {
  try {
    const response = await api.post('/app-password/test', {
      email_address: account.email_address,
      app_password: account.app_password, 
      provider: account.provider
    })
    
    if (response.data.data && response.data.data.success) {
      ElMessage.success('连接测试成功！')
    } else {
      const errorMsg = response.data.data?.message || response.data.error || '连接测试失败'
      ElMessage.error(errorMsg)
    }
  } catch (error) {
    ElMessage.error('连接测试失败: ' + (error.response?.data?.error || error.message))
  }
}

const viewEmails = (account) => {
  if (account.auth_type === 'app_password') {
    // 使用应用密码获取邮件
    router.push(`/app-password/emails/${account.id}`)
  } else {
    // 使用原有的邮件查看
    router.push(`/inbox?account=${account.id}`)
  }
}

const handleSetupDialogClose = () => {
  setupDialog.visible = false
  setupDialog.emailAccount = null
}

const handleSetupSuccess = () => {
  ElMessage.success('应用密码设置成功！')
  setupDialog.visible = false
  loadAccounts() // 重新加载账户列表
}

const handleSetupCancel = () => {
  setupDialog.visible = false
}

const handleMigrateSuccess = () => {
  ElMessage.success('迁移成功！现在使用应用密码访问邮箱')
  migrateDialog.visible = false
  loadAccounts()
}

const handleMigrateCancel = () => {
  migrateDialog.visible = false
}

const getProviderName = (providerId) => {
  const provider = supportedProviders.value.find(p => p.id === providerId)
  return provider?.name || providerId || '-'
}

const formatDate = (dateStr) => {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return date.toLocaleDateString('zh-CN', { 
    year: 'numeric', 
    month: '2-digit', 
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}
</script>

<style scoped>
.app-password-view {
  padding: 20px;
  min-height: 100vh;
  background: #f5f5f5;
}

.page-header {
  margin-bottom: 20px;
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: white;
  padding: 24px;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.header-info {
  flex: 1;
}

.page-title {
  display: flex;
  align-items: center;
  margin: 0 0 8px 0;
  font-size: 28px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.title-icon {
  margin-right: 12px;
  font-size: 32px;
  color: var(--el-color-primary);
}

.page-description {
  margin: 0;
  color: var(--el-text-color-regular);
  font-size: 16px;
}

.stats-cards {
  margin-bottom: 20px;
}

.stat-card {
  height: 100px;
}

.stat-content {
  display: flex;
  align-items: center;
  height: 100%;
}

.stat-icon {
  width: 48px;
  height: 48px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 16px;
  font-size: 24px;
}

.stat-icon.success {
  background: var(--el-color-success-light-9);
  color: var(--el-color-success);
}

.stat-icon.primary {
  background: var(--el-color-primary-light-9);
  color: var(--el-color-primary);
}

.stat-icon.warning {
  background: var(--el-color-warning-light-9);
  color: var(--el-color-warning);
}

.stat-value {
  font-size: 28px;
  font-weight: bold;
  color: var(--el-text-color-primary);
  line-height: 1;
}

.stat-label {
  font-size: 14px;
  color: var(--el-text-color-regular);
  margin-top: 4px;
}

.feature-card {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-title {
  font-weight: 600;
  font-size: 16px;
}

.feature-content h3 {
  margin: 0 0 12px 0;
  color: var(--el-color-primary);
}

.feature-list {
  margin: 0;
  padding-left: 20px;
}

.feature-list li {
  margin: 8px 0;
  line-height: 1.6;
}

.comparison-table {
  border: 1px solid var(--el-border-color);
  border-radius: 4px;
  overflow: hidden;
}

.comparison-row {
  display: grid;
  grid-template-columns: 1fr 1fr 1fr;
  border-bottom: 1px solid var(--el-border-color);
}

.comparison-row:last-child {
  border-bottom: none;
}

.comparison-row.header {
  background: var(--el-fill-color-light);
  font-weight: bold;
}

.comparison-row > span {
  padding: 12px;
  border-right: 1px solid var(--el-border-color);
  text-align: center;
}

.comparison-row > span:last-child {
  border-right: none;
}

.highlight {
  color: var(--el-color-success);
  font-weight: bold;
}

.accounts-card {
  background: white;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.email-cell {
  display: flex;
  align-items: center;
}

.email-icon {
  margin-right: 8px;
  color: var(--el-color-primary);
}

.provider-name {
  font-weight: 500;
}

.action-buttons {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  justify-content: center;
}

.pagination-container {
  margin-top: 20px;
  display: flex;
  justify-content: center;
}

.migrate-content {
  padding: 20px 0;
}

.migrate-alert {
  margin-bottom: 20px;
}

@media (max-width: 768px) {
  .header-content {
    flex-direction: column;
    text-align: center;
    gap: 16px;
  }
  
  .action-buttons {
    flex-direction: column;
  }
}
</style> 