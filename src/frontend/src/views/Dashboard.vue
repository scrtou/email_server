<template>
  <div class="dashboard">
    <div v-if="loading" class="loading-container">
      <el-loading text="加载中..." fullscreen />
    </div>
    <div v-else-if="error" class="error-container">
      <el-empty :description="`加载失败: ${error}`" />
    </div>
    <div v-else-if="summaryData">
      <!-- 其他统计信息 -->
       <el-row :gutter="20" class="stat-row">
        <el-col :xs="12" :sm="6">
          <el-card shadow="hover" class="text-center">
            <div class="stat-label">总邮箱账户</div>
            <div class="stat-value">{{ summaryData.total_email_accounts }}</div>
          </el-card>
        </el-col>
        <el-col :xs="12" :sm="6">
          <el-card shadow="hover" class="text-center">
            <div class="stat-label">总平台数</div>
            <div class="stat-value">{{ summaryData.total_platforms }}</div>
          </el-card>
        </el-col>
        <el-col :xs="12" :sm="6">
          <el-card shadow="hover" class="text-center">
            <div class="stat-label">总平台注册</div>
            <div class="stat-value">{{ summaryData.total_platform_registrations }}</div>
          </el-card>
        </el-col>
        <el-col :xs="12" :sm="6">
          <el-card shadow="hover" class="text-center">
            <div class="stat-label">总服务订阅</div>
            <div class="stat-value">{{ summaryData.total_service_subscriptions }}</div>
          </el-card>
        </el-col>
      </el-row>

      <el-divider />

      <!-- 顶部统计卡片 -->
      <el-row :gutter="20" class="stat-row">
        <el-col :xs="24" :sm="12" :md="6">
          <el-card shadow="hover">
            <el-statistic title="活跃订阅数" :value="summaryData.active_subscriptions_count" />
          </el-card>
        </el-col>
        <el-col :xs="24" :sm="12" :md="6">
          <el-card shadow="hover">
            <el-statistic title="预估月支出" :value="summaryData.estimated_monthly_spending">
              <template #formatter>
                {{ formatCurrency(summaryData.estimated_monthly_spending) }}
              </template>
            </el-statistic>
          </el-card>
        </el-col>
        <el-col :xs="24" :sm="12" :md="6">
          <el-card shadow="hover">
            <el-statistic title="预估年支出" :value="summaryData.estimated_yearly_spending">
               <template #formatter>
                {{ formatCurrency(summaryData.estimated_yearly_spending) }}
              </template>
            </el-statistic>
          </el-card>
        </el-col>
        <el-col :xs="24" :sm="12" :md="6">
          <el-card shadow="hover">
            <el-statistic title="即将到期 (30天内)" :value="upcomingRenewalsData?.length || 0" />
          </el-card>
        </el-col>
      </el-row>

      <el-divider />

      <!-- 图表和列表 -->
      <el-row :gutter="20">
        <!-- 即将到期订阅列表 -->
        <el-col :xs="24" :md="12">
          <el-card class="dashboard-list-card" shadow="hover">
            <template #header>
              <div class="card-header">
                <span>即将到期订阅 (未来30天)</span>
              </div>
            </template>
            <el-table :data="upcomingRenewalsData" style="width: 100%" height="380px" empty-text="暂无即将到期的订阅">
              <el-table-column prop="service_name" label="服务名称" sortable />
              <el-table-column prop="platform_name" label="平台" sortable />
              <el-table-column prop="next_renewal_date" label="到期日" sortable>
                <template #default="scope">
                  {{ formatDate(scope.row.next_renewal_date) }}
                </template>
              </el-table-column>
              <el-table-column prop="cost" label="费用" sortable>
                <template #default="scope">
                  {{ formatCurrency(scope.row.cost) }} / {{ scope.row.billing_cycle }}
                </template>
              </el-table-column>
               <el-table-column prop="status" label="状态" width="100" sortable>
                <template #default="scope">
                  <el-tag :type="getStatusTagType(scope.row.status)" disable-transitions>{{ scope.row.status }}</el-tag>
                </template>
              </el-table-column>
            </el-table>
             <div v-if="!loading && upcomingRenewalsData?.length === 0" class="no-data-table">
              <el-empty description="太棒了！近期没有即将到期的订阅。" />
            </div>
          </el-card>
        </el-col>

        <!-- 各平台订阅数量分布 -->
        <el-col :xs="24" :md="12">
          <el-card class="dashboard-chart-card" shadow="hover">
            <template #header>
              <div class="card-header">
                <span>各平台订阅数量</span>
              </div>
            </template>
            <v-chart class="chart" :option="platformDistributionChartOptions" autoresize />
             <div v-if="!loading && summaryData.subscriptions_by_platform?.length === 0" class="no-data-chart">
              <el-empty description="暂无平台订阅数据以生成图表。" />
            </div>
          </el-card>
        </el-col>
      </el-row>

    </div>
    <div v-else class="no-data-container">
       <el-empty description="暂无仪表盘数据" />
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted } from 'vue' // ref is not directly used here for reactive data from store
import { useDashboardStore } from '@/stores/dashboard'
// ElMessage, ElRow, ElCol, ElCard, ElTable, ElTableColumn, ElLoading, ElEmpty, ElStatistic, ElDivider, ElTag are used in template, auto-imported by unplugin-vue-components/vite or unplugin-auto-import/vite if configured, otherwise keep explicit imports
import { format, parseISO } from 'date-fns' // 用于日期格式化

