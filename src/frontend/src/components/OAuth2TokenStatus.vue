<template>
  <div class="oauth2-token-status">
    <el-card class="status-card" shadow="hover">
      <template #header>
        <div class="card-header">
          <span class="title">
            <el-icon><Key /></el-icon>
            OAuth2令牌状态
          </span>
          <el-button 
            type="primary" 
            size="small" 
            @click="refreshAllStatus"
            :loading="refreshing"
          >
            <el-icon><Refresh /></el-icon>
            刷新状态
          </el-button>
        </div>
      </template>

      <div v-if="loading" class="loading-container">
        <el-skeleton :rows="3" animated />
      </div>

      <div v-else-if="tokenStatuses.length === 0" class="empty-container">
        <el-empty description="暂无OAuth2邮箱账户" />
      </div>

      <div v-else class="status-list">
        <div 
          v-for="status in tokenStatuses" 
          :key="status.accountId"
          class="status-item"
        >
          <div class="account-info">
            <div class="account-header">
              <span class="email">{{ status.email }}</span>
              <el-tag 
                :type="getStatusTagType(status.status)"
                size="small"
              >
                {{ getStatusText(status.status) }}
              </el-tag>
            </div>
            <div class="provider-info">
              <el-icon><Message /></el-icon>
              <span>{{ getProviderDisplayName(status.provider) }}</span>
              <span class="account-id">ID: {{ status.accountId }}</span>
            </div>
          </div>

          <div class="status-actions">
            <el-button 
              size="small" 
              @click="checkSingleStatus(status.accountId)"
              :loading="status.checking"
            >
              <el-icon><Search /></el-icon>
              检查
            </el-button>
            
            <el-button 
              v-if="status.status === 'expired' || status.status === 'error'"
              type="warning" 
              size="small"
              @click="reauthorize(status.provider, status.accountId)"
            >
              <el-icon><Refresh /></el-icon>
              重新授权
            </el-button>
          </div>
        </div>
      </div>

      <div class="status-summary">
        <el-descriptions :column="3" size="small" border>
          <el-descriptions-item label="总账户数">
            {{ tokenStatuses.length }}
          </el-descriptions-item>
          <el-descriptions-item label="正常账户">
            <el-tag type="success" size="small">
              {{ validCount }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="需要重新授权">
            <el-tag type="danger" size="small">
              {{ expiredCount }}
            </el-tag>
          </el-descriptions-item>
        </el-descriptions>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Key, Refresh, Message, Search } from '@element-plus/icons-vue'
import { emailAccountAPI, checkOAuth2TokenStatus } from '@/utils/api'

const loading = ref(false)
const refreshing = ref(false)
const tokenStatuses = ref([])

// 计算属性
const validCount = computed(() => 
  tokenStatuses.value.filter(s => s.status === 'valid').length
)

const expiredCount = computed(() => 
  tokenStatuses.value.filter(s => s.status === 'expired' || s.status === 'error').length
)

// 获取状态标签类型
const getStatusTagType = (status) => {
  switch (status) {
    case 'valid': return 'success'
    case 'expired': return 'danger'
    case 'error': return 'warning'
    default: return 'info'
  }
}

// 获取状态文本
const getStatusText = (status) => {
  switch (status) {
    case 'valid': return '正常'
    case 'expired': return '已过期'
    case 'error': return '错误'
    case 'checking': return '检查中...'
    default: return '未知'
  }
}

// 获取提供商显示名称
const getProviderDisplayName = (provider) => {
  const names = {
    google: 'Google',
    microsoft: 'Microsoft',
    outlook: 'Outlook'
  }
  return names[provider] || provider
}

// 加载OAuth2邮箱账户
const loadOAuth2Accounts = async () => {
  try {
    loading.value = true
    const response = await emailAccountAPI.getConfigured()
    const accounts = response.data.data || []
    
    // 过滤出OAuth2账户（有provider信息的）
    const oauth2Accounts = accounts.filter(account => 
      account.provider && 
      (account.provider.toLowerCase().includes('google') || 
       account.provider.toLowerCase().includes('microsoft') ||
       account.provider.toLowerCase().includes('outlook'))
    )
    
    tokenStatuses.value = oauth2Accounts.map(account => ({
      accountId: account.id,
      email: account.email_address,
      provider: detectProvider(account.provider),
      status: 'unknown',
      checking: false,
      message: ''
    }))
    
    // 自动检查所有账户的状态
    await checkAllStatuses()
  } catch (error) {
    console.error('加载OAuth2账户失败:', error)
    ElMessage.error('加载OAuth2账户失败')
  } finally {
    loading.value = false
  }
}

