<template>
  <div class="unified-token-view">
    <div class="page-header">
      <h1>OAuth2令牌管理</h1>
      <p class="description">
        统一管理和监控您的Google、Microsoft等OAuth2邮箱账户的授权状态和令牌健康
      </p>
    </div>

    <!-- 概览统计卡片 -->
    <div class="overview-cards">
      <el-card class="overview-card">
        <div class="stat-item">
          <div class="stat-number">{{ tokenStatus.total || 0 }}</div>
          <div class="stat-label">总令牌数</div>
        </div>
      </el-card>
      <el-card class="overview-card active">
        <div class="stat-item">
          <div class="stat-number">{{ tokenStatus.active || 0 }}</div>
          <div class="stat-label">活跃令牌</div>
        </div>
      </el-card>
      <el-card class="overview-card expired">
        <div class="stat-item">
          <div class="stat-number">{{ tokenStatus.expired || 0 }}</div>
          <div class="stat-label">已过期</div>
        </div>
      </el-card>
      <el-card class="overview-card warning">
        <div class="stat-item">
          <div class="stat-number">{{ tokenStatus.expiring_next_24h || 0 }}</div>
          <div class="stat-label">24小时内过期</div>
        </div>
      </el-card>
      <el-card class="overview-card health" :class="getHealthClass()">
        <div class="stat-item">
          <div class="stat-number">{{ tokenStatus.health_score || 0 }}</div>
          <div class="stat-label">健康评分</div>
        </div>
      </el-card>
    </div>

    <!-- 操作按钮区 -->
    <el-card class="actions-card">
      <div class="actions-header">
        <h3>令牌管理操作</h3>
        <div class="actions-buttons">
          <el-button 
            type="primary" 
            @click="refreshTokens"
            :loading="refreshing"
            icon="Refresh">
            刷新令牌
          </el-button>
          <el-button 
            type="warning" 
            @click="cleanupTokens"
            :loading="cleaning"
            icon="Delete">
            清理重复令牌
          </el-button>
          <el-button 
            @click="loadTokenStatus" 
            icon="Refresh">
            刷新状态
          </el-button>
          <el-button
            type="info"
            @click="showProviderDebugInfo"
            icon="Setting">
            系统调试
          </el-button>
        </div>
      </div>
    </el-card>

    <!-- 系统信息 -->
    <el-card class="system-info-card">
      <template #header>
        <h3>系统监控信息</h3>
      </template>
      <el-descriptions :column="3" border>
        <el-descriptions-item label="最后检查">
          {{ formatTime(tokenStatus.last_check) }}
        </el-descriptions-item>
        <el-descriptions-item label="下次检查">
          {{ tokenStatus.next_check_in || 'N/A' }}
        </el-descriptions-item>
        <el-descriptions-item label="刷新阈值">
          {{ tokenStatus.refresh_threshold_hours ? tokenStatus.refresh_threshold_hours.toFixed(1) + ' 小时' : 'N/A' }}
        </el-descriptions-item>
      </el-descriptions>
    </el-card>

    <!-- 提供商统计 -->
    <el-card class="provider-stats-card" v-if="tokenStatus.provider_stats">
      <template #header>
        <h3>提供商分布统计</h3>
      </template>
      <div class="provider-grid">
        <div 
          v-for="(stats, provider) in tokenStatus.provider_stats" 
          :key="provider" 
          class="provider-card">
          <div class="provider-name">{{ provider.toUpperCase() }}</div>
          <div class="provider-stats">
            <span class="stat">总数: {{ stats.total }}</span>
            <span class="stat active">活跃: {{ stats.active }}</span>
            <span class="stat expired">过期: {{ stats.expired }}</span>
          </div>
        </div>
      </div>
    </el-card>

    <!-- 令牌详细列表 -->
    <el-card class="token-details-card">
      <template #header>
        <h3>OAuth2令牌详细状态</h3>
      </template>
      
      <div v-if="loading" class="loading-container">
        <el-skeleton :rows="3" animated />
      </div>

      <div v-else-if="!tokenStatus.token_details || tokenStatus.token_details.length === 0" class="empty-container">
        <el-empty description="暂无OAuth2邮箱账户令牌" />
      </div>

      <div v-else>
        <!-- 简化列表视图 -->
        <div class="token-list-view" v-if="viewMode === 'list'">
          <div 
            v-for="token in tokenStatus.token_details" 
            :key="token.account_id"
            class="token-item"
          >
            <div class="token-info">
              <div class="token-header">
                <span class="email">{{ token.email }}</span>
                <el-tag 
                  :type="getStatusTagType(token.status)"
                  size="small"
                >
                  {{ getStatusText(token.status) }}
                </el-tag>
              </div>
              <div class="provider-info">
                <el-icon><Message /></el-icon>
                <span>{{ getProviderDisplayName(token.provider) }}</span>
                <span class="account-id">ID: {{ token.account_id }}</span>
                <span class="expiry-info">{{ formatTimeToExpiry(token.time_to_expiry) }}</span>
              </div>
            </div>

            <div class="token-actions">
              <el-button 
                size="small" 
                @click="checkSingleToken(token.account_id)"
                :loading="checkingTokens[token.account_id]"
              >
                <el-icon><Search /></el-icon>
                检查
              </el-button>
              
              <el-button 
                v-if="token.status === 'expired' || token.status === 'invalid'"
                type="warning" 
                size="small"
                @click="reauthorize(token.provider, token.account_id)"
              >
                <el-icon><Refresh /></el-icon>
                重新授权
              </el-button>
            </div>
          </div>
        </div>

        <!-- 详细表格视图 -->
        <el-table 
          v-else
          :data="tokenStatus.token_details || []" 
          style="width: 100%"
          :default-sort="{ prop: 'expiry', order: 'ascending' }">
          
          <el-table-column prop="email" label="邮箱账户" min-width="200">
            <template #default="{ row }">
              <div class="email-cell">
                <el-tag :type="getProviderTagType(row.provider)" size="small">
                  {{ row.provider.toUpperCase() }}
                </el-tag>
                <span class="email">{{ row.email }}</span>
              </div>
            </template>
          </el-table-column>

          <el-table-column prop="status" label="状态" width="120">
            <template #default="{ row }">
              <el-tag :type="getStatusTagType(row.status)" size="small">
                {{ getStatusText(row.status) }}
              </el-tag>
            </template>
          </el-table-column>

          <el-table-column prop="created_at" label="接入时间" width="160">
            <template #default="{ row }">
              <div class="time-cell">
                <div>{{ formatDate(row.created_at) }}</div>
                <small class="age">{{ row.age_hours }}小时前</small>
              </div>
            </template>
          </el-table-column>

          <el-table-column prop="updated_at" label="最后刷新" width="160">
            <template #default="{ row }">
              <div class="time-cell">
                <div>{{ formatDate(row.updated_at) }}</div>
                <small class="age">{{ row.last_refresh_hours }}小时前</small>
              </div>
            </template>
          </el-table-column>

          <el-table-column prop="expiry" label="过期时间" width="180">
            <template #default="{ row }">
              <div class="expiry-cell">
                <div>{{ formatDate(row.expiry) }}</div>
                <small :class="getExpiryClass(row.time_to_expiry)">
                  {{ formatTimeToExpiry(row.time_to_expiry) }}
                </small>
              </div>
            </template>
          </el-table-column>

          <el-table-column label="操作" width="150">
            <template #default="{ row }">
              <el-button-group>
                <el-button 
                  size="small" 
                  @click="checkSingleToken(row.account_id)"
                  :loading="checkingTokens[row.account_id]"
                >
                  检查
                </el-button>
                <el-button 
                  v-if="row.status === 'expired' || row.status === 'invalid'"
                  type="warning" 
                  size="small"
                  @click="reauthorize(row.provider, row.account_id)"
                >
                  重新授权
                </el-button>
              </el-button-group>
            </template>
          </el-table-column>
        </el-table>

        <!-- 视图切换 -->
        <div class="view-controls">
          <el-radio-group v-model="viewMode" size="small">
            <el-radio-button label="list">列表视图</el-radio-button>
            <el-radio-button label="table">表格视图</el-radio-button>
          </el-radio-group>
        </div>
      </div>
    </el-card>

    <!-- 帮助信息 -->
    <el-card class="help-card" shadow="never">
      <template #header>
        <div class="help-header">
          <el-icon><QuestionFilled /></el-icon>
          <span>帮助信息</span>
        </div>
      </template>

      <el-collapse>
        <el-collapse-item title="什么是OAuth2令牌？" name="1">
          <p>OAuth2令牌是用于访问第三方服务（如Gmail、Outlook）的授权凭证。包含：</p>
          <ul>
            <li><strong>访问令牌</strong>：用于实际API调用，通常1小时过期</li>
            <li><strong>刷新令牌</strong>：用于获取新的访问令牌，可长期有效</li>
          </ul>
        </el-collapse-item>

        <el-collapse-item title="令牌状态说明" name="2">
          <ul>
            <li><el-tag type="success" size="small">正常</el-tag> - 令牌有效，可正常使用</li>
            <li><el-tag type="warning" size="small">即将过期</el-tag> - 令牌将在1小时内过期</li>
            <li><el-tag type="danger" size="small">已过期</el-tag> - 令牌已过期，需要刷新</li>
            <li><el-tag type="info" size="small">未知</el-tag> - 令牌状态未知，需要检查</li>
          </ul>
        </el-collapse-item>

        <el-collapse-item title="如何处理令牌问题？" name="3">
          <p>系统提供多种处理方式：</p>
          <ul>
            <li><strong>自动刷新</strong>：系统会自动刷新即将过期的令牌</li>
            <li><strong>手动刷新</strong>：点击"刷新令牌"按钮手动触发刷新</li>
            <li><strong>重新授权</strong>：当令牌无法刷新时，点击"重新授权"重新获取授权</li>
            <li><strong>清理重复</strong>：清理重复和无效的令牌记录</li>
          </ul>
        </el-collapse-item>
      </el-collapse>
    </el-card>
  </div>