// ECharts
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { PieChart, BarChart } from 'echarts/charts' // BarChart might be for future use
import {
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  GridComponent,
  // DatasetComponent, // Not directly used in this pie chart example
  // TransformComponent // Not directly used in this pie chart example
} from 'echarts/components'
import VChart from 'vue-echarts' // Removed THEME_KEY as it's not used directly here without provide

use([
  CanvasRenderer,
  PieChart,
  BarChart, // Keep if future charts might use it
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  GridComponent,
  // DatasetComponent,
  // TransformComponent
])

const dashboardStore = useDashboardStore()

// 从 store 获取计算属性
const summaryData = computed(() => dashboardStore.getSummaryData) // Keep for other parts of summaryData
const upcomingRenewalsData = computed(() => dashboardStore.upcomingRenewals)
const loading = computed(() => dashboardStore.isLoading)
const error = computed(() => dashboardStore.error)

const formatDate = (dateString) => {
  if (!dateString) return '-'
  try {
    return format(parseISO(dateString), 'yyyy-MM-dd')
  } catch (e) {
    console.error('Error formatting date:', dateString, e)
    return dateString
  }
}

const formatCurrency = (value) => {
  if (typeof value !== 'number') {
    return 'N/A'
  }
  return `¥ ${value.toFixed(2)}`
}

const getStatusTagType = (status) => {
  switch (status?.toLowerCase()) {
    case 'active':
      return 'success'
    case 'cancelled':
      return 'info'
    case 'free_trial':
      return 'warning'
    case 'expired':
      return 'danger'
    default:
      return '' // Default or 'primary'
  }
}

// ECharts options for Platform Distribution
const platformDistributionChartOptions = computed(() => {
  const data = summaryData.value?.subscriptions_by_platform || []
  if (!data.length) {
    return {
      title: {
        text: '各平台订阅数量分布',
        left: 'center',
        textStyle: { fontSize: 16 }
      },
      graphic: { // Show a message when no data
        type: 'text',
        left: 'center',
        top: 'middle',
        style: {
          fill: '#999',
          text: '暂无数据',
          font: '14px Microsoft YaHei'
        }
      }
    };
  }
  return {
    title: {
      text: '各平台订阅数量分布',
      left: 'center',
      textStyle: { fontSize: 16 }
    },
    tooltip: {
      trigger: 'item',
      formatter: '{a} <br/>{b} : {c} ({d}%)'
    },
    legend: {
      orient: 'horizontal', // Changed to horizontal for better fit
      bottom: '0', // Positioned at the bottom
      data: data.map(item => item.platform_name)
    },
    series: [
      {
        name: '订阅数',
        type: 'pie',
        radius: ['40%', '70%'], // Make it a donut chart
        center: ['50%', '50%'], // Adjusted center for legend at bottom
        avoidLabelOverlap: false,
        label: {
          show: false,
          position: 'center'
        },
        emphasis: {
          label: {
            show: true,
            fontSize: '20',
            fontWeight: 'bold'
          }
        },
        labelLine: {
          show: false
        },
        data: data.map(item => ({ value: item.subscription_count, name: item.platform_name })),
      }
    ]
  }
})

onMounted(() => {
  dashboardStore.fetchDashboardSummary()
})

</script>

