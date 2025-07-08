<template>
  <div class="token-status-view">
    <el-card class="overview-card">
      <template #header>
        <div class="card-header">
          <h3>OAuth2 令牌状态监控</h3>
          <div class="actions">
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
            <el-button @click="loadTokenStatus" icon="Refresh">刷新状态</el-button>
          </div>
        </div>
      </template>

      <!-- 统计概览 -->
      <div class="stats-grid">
        <div class="stat-card">
          <div class="stat-number">{{ tokenStatus.total || 0 }}</div>
          <div class="stat-label">总令牌数</div>
        </div>
        <div class="stat-card active">
          <div class="stat-number">{{ tokenStatus.active || 0 }}</div>
          <div class="stat-label">活跃令牌</div>
        </div>
        <div class="stat-card expired">
          <div class="stat-number">{{ tokenStatus.expired || 0 }}</div>
          <div class="stat-label">已过期</div>
        </div>
        <div class="stat-card warning">
          <div class="stat-number">{{ tokenStatus.expiring_next_24h || 0 }}</div>
          <div class="stat-label">24小时内过期</div>
        </div>
        <div class="stat-card health" :class="getHealthClass()">
          <div class="stat-number">{{ tokenStatus.health_score || 0 }}</div>
          <div class="stat-label">健康评分</div>
        </div>
      </div>

      <!-- 系统信息 -->
      <div class="system-info">
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
      </div>
    </el-card>

    <!-- 提供商统计 -->
    <el-card class="provider-stats-card">
      <template #header>
        <h3>提供商统计</h3>
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

    <!-- 详细令牌列表 -->
    <el-card class="token-details-card">
      <template #header>
        <h3>令牌详细信息</h3>
      </template>
      
      <el-table 
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

        <el-table-column label="操作" width="100">
          <template #default="{ row }">
            <el-button 
              v-if="row.status === 'expired' || row.status === 'invalid'"
              type="warning" 
              size="small"
              @click="handleReauthorize(row)">
              重新授权
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script>
import { ref, onMounted, onUnmounted } from 'vue'
import { ElMessage } from 'element-plus'
import api from '@/utils/api'

export default {
  name: 'TokenStatusView',
  setup() {
    const tokenStatus = ref({})
    const refreshing = ref(false)
    const cleaning = ref(false)
    let refreshTimer = null

    const loadTokenStatus = async () => {
      try {
        const response = await api.get('/tokens/status')
        console.log('令牌状态API响应:', response)
        // 由于响应拦截器的处理，response现在是 { status: "success", data: {...} }
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
      }
    }

    const refreshTokens = async () => {
      refreshing.value = true
      try {
        await api.post('/tokens/refresh')
        ElMessage.success('令牌刷新任务已启动')
        // 等待2秒后刷新状态
        setTimeout(loadTokenStatus, 2000)
      } catch (error) {
        console.error('刷新令牌失败:', error)
        ElMessage.error('刷新令牌失败')
      } finally {
        refreshing.value = false
      }
    }

    const cleanupTokens = async () => {
      cleaning.value = true
      try {
        const response = await api.post('/tokens/cleanup')
        const result = response.data
        
        ElMessage.success(`清理完成！删除了${result.total_tokens_deleted}个令牌记录`)
        
        // 显示详细结果
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
        
        // 显示发现的问题统计
        let stats = []
        if (result.duplicate_email_groups > 0) {
          stats.push(`${result.duplicate_email_groups}组重复邮箱`)
        }
        if (result.test_records_found > 0) {
          stats.push(`${result.test_records_found}条测试记录`)
        }
        
        if (stats.length > 0) {
          ElMessage.info(`发现问题: ${stats.join(', ')}`)
        }
        
        // 刷新状态
        setTimeout(loadTokenStatus, 1000)
      } catch (error) {
        console.error('清理令牌失败:', error)
        ElMessage.error('清理令牌失败')
      } finally {
        cleaning.value = false
      }
    }

    const handleReauthorize = (token) => {
      // 跳转到重新授权页面
      ElMessage.info(`请重新授权 ${token.email} 的 ${token.provider} 账户`)
      // 这里可以添加具体的重新授权逻辑
    }

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
      
      // 解析时间字符串 (如 "2h30m15s")
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

    const getHealthClass = () => {
      const score = tokenStatus.value.health_score || 0
      if (score >= 80) return 'excellent'
      if (score >= 60) return 'good'
      if (score >= 40) return 'warning'
      return 'danger'
    }

    const getStatusText = (status) => {
      const statusMap = {
        'active': '正常',
        'expired': '已过期',
        'expiring_soon': '即将过期',
        'expiring_today': '今日过期',
        'invalid': '需重新授权'
      }
      return statusMap[status] || status
    }

    const getStatusTagType = (status) => {
      const typeMap = {
        'active': 'success',
        'expired': 'danger',
        'expiring_soon': 'warning',
        'expiring_today': 'warning',
        'invalid': 'danger'
      }
      return typeMap[status] || 'info'
    }

    const getProviderTagType = (provider) => {
      const typeMap = {
        'google': 'primary',
        'microsoft': 'success'
      }
      return typeMap[provider] || 'info'
    }

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
      // 每30秒自动刷新一次状态
      refreshTimer = setInterval(loadTokenStatus, 30000)
    })

    onUnmounted(() => {
      if (refreshTimer) {
        clearInterval(refreshTimer)
      }
    })

    return {
      tokenStatus,
      refreshing,
      cleaning,
      loadTokenStatus,
      refreshTokens,
      cleanupTokens,
      handleReauthorize,
      formatTime,
      formatDate,
      formatTimeToExpiry,
      getHealthClass,
      getStatusText,
      getStatusTagType,
      getProviderTagType,
      getExpiryClass
    }
  }
}
</script>