</template>

<script>
import { ref, onMounted, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Message, Search, QuestionFilled } from '@element-plus/icons-vue'
import { emailAccountAPI, oauth2API } from '@/utils/api'
import api from '@/utils/api'

export default {
  name: 'UnifiedTokenView',
  components: {
    Refresh, Message, Search, QuestionFilled
  },
  setup() {
    const tokenStatus = ref({})
    const loading = ref(false)
    const refreshing = ref(false)
    const cleaning = ref(false)
    const checkingTokens = ref({})
    const viewMode = ref('table')
    let refreshTimer = null

    // 加载令牌状态
    const loadTokenStatus = async () => {
      try {
        loading.value = true
        const response = await api.get('/tokens/status')
        console.log('令牌状态API响应:', response)
        
        if (response && response.status === 'success' && response.data) {
          tokenStatus.value = response.data
          console.log('设置的令牌状态数据:', tokenStatus.value)
        } else {
          console.error('API响应状态异常:', response)
          ElMessage.error('获取令牌状态失败：响应格式错误')
        }
      } catch (error) {
        console.error('获取令牌状态失败:', error)
        ElMessage.error('获取令牌状态失败')
      } finally {
        loading.value = false
      }
    }

    // 刷新令牌
    const refreshTokens = async () => {
      refreshing.value = true
      try {
        await api.post('/tokens/refresh')
        ElMessage.success('令牌刷新任务已启动')
        setTimeout(loadTokenStatus, 2000)
      } catch (error) {
        console.error('刷新令牌失败:', error)
        ElMessage.error('刷新令牌失败')
      } finally {
        refreshing.value = false
      }
    }

    // 清理令牌
    const cleanupTokens = async () => {
      cleaning.value = true
      try {
        const response = await api.post('/tokens/cleanup')
        const result = response.data
        
        ElMessage.success(`清理完成！删除了${result.total_tokens_deleted}个令牌记录`)
        
        let details = []
        if (result.duplicates_deleted > 0) {
          details.push(`重复邮箱令牌: ${result.duplicates_deleted}个`)
        }
        if (result.test_records_deleted > 0) {
          details.push(`测试记录: ${result.test_records_deleted}个`)
        }
        if (result.expired_tokens_deleted > 0) {
          details.push(`过期令牌: ${result.expired_tokens_deleted}个`)
        }
        if (result.orphan_tokens_deleted > 0) {
          details.push(`孤儿令牌: ${result.orphan_tokens_deleted}个`)
        }
        
        if (details.length > 0) {
          ElMessage.info(`清理详情: ${details.join(', ')}`)
        }
        
        setTimeout(loadTokenStatus, 1000)
      } catch (error) {
        console.error('清理令牌失败:', error)
        ElMessage.error('清理令牌失败')
      } finally {
        cleaning.value = false
      }
    }

    // 检查单个令牌
    const checkSingleToken = async (accountId) => {
      checkingTokens.value[accountId] = true
      try {
        // 重新加载状态
        await loadTokenStatus()
        ElMessage.success('令牌状态检查完成')
      } catch (error) {
        console.error(`检查账户 ${accountId} 状态失败:`, error)
        ElMessage.error('检查令牌状态失败')
      } finally {
        checkingTokens.value[accountId] = false
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

        const response = await oauth2API.getConnectURL(provider, accountId)
        console.log('OAuth2 API响应:', response)

        if (response && response.auth_url) {
          window.location.href = response.auth_url
        } else {
          ElMessage.error('无法获取重新授权链接，请稍后重试。')
        }
      } catch (error) {
        console.log('用户取消了重新授权')
      }
    }

    // 显示提供商调试信息
    const showProviderDebugInfo = async () => {
      try {
        const allAccountsResponse = await emailAccountAPI.getAll()
        const allAccounts = allAccountsResponse.data || []

        const configuredResponse = await emailAccountAPI.getConfigured()
        const configuredAccounts = configuredResponse.data || []

        const providersResponse = await oauth2API.getProviders()
        const providers = providersResponse.data.providers || []

        let debugInfo = `=== 令牌管理调试信息 ===\n\n`
        debugInfo += `所有邮箱账户数量: ${allAccounts.length}\n`
        debugInfo += `已配置邮箱账户数量: ${configuredAccounts.length}\n`
        debugInfo += `OAuth2账户数量: ${configuredAccounts.filter(acc => acc.is_oauth_connected).length}\n`
        debugInfo += `当前令牌记录数量: ${tokenStatus.value.total || 0}\n\n`

        debugInfo += `=== 令牌健康状态 ===\n`
        debugInfo += `活跃令牌: ${tokenStatus.value.active || 0}\n`
        debugInfo += `过期令牌: ${tokenStatus.value.expired || 0}\n`
        debugInfo += `健康评分: ${tokenStatus.value.health_score || 0}\n\n`

        debugInfo += `=== OAuth2提供商配置 (${providers.length}个) ===\n`
        providers.forEach((provider, index) => {
          debugInfo += `${index + 1}. ${provider.name}\n`
          debugInfo += `   Client ID: ${provider.client_id}\n`
          debugInfo += `   Auth URL: ${provider.auth_url}\n`
          debugInfo += `   Has Secret: ${provider.has_secret ? '是' : '否'}\n\n`
        })

        await ElMessageBox.alert(debugInfo, '系统调试信息', {
          confirmButtonText: '确定',
          type: 'info',
          customStyle: {
            width: '700px',
            maxHeight: '80vh',
            overflow: 'auto'
          }
        })
      } catch (error) {
        console.error('获取调试信息失败:', error)
        ElMessage.error('获取调试信息失败')
      }
    }

    // 格式化时间
    const formatTime = (timeStr) => {
      if (!timeStr) return 'N/A'
      return new Date(timeStr).toLocaleString('zh-CN')
    }

    const formatDate = (dateStr) => {
      if (!dateStr) return 'N/A'
      return new Date(dateStr).toLocaleString('zh-CN', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit'
      })
    }

    const formatTimeToExpiry = (timeStr) => {
      if (!timeStr) return 'N/A'
      
      const match = timeStr.match(/(-?)(?:(\d+)h)?(?:(\d+)m)?(?:(\d+(?:\.\d+)?)s)?/)
      if (!match) return timeStr
      
      const [, negative, hours, minutes] = match
      const isNegative = negative === '-'
      
      if (isNegative) {
        return '已过期'
      }
      
      const h = parseInt(hours) || 0
      const m = parseInt(minutes) || 0
      
      if (h > 24) {
        const days = Math.floor(h / 24)
        const remainingHours = h % 24
        return `${days}天${remainingHours}小时后过期`
      } else if (h > 0) {
        return `${h}小时${m}分钟后过期`
      } else if (m > 0) {
        return `${m}分钟后过期`
      } else {
        return '即将过期'
      }
    }

    // 获取健康评分样式
    const getHealthClass = () => {
      const score = tokenStatus.value.health_score || 0
      if (score >= 80) return 'excellent'
      if (score >= 60) return 'good'
      if (score >= 40) return 'warning'
      return 'danger'
    }

    // 获取状态文本
    const getStatusText = (status) => {
      const statusMap = {
        'active': '正常',
        'expired': '已过期',
        'expiring_soon': '即将过期',
        'expiring_today': '今日过期',
        'invalid': '需重新授权',
        'valid': '正常',
        'error': '错误'
      }
      return statusMap[status] || status
    }

    // 获取状态标签类型
    const getStatusTagType = (status) => {
      const typeMap = {
        'active': 'success',
        'expired': 'danger',
        'expiring_soon': 'warning',
        'expiring_today': 'warning',
        'invalid': 'danger',
        'valid': 'success',
        'error': 'danger'
      }
      return typeMap[status] || 'info'
    }

    // 获取提供商标签类型
    const getProviderTagType = (provider) => {
      const typeMap = {
        'google': 'primary',
        'microsoft': 'success'
      }
      return typeMap[provider] || 'info'
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

    // 获取过期状态样式
    const getExpiryClass = (timeStr) => {
      if (!timeStr || timeStr.includes('-')) return 'expired'
      
      const match = timeStr.match(/(?:(\d+)h)?(?:(\d+)m)?/)
      if (!match) return ''
      
      const hours = parseInt(match[1]) || 0
      const minutes = parseInt(match[2]) || 0
      const totalMinutes = hours * 60 + minutes
      
      if (totalMinutes <= 60) return 'expiring-soon'
      if (totalMinutes <= 360) return 'expiring-today'
      return 'normal'
    }

    onMounted(() => {
      loadTokenStatus()
      refreshTimer = setInterval(loadTokenStatus, 30000)
    })

    onUnmounted(() => {
      if (refreshTimer) {
        clearInterval(refreshTimer)
      }
    })

    return {
      tokenStatus,
      loading,
      refreshing,
      cleaning,
      checkingTokens,
      viewMode,
      loadTokenStatus,
      refreshTokens,
      cleanupTokens,
      checkSingleToken,
      reauthorize,
      showProviderDebugInfo,
      formatTime,
      formatDate,
      formatTimeToExpiry,
      getHealthClass,
      getStatusText,
      getStatusTagType,
      getProviderTagType,
      getProviderDisplayName,
      getExpiryClass
    }
  }
}
</script>