<style scoped>
.dashboard {
  /* Use global design tokens */
  font-family: var(--font-family-sans);
  background: transparent;
  padding: var(--space-4); /* 减小四周内边距 */
  color: var(--color-gray-800);
  position: relative;
  max-height: 100vh; /* 限制最大高度为视口高度 */
  overflow-y: auto; /* 超出时垂直滚动 */
}

/* === Enhanced Card Styles === */
:deep(.el-card) {
  border-radius: var(--radius-xl) !important;
  border: 1px solid var(--color-gray-200) !important;
  box-shadow: var(--shadow-base) !important;
  transition: all var(--transition-base) !important;
  background: rgba(255, 255, 255, 0.8) !important;
  backdrop-filter: blur(10px) !important;
  position: relative;
  overflow: hidden;
}

:deep(.el-card::before) {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 4px;
  background: linear-gradient(90deg, var(--color-primary-500), var(--color-success-500), var(--color-warning-500));
  opacity: 0;
  transition: opacity var(--transition-base);
}

:deep(.el-card:hover) {
  box-shadow: var(--shadow-xl) !important;
  transform: translateY(-2px) scale(1.01); /* 减小hover效果 */
  border-color: var(--color-primary-200) !important;
}

:deep(.el-card:hover::before) {
  opacity: 1;
}

:deep(.el-card__header) {
  font-size: var(--text-base); /* 减小标题字体 */
  font-weight: var(--font-semibold);
  color: var(--color-gray-800);
  border-bottom: 1px solid var(--color-gray-200);
  padding: var(--space-3) var(--space-4); /* 减小header内边距 */
  background: rgba(249, 250, 251, 0.5);
}

:deep(.el-card__body) {
  padding: var(--space-4) !important; /* 减小卡片内边距 */
}

/* === Enhanced Statistics Styles === */
.stat-row .el-card {
  text-align: center;
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.9), rgba(249, 250, 251, 0.8)) !important;
}

