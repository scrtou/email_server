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
            <el-table :data="upcomingRenewalsData" style="width: 100%" height="300px" empty-text="暂无即将到期的订阅">
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
  --primary-color: #4A90E2; /* 活力蓝 */
  --secondary-color: #50E3C2; /* 清新绿 */
  --accent-color: #F5A623; /* 强调橙 */
  --text-color-primary: #2c3e50; /* 深灰蓝 - 主要文字 */
  --text-color-secondary: #57606f; /* 次要文字 - 稍浅灰 */
  --text-color-light: #8a94a6; /* 辅助文字 - 更浅灰 */
  --bg-color: #f4f6f9; /* 更柔和的背景 */
  --card-bg-color: #ffffff;
  --border-color: #e6eaf0; /* 边框颜色 */
  --shadow-color-light: rgba(0, 0, 0, 0.05);
  --shadow-color-medium: rgba(0, 0, 0, 0.08);

  --font-family-main: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
  --font-size-xl: 28px; /* 用于主要统计数值 */
  --font-size-large: 20px; /* 用于卡片标题或重要数字 */
  --font-size-medium: 16px; /* 用于副标题、标签 */
  --font-size-normal: 14px; /* 正文、表格内容 */
  --font-size-small: 12px; /* 辅助性小字 */

  --card-border-radius: 12px; /* 更圆润的卡片 */
  --card-padding: 24px; /* 统一卡片内边距 */
  --card-shadow: 0 4px 12px var(--shadow-color-light);
  --card-hover-shadow: 0 6px 18px var(--shadow-color-medium);
  --el-card-padding: var(--card-padding); /* 覆盖 Element Plus 默认内边距 */

  font-family: var(--font-family-main);
  background-color: var(--bg-color);
  padding: 24px;
  color: var(--text-color-primary);
}

/* 统一 el-card 样式 */
:deep(.el-card) {
  border-radius: var(--card-border-radius) !important;
  border: 1px solid var(--border-color) !important;
  box-shadow: var(--card-shadow) !important;
  transition: box-shadow 0.3s ease-in-out, transform 0.3s ease-in-out !important;
  background-color: var(--card-bg-color) !important;
}

:deep(.el-card:hover) {
  box-shadow: var(--card-hover-shadow) !important;
  transform: translateY(-3px);
}

:deep(.el-card__header) {
  font-size: var(--font-size-medium);
  font-weight: 600;
  color: var(--text-color-primary);
  border-bottom: 1px solid var(--border-color);
  padding: 16px var(--card-padding); /* 调整头部内边距 */
}

:deep(.el-card__body) {
  padding: var(--card-padding)  !important; /* 确保内边距应用 */
}

/* 顶部统计卡片 - 四个总数 */
.stat-row .el-card {
  text-align: center;
}
.stat-label {
  font-size: var(--font-size-normal);
  color: var(--text-color-secondary);
  margin-bottom: 8px;
}
.stat-value {
  font-size: var(--font-size-xl);
  font-weight: 700;
  color: var(--primary-color);
  line-height: 1.2;
}

/* Element Plus Statistic 组件样式调整 */
:deep(.el-statistic__head) {
  font-size: var(--font-size-normal);
  color: var(--text-color-secondary) !important;
  margin-bottom: 8px !important;
}
:deep(.el-statistic__content) {
  font-size: var(--font-size-large) !important;
  font-weight: 600 !important;
  color: var(--text-color-primary) !important;
}
:deep(.el-statistic__content .el-statistic__formatter) {
  font-size: var(--font-size-large) !important;
  font-weight: 600 !important;
  color: var(--text-color-primary) !important;
}


/* 分割线 */
:deep(.el-divider--horizontal) {
  margin: 24px 0; /* 增加上下间距 */
  border-top: 1px solid var(--border-color);
}

/* 卡片头部自定义 */
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  /* 字体已在 el-card__header 中定义 */
}

/* 表格样式 */
:deep(.el-table) {
  border-radius: 8px; /* 给表格本身也加圆角 */
  overflow: hidden; /* 配合圆角 */
}
:deep(.el-table th.el-table__cell) {
  background-color: var(--bg-color) !important; /* 表头背景 */
  color: var(--text-color-primary) !important;
  font-weight: 600;
  font-size: var(--font-size-normal);
}
:deep(.el-table td.el-table__cell),
:deep(.el-table th.el-table__cell.is-leaf) {
  border-bottom: 1px solid var(--border-color);
}
:deep(.el-table .el-table__row:hover > td) {
  background-color: #eef2f7 !important; /* 行悬停颜色 */
}
:deep(.el-table__empty-text) {
  color: var(--text-color-light);
  font-size: var(--font-size-normal);
}

/* 标签样式 */
:deep(.el-tag) {
  border-radius: 4px;
  font-weight: 500;
}
:deep(.el-tag--success) { background-color: #e1f5ec; color: #00796b; border-color: #b2dfdb;}
:deep(.el-tag--info) { background-color: #e3f2fd; color: #1565c0; border-color: #bbdefb;}
:deep(.el-tag--warning) { background-color: #fff8e1; color: #f57f17; border-color: #ffecb3;}
:deep(.el-tag--danger) { background-color: #ffebee; color: #c62828; border-color: #ffcdd2;}


/* 加载和错误状态 */
.loading-container,
.error-container,
.no-data-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 300px; /* 确保有一定高度 */
  padding: 20px;
}
:deep(.el-loading-text) {
  color: var(--primary-color) !important;
}
:deep(.el-empty__description p) {
  color: var(--text-color-secondary) !important;
  font-size: var(--font-size-medium);
}
.no-data-table .el-empty, .no-data-chart .el-empty {
  padding: 20px 0; /* 调整卡片内无数据提示的间距 */
}


/* 图表容器 */
.chart {
  width: 100%;
  height: 300px; /* 确保图表有固定高度 */
}

/* 列表卡片特定样式 */
.dashboard-list-card {
  height: 100%;
  display: flex;
  flex-direction: column;
}
.dashboard-list-card > :deep(.el-card__body) {
  flex-grow: 1;
  overflow-y: auto;
  padding: 0 !important; /* 表格通常不需要卡片body的内边距 */
}
.dashboard-list-card > :deep(.el-card__body .el-table) {
  border-radius: 0 0 var(--card-border-radius) var(--card-border-radius); /* 底部圆角与卡片一致 */
}

/* 图表卡片特定样式 */
.dashboard-chart-card > :deep(.el-card__body) {
  display: flex;
  justify-content: center;
  align-items: center;
}

/* 响应式调整 */
@media (max-width: 768px) {
  .dashboard {
    padding: 16px;
  }
  :deep(.el-card) {
    --card-padding: 16px; /* 移动端减小内边距 */
  }
  .stat-value {
    font-size: calc(var(--font-size-xl) * 0.85); /* 缩小统计数字 */
  }
  :deep(.el-statistic__content),
  :deep(.el-statistic__content .el-statistic__formatter) {
    font-size: calc(var(--font-size-large) * 0.9) !important;
  }
  .chart {
    height: 250px; /* 移动端图表高度 */
  }
}
</style>