<style scoped>
.unified-token-view {
  padding: 20px;
  max-width: 1400px;
  margin: 0 auto;
}

.page-header {
  margin-bottom: 24px;
  text-align: center;
}

.page-header h1 {
  margin: 0 0 8px 0;
  color: #303133;
  font-size: 28px;
}

.description {
  color: #606266;
  font-size: 14px;
  margin: 0;
}

/* 概览卡片 */
.overview-cards {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 16px;
  margin-bottom: 24px;
}

.overview-card {
  text-align: center;
  cursor: pointer;
  transition: all 0.3s;
}

.overview-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.stat-item {
  padding: 20px;
}

.stat-number {
  font-size: 32px;
  font-weight: bold;
  margin-bottom: 8px;
  color: #303133;
}

.stat-label {
  font-size: 14px;
  color: #909399;
}

.overview-card.active {
  background: linear-gradient(135deg, #409eff 0%, #66b3ff 100%);
  color: white;
}

.overview-card.active .stat-number,
.overview-card.active .stat-label {
  color: white;
}

.overview-card.expired {
  background: linear-gradient(135deg, #f56c6c 0%, #ff8a8a 100%);
  color: white;
}

.overview-card.expired .stat-number,
.overview-card.expired .stat-label {
  color: white;
}

.overview-card.warning {
  background: linear-gradient(135deg, #e6a23c 0%, #f2c55c 100%);
  color: white;
}

.overview-card.warning .stat-number,
.overview-card.warning .stat-label {
  color: white;
}

.overview-card.health.excellent {
  background: linear-gradient(135deg, #67c23a 0%, #85ce5d 100%);
  color: white;
}

.overview-card.health.good {
  background: linear-gradient(135deg, #409eff 0%, #66b3ff 100%);
  color: white;
}

.overview-card.health.warning {
  background: linear-gradient(135deg, #e6a23c 0%, #f2c55c 100%);
  color: white;
}

.overview-card.health.danger {
  background: linear-gradient(135deg, #f56c6c 0%, #ff8a8a 100%);
  color: white;
}

.overview-card.health .stat-number,
.overview-card.health .stat-label {
  color: white;
}

/* 操作按钮区 */
.actions-card {
  margin-bottom: 24px;
}

.actions-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 16px;
}

.actions-header h3 {
  margin: 0;
  color: #303133;
}

.actions-buttons {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

/* 系统信息卡片 */
.system-info-card {
  margin-bottom: 24px;
}

/* 提供商统计 */
.provider-stats-card {
  margin-bottom: 24px;
}

.provider-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 16px;
}

.provider-card {
  padding: 16px;
  border-radius: 8px;
  border: 1px solid #e4e7ed;
  background: #fafafa;
  transition: all 0.3s;
}

.provider-card:hover {
  border-color: #409eff;
  background: #f0f9ff;
}

.provider-name {
  font-weight: bold;
  margin-bottom: 12px;
  color: #409eff;
  font-size: 16px;
}

.provider-stats {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.provider-stats .stat {
  font-size: 14px;
  color: #606266;
}

.provider-stats .stat.active {
  color: #67c23a;
  font-weight: 600;
}

.provider-stats .stat.expired {
  color: #f56c6c;
  font-weight: 600;
}

/* 令牌详细列表 */
.token-details-card {
  margin-bottom: 24px;
}

.loading-container,
.empty-container {
  padding: 40px;
  text-align: center;
}

/* 列表视图 */
.token-list-view {
  margin-bottom: 20px;
}

.token-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  margin-bottom: 12px;
  transition: all 0.3s;
  background: #ffffff;
}

.token-item:hover {
  border-color: #409eff;
  box-shadow: 0 2px 8px rgba(64, 158, 255, 0.1);
  transform: translateY(-1px);
}

.token-info {
  flex: 1;
}

.token-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 8px;
}

.email {
  font-weight: 600;
  color: #303133;
  font-family: 'SF Mono', Monaco, monospace;
}

.provider-info {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #909399;
  font-size: 13px;
}

.account-id {
  color: #c0c4cc;
  font-size: 12px;
}

.expiry-info {
  color: #e6a23c;
  font-size: 12px;
  font-weight: 500;
}

.token-actions {
  display: flex;
  gap: 8px;
}

/* 表格视图 */
.email-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}

.time-cell {
  display: flex;
  flex-direction: column;
}

.time-cell .age {
  color: #c0c4cc;
  font-size: 12px;
}

.expiry-cell {
  display: flex;
  flex-direction: column;
}

.expiry-cell .expired {
  color: #f56c6c;
  font-weight: bold;
}

.expiry-cell .expiring-soon {
  color: #e6a23c;
  font-weight: bold;
}

.expiry-cell .expiring-today {
  color: #fa8c16;
}

.expiry-cell .normal {
  color: #67c23a;
}

/* 视图控制 */
.view-controls {
  margin-top: 16px;
  text-align: center;
  padding-top: 16px;
  border-top: 1px solid #e4e7ed;
}

/* 帮助信息 */
.help-card {
  margin-bottom: 24px;
}

.help-header {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #409eff;
  font-weight: 600;
}

.help-card ul {
  margin: 8px 0;
  padding-left: 20px;
}

.help-card li {
  margin: 4px 0;
  color: #606266;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .unified-token-view {
    padding: 16px;
  }
  
  .overview-cards {
    grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
    gap: 12px;
  }
  
  .actions-header {
    flex-direction: column;
    align-items: stretch;
  }
  
  .actions-buttons {
    justify-content: center;
  }
  
  .token-item {
    flex-direction: column;
    align-items: stretch;
    gap: 12px;
  }
  
  .token-actions {
    justify-content: center;
  }
}
</style>