.stat-label {
  font-size: var(--text-xs); /* 保持小字体 */
  color: var(--color-gray-600);
  margin-bottom: var(--space-1); /* 减小底部间距 */
  font-weight: var(--font-medium);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.stat-value {
  font-size: var(--text-xl); /* 减小数值字体 */
  font-weight: var(--font-bold);
  background: linear-gradient(135deg, var(--color-primary-600), var(--color-primary-500));
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  line-height: var(--leading-tight);
  margin-bottom: 0; /* 移除底部间距 */
}

.stat-row .el-card > :deep(.el-card__body) {
  padding: var(--space-3) var(--space-2) !important; /* 进一步减小统计卡片的内边距 */
}

/* Enhanced Element Plus Statistic Styles */
:deep(.el-statistic__head) {
  font-size: var(--text-xs) !important;
  color: var(--color-gray-600) !important;
  margin-bottom: var(--space-1) !important; /* 减小间距 */
  font-weight: var(--font-medium) !important;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

:deep(.el-statistic__content) {
  font-size: var(--text-lg) !important; /* 减小字体 */
  font-weight: var(--font-bold) !important;
  background: linear-gradient(135deg, var(--color-primary-600), var(--color-success-500)) !important;
  -webkit-background-clip: text !important;
  -webkit-text-fill-color: transparent !important;
  background-clip: text !important;
}

:deep(.el-statistic__content .el-statistic__formatter) {
  font-size: var(--text-lg) !important; /* 减小字体 */
  font-weight: var(--font-bold) !important;
}

/* === Enhanced Divider Styles === */
:deep(.el-divider--horizontal) {
  margin: var(--space-4) 0; /* 大幅减小分隔线间距 */
  border-top: 2px solid transparent;
  background: linear-gradient(90deg, transparent, var(--color-gray-200), transparent);
  height: 2px;
  border: none;
}

/* === Enhanced Card Header Styles === */
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

/* === Enhanced Table Styles === */
:deep(.el-table) {
  border-radius: var(--radius-lg);
  overflow: hidden;
  box-shadow: var(--shadow-sm);
  border: 1px solid var(--color-gray-200);
}

:deep(.el-table th.el-table__cell) {
  background: linear-gradient(135deg, var(--color-gray-50), var(--color-gray-100)) !important;
  color: var(--color-gray-700) !important;
  font-weight: var(--font-semibold);
  font-size: var(--text-sm);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  border-bottom: 2px solid var(--color-gray-200);
  padding: var(--space-2) var(--space-3); /* 减小表格header内边距 */
}

:deep(.el-table td.el-table__cell) {
  border-bottom: 1px solid var(--color-gray-100);
  font-size: var(--text-sm);
  color: var(--color-gray-700);
  padding: var(--space-2) var(--space-3); /* 减小表格cell内边距 */
}

:deep(.el-table .el-table__row:hover > td) {
  background: linear-gradient(135deg, var(--color-primary-50), rgba(59, 130, 246, 0.05)) !important;
  transform: scale(1.005); /* 减小hover效果 */
  transition: all var(--transition-base);
}

:deep(.el-table__empty-text) {
  color: var(--color-gray-500);
  font-size: var(--text-base);
  font-weight: var(--font-medium);
}

/* === Enhanced Tag Styles === */
:deep(.el-tag) {
  border-radius: var(--radius-full);
  font-weight: var(--font-medium);
  font-size: var(--text-xs);
  padding: var(--space-1) var(--space-2); /* 减小tag内边距 */
  border: none;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

:deep(.el-tag--success) {
  background: linear-gradient(135deg, var(--color-success-100), var(--color-success-200));
  color: var(--color-success-700);
}

:deep(.el-tag--info) {
  background: linear-gradient(135deg, var(--color-info-100), var(--color-info-200));
  color: var(--color-info-700);
}

:deep(.el-tag--warning) {
  background: linear-gradient(135deg, var(--color-warning-100), var(--color-warning-200));
  color: var(--color-warning-700);
}

:deep(.el-tag--danger) {
  background: linear-gradient(135deg, var(--color-error-100), var(--color-error-200));
  color: var(--color-error-700);
}

/* === Enhanced Loading and Error States === */
.loading-container,
.error-container,
.no-data-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 200px; /* 减小最小高度 */
  padding: var(--space-4); /* 减小内边距 */
  border-radius: var(--radius-lg);
  background: rgba(255, 255, 255, 0.5);
  backdrop-filter: blur(10px);
}

:deep(.el-loading-text) {
  color: var(--color-primary-600) !important;
  font-weight: var(--font-medium) !important;
}

:deep(.el-empty__description p) {
  color: var(--color-gray-600) !important;
  font-size: var(--text-base);
  font-weight: var(--font-medium);
}

.no-data-table .el-empty,
.no-data-chart .el-empty {
  padding: var(--space-4) 0; /* 减小内边距 */
}

/* === Enhanced Chart Styles === */
.chart {
  width: 100%;
  height: 380px; /* 减小图表高度 */
  border-radius: var(--radius-lg);
  overflow: hidden;
}

/* === Enhanced Card Layout Styles === */
.dashboard-list-card {
  height: 100%;
  display: flex;
  flex-direction: column;
  max-height: 420px; /* 限制卡片最大高度 */
}

.dashboard-list-card > :deep(.el-card__body) {
  flex-grow: 1;
  overflow-y: auto;
  padding: 0 !important;
}

.dashboard-list-card > :deep(.el-card__body .el-table) {
  border-radius: 0 0 var(--radius-xl) var(--radius-xl);
  border: none;
}

.dashboard-chart-card {
  height: 100%;
  max-height: 420px; /* 限制图表卡片最大高度 */
}

.dashboard-chart-card > :deep(.el-card__body) {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: var(--space-3) !important; /* 减小图表卡片内边距 */
}

/* === 额外的行间距优化 === */
.stat-row {
  margin-bottom: var(--space-2); /* 减小统计行的底部间距 */
}

/* === 表格高度优化 === */
.dashboard-list-card .el-table {
  max-height: 400px; /* 限制表格最大高度 */
}

/* === Enhanced Responsive Design === */
@media (max-width: 768px) {
  .dashboard {
    padding: var(--space-3); /* 移动端进一步减小内边距 */
  }

  :deep(.el-card__body) {
    padding: var(--space-3) !important;
  }

  .stat-value {
    font-size: var(--text-lg); /* 移动端减小字体 */
  }

  :deep(.el-statistic__content),
  :deep(.el-statistic__content .el-statistic__formatter) {
    font-size: var(--text-base) !important; /* 移动端减小字体 */
  }

  .chart {
    height: 200px; /* 移动端进一步减小图表高度 */
  }

  .dashboard-list-card,
  .dashboard-chart-card {
    max-height: 280px; /* 移动端减小卡片高度 */
  }

  :deep(.el-divider--horizontal) {
    margin: var(--space-3) 0; /* 移动端减小分隔线间距 */
  }
}
</style>