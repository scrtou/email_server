<template>
  <div class="dashboard">
    
    <!-- 加载状态 -->
    <div v-if="loading" class="loading-container">
      <el-loading text="加载中..." />
    </div>
    
    <!-- 统计卡片 -->
    <el-row v-else :gutter="20" style="margin-bottom: 20px">
      <el-col :span="8">
        <el-card class="stat-card">
          <div class="stat-item">
            <div class="stat-number">{{ dashboardData.email_account_count || 0 }}</div>
            <div class="stat-label">邮箱总数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card class="stat-card">
          <div class="stat-item">
            <div class="stat-number">{{ dashboardData.platform_count || 0 }}</div>
            <div class="stat-label">服务总数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card class="stat-card">
          <div class="stat-item">
            <div class="stat-number">{{ dashboardData.relation_count || 0 }}</div>
            <div class="stat-label">关联总数</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 最近添加的邮箱和服务 -->
    <el-row :gutter="20">
      <el-col :span="12">
        <el-card class="dashboard-list-card">
          <template #header>
            <div class="card-header">
              <span>最近添加的邮箱</span>
            </div>
          </template>
          
          <el-table
            :data="dashboardData.recent_email_accounts || []"
            style="width: 100%"
            v-loading="loading"
          >
            <el-table-column prop="email_address" label="邮箱" />
            <el-table-column prop="platform_count" label="服务数量" /> <!-- Assuming service_count is a field in EmailAccount model -->
            <el-table-column prop="created_at" label="创建时间">
              <template #default="scope">
                {{ formatDate(scope.row.created_at) }}
              </template>
            </el-table-column>
          </el-table>
          
          <div v-if="!loading && (!dashboardData.recent_email_accounts || dashboardData.recent_email_accounts.length === 0)"
               class="no-data">
            <el-empty description="暂无数据" />
          </div>
        </el-card>
      </el-col>
      
      <el-col :span="12">
        <el-card class="dashboard-list-card">
          <template #header>
            <div class="card-header">
              <span>最近添加的服务</span>
            </div>
          </template>
          
          <el-table
            :data="dashboardData.recent_platforms || []"
            style="width: 100%"
            v-loading="loading"
          >
            <el-table-column prop="name" label="服务名称" />
            <el-table-column prop="email_account_count" label="邮箱数量" /> <!-- Assuming email_count is a field in Platform model -->
            <el-table-column prop="created_at" label="创建时间">
              <template #default="scope">
                {{ formatDate(scope.row.created_at) }}
              </template>
            </el-table-column>
          </el-table>
          
          <div v-if="!loading && (!dashboardData.recent_platforms || dashboardData.recent_platforms.length === 0)"
               class="no-data">
            <el-empty description="暂无数据" />
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script>
import { ref, onMounted } from 'vue'
import { dashboardAPI } from '@/utils/api'
import { ElMessage } from 'element-plus'

export default {
  name: 'DashboardPage',
  setup() {
    const dashboardData = ref({})
    const loading = ref(true)

    const loadDashboardData = async () => {
      try {
        loading.value = true
        console.log('开始加载Dashboard数据...')
        
        const data = await dashboardAPI.getDashboard()
        console.log('Dashboard数据:', data)
        
        dashboardData.value = data
      } catch (error) {
        console.error('加载Dashboard数据失败:', error)
        ElMessage.error('加载数据失败: ' + error.message)
      } finally {
        loading.value = false
      }
    }

    const formatDate = (dateString) => {
      if (!dateString) return '-'
      try {
        return new Date(dateString).toLocaleDateString('zh-CN')
      } catch {
        return '-'
      }
    }

    onMounted(() => {
      loadDashboardData()
    })

    return {
      dashboardData,
      loading,
      formatDate
    }
  }
}
</script>

<style scoped>
.dashboard {
  padding: 20px;
}

.stat-card {
  text-align: center;
  cursor: pointer;
  transition: all 0.3s;
}

.stat-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.1);
}

.stat-item {
  padding: 20px;
}

.stat-number {
  font-size: 32px;
  font-weight: bold;
  color: #409EFF;
  margin-bottom: 8px;
}

.stat-label {
  font-size: 14px;
  color: #606266;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-weight: bold;
}

.no-data {
  padding: 20px;
  text-align: center;
}

.loading-container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 200px;
}

.dashboard-list-card {
  height: 100%; /* 使卡片填满其 <el-col> 父容器的高度 */
  display: flex;
  flex-direction: column;
}

.dashboard-list-card > :deep(.el-card__header) {
  flex-shrink: 0; /* 防止 header 在空间不足时被压缩 */
}

.dashboard-list-card > :deep(.el-card__body) {
  flex-grow: 1; /* 使卡片主体部分伸展以填充剩余空间 */
  overflow-y: auto; /* 如果内部表格内容过高，允许卡片主体滚动 */
}
</style>