// 检测提供商类型
const detectProvider = (providerString) => {
  const lower = providerString.toLowerCase()
  if (lower.includes('google') || lower.includes('gmail')) {
    return 'google'
  } else if (lower.includes('microsoft') || lower.includes('outlook') || lower.includes('hotmail')) {
    return 'microsoft'
  }
  return providerString
}

// 检查所有账户状态
const checkAllStatuses = async () => {
  const promises = tokenStatuses.value.map(status => 
    checkSingleStatus(status.accountId, false)
  )
  await Promise.all(promises)
}

// 检查单个账户状态
const checkSingleStatus = async (accountId, showMessage = true) => {
  const statusItem = tokenStatuses.value.find(s => s.accountId === accountId)
  if (!statusItem) return
  
  try {
    statusItem.checking = true
    statusItem.status = 'checking'
    
    const response = await checkOAuth2TokenStatus(accountId)
    const { status, message } = response.data
    
    statusItem.status = status
    statusItem.message = message
    
    if (showMessage) {
      if (status === 'valid') {
        ElMessage.success(message || '令牌状态正常')
      } else {
        ElMessage.warning(message || '令牌需要处理')
      }
    }
  } catch (error) {
    console.error(`检查账户 ${accountId} 状态失败:`, error)
    statusItem.status = 'error'
    statusItem.message = error.message || '检查失败'
    
    if (showMessage) {
      ElMessage.error('检查令牌状态失败')
    }
  } finally {
    statusItem.checking = false
  }
}

// 刷新所有状态
const refreshAllStatus = async () => {
  refreshing.value = true
  try {
    await checkAllStatuses()
    ElMessage.success('状态刷新完成')
  } catch (error) {
    ElMessage.error('刷新状态失败')
  } finally {
    refreshing.value = false
  }
}

// 重新授权
const reauthorize = async (provider, accountId) => {
  try {
    await ElMessageBox.confirm(
      `确定要重新授权 ${getProviderDisplayName(provider)} 账户吗？`,
      '重新授权确认',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    // 构建OAuth2授权URL
    const baseUrl = process.env.VUE_APP_API_BASE_URL || 'http://localhost:5555/api/v1'
    const authUrl = `${baseUrl}/oauth2/connect/${provider}?account_id=${accountId}`
    
    // 在新窗口中打开授权页面
    const authWindow = window.open(
      authUrl,
      'oauth2_reauth',
      'width=600,height=700,scrollbars=yes,resizable=yes'
    )
    
    // 监听授权完成
    const checkClosed = setInterval(() => {
      if (authWindow.closed) {
        clearInterval(checkClosed)
        ElMessage.info('授权窗口已关闭，正在检查最新状态...')
        // 延迟检查状态，给服务器时间处理
        setTimeout(() => {
          checkSingleStatus(accountId, true)
        }, 2000)
      }
    }, 1000)
  } catch (error) {
    // 用户取消了重新授权
    console.log('用户取消了重新授权')
  }
}

// 组件挂载时加载数据
onMounted(() => {
  loadOAuth2Accounts()
})
</script>

<style scoped>
.oauth2-token-status {
  margin: 20px 0;
}

.status-card {
  max-width: 800px;
  margin: 0 auto;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
}

.loading-container,
.empty-container {
  padding: 20px;
  text-align: center;
}

.status-list {
  margin-bottom: 20px;
}

.status-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  margin-bottom: 12px;
  transition: all 0.3s;
}

.status-item:hover {
  border-color: #409eff;
  box-shadow: 0 2px 8px rgba(64, 158, 255, 0.1);
}

.account-info {
  flex: 1;
}

.account-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 8px;
}

.email {
  font-weight: 600;
  color: #303133;
}

.provider-info {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #606266;
  font-size: 14px;
}

.account-id {
  color: #909399;
  font-size: 12px;
}

.status-actions {
  display: flex;
  gap: 8px;
}

.status-summary {
  margin-top: 20px;
  padding-top: 20px;
  border-top: 1px solid #e4e7ed;
}
</style>