<style scoped>
.token-status-view {
  padding: 20px;
  max-width: 1400px;
  margin: 0 auto;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-header h3 {
  margin: 0;
}

.overview-card {
  margin-bottom: 20px;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  gap: 20px;
  margin-bottom: 20px;
}

.stat-card {
  text-align: center;
  padding: 20px;
  border-radius: 8px;
  border: 1px solid #e1e1e1;
  background: #fafafa;
}

.stat-card.active {
  background: #f0f9ff;
  border-color: #1890ff;
}

.stat-card.expired {
  background: #fff1f0;
  border-color: #ff4d4f;
}

.stat-card.warning {
  background: #fffbe6;
  border-color: #faad14;
}

.stat-card.health.excellent {
  background: #f6ffed;
  border-color: #52c41a;
}

.stat-card.health.good {
  background: #f0f9ff;
  border-color: #1890ff;
}

.stat-card.health.warning {
  background: #fffbe6;
  border-color: #faad14;
}

.stat-card.health.danger {
  background: #fff1f0;
  border-color: #ff4d4f;
}

.stat-number {
  font-size: 24px;
  font-weight: bold;
  margin-bottom: 5px;
}

.stat-label {
  font-size: 14px;
  color: #666;
}

.system-info {
  margin-top: 20px;
}

.provider-stats-card {
  margin-bottom: 20px;
}

.provider-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 15px;
}

.provider-card {
  padding: 15px;
  border-radius: 8px;
  border: 1px solid #e1e1e1;
  background: #fafafa;
}

.provider-name {
  font-weight: bold;
  margin-bottom: 10px;
  color: #1890ff;
}

.provider-stats {
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.provider-stats .stat {
  font-size: 14px;
}

.provider-stats .stat.active {
  color: #52c41a;
}

.provider-stats .stat.expired {
  color: #ff4d4f;
}

.email-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}

.email {
  font-family: monospace;
}

.time-cell {
  display: flex;
  flex-direction: column;
}

.time-cell .age {
  color: #999;
  font-size: 12px;
}

.expiry-cell {
  display: flex;
  flex-direction: column;
}

.expiry-cell .expired {
  color: #ff4d4f;
  font-weight: bold;
}

.expiry-cell .expiring-soon {
  color: #faad14;
  font-weight: bold;
}

.expiry-cell .expiring-today {
  color: #fa8c16;
}

.expiry-cell .normal {
  color: #52c41a;
}

.token-details-card {
  margin-bottom: 20px;
}
</style>