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
      <el-row :gutter="15">
        <!-- 即将到期订阅列表 -->
        <el-col :xs="24" :md="14">
          <el-card class="dashboard-list-card" shadow="hover">
            <template #header>
              <div class="card-header">
                <span>即将到期订阅 (未来30天)</span>
              </div>
            </template>
            
            <!-- 优化后的表格容器 -->
            <div class="table-container">
              <!-- 有数据或加载中时显示表格 -->
              <el-table 
                v-if="loading || (upcomingRenewalsData && upcomingRenewalsData.length > 0)"
                :data="upcomingRenewalsData" 
                style="width: 100%" 
                :max-height="calculateTableHeight()"
                :sticky-header="true"
                :show-overflow-tooltip="true"
                table-layout="auto"
                v-loading="loading"
                element-loading-text="加载中..."
              >
                <el-table-column 
                  prop="service_name" 
                  label="服务名称" 
                  sortable
                  min-width="120"
                  :show-overflow-tooltip="true"
                />
                <el-table-column 
                  prop="platform_name" 
                  label="平台" 
                  sortable
                  min-width="80"
                  :show-overflow-tooltip="true"
                />
                <el-table-column 
                  prop="next_renewal_date" 
                  label="到期日" 
                  sortable
                  min-width="110"
                  :show-overflow-tooltip="true"
                >
                  <template #default="scope">
                    {{ formatDate(scope.row.next_renewal_date) }}
                  </template>
                </el-table-column>
                <el-table-column 
                  prop="cost" 
                  label="费用" 
                  sortable
                  min-width="140"
                  :show-overflow-tooltip="true"
                >
                  <template #default="scope">
                    {{ formatCurrency(scope.row.cost) }} / {{ scope.row.billing_cycle }}
                  </template>
                </el-table-column>
                <el-table-column 
                  prop="status" 
                  label="状态" 
                  sortable
                  min-width="90"
                  :show-overflow-tooltip="true"
                >
                  <template #default="scope">
                    <el-tag :type="getStatusTagType(scope.row.status)" disable-transitions>
                      {{ scope.row.status }}
                    </el-tag>
                  </template>
                </el-table-column>
              </el-table>
              
              <!-- 无数据时显示自定义空状态 -->
              <div v-else class="no-data-table">
                <el-empty 
                  description="太棒了！近期没有即将到期的订阅。"
                  :image-size="160"
                >
                  <template #image>
                    <svg viewBox="0 0 64 41" xmlns="http://www.w3.org/2000/svg" style="width: 160px; height: 100px;">
                      <g transform="translate(0 1)" fill="none" fill-rule="evenodd">
                        <ellipse fill="#f5f5f5" cx="32" cy="33" rx="32" ry="7"/>
                        <g fill-rule="nonzero" stroke="#d9d9d9">
                          <path d="M55 12.76L44.854 1.258C44.367.474 43.656 0 42.907 0H21.093c-.749 0-1.46.474-1.947 1.257L9 12.761V22h46v-9.24z"/>
                          <path d="M41.613 15.931c0-1.605.994-2.93 2.227-2.931H55v18.137C55 33.26 53.68 35 52.05 35h-40.1C10.32 35 9 33.259 9 31.137V13h11.16c1.233 0 2.227 1.323 2.227 2.928v.022c0 1.605 1.005 2.901 2.237 2.901h14.752c1.232 0 2.237-1.308 2.237-2.913v-.007z" fill="#fafafa"/>
                        </g>
                      </g>
                    </svg>
                  </template>
                </el-empty>
              </div>
            </div>
          </el-card>
        </el-col>

        <!-- 各平台订阅数量分布 -->
        <el-col :xs="24" :md="10">
          <el-card class="dashboard-chart-card" shadow="hover">
            <template #header>
              <div class="card-header">
                <span>各平台订阅数量</span>
              </div>
            </template>
            <div class="chart-container">
              <v-chart
                v-if="!loading && summaryData.subscriptions_by_platform && summaryData.subscriptions_by_platform.length > 0"
                class="chart"
                :option="platformDistributionChartOptions"
                autoresize
              />
              <div v-if="!loading && (!summaryData.subscriptions_by_platform || summaryData.subscriptions_by_platform.length === 0)" class="no-data-chart">
                <el-empty description="暂无平台订阅数据" :image-size="160">
                  <template #image>
                    <svg viewBox="0 0 64 41" xmlns="http://www.w3.org/2000/svg" style="width: 160px; height: 100px;">
                      <g transform="translate(0 1)" fill="none" fill-rule="evenodd">
                        <ellipse fill="#f5f5f5" cx="32" cy="33" rx="32" ry="7"/>
                        <g fill-rule="nonzero" stroke="#d9d9d9">
                          <path d="M55 12.76L44.854 1.258C44.367.474 43.656 0 42.907 0H21.093c-.749 0-1.46.474-1.947 1.257L9 12.761V22h46v-9.24z"/>
                          <path d="M41.613 15.931c0-1.605.994-2.93 2.227-2.931H55v18.137C55 33.26 53.68 35 52.05 35h-40.1C10.32 35 9 33.259 9 31.137V13h11.16c1.233 0 2.227 1.323 2.227 2.928v.022c0 1.605 1.005 2.901 2.237 2.901h14.752c1.232 0 2.237-1.308 2.237-2.913v-.007z" fill="#fafafa"/>
                        </g>
                      </g>
                    </svg>
                  </template>
                </el-empty>
              </div>
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
const upcomingRenewalsData = computed(() => dashboardStore.upcomingRenewals) // This is from dashboardStore, not notificationStore
const loading = computed(() => dashboardStore.isLoading) // Dashboard loading state
const error = computed(() => dashboardStore.error) // Dashboard error state

// 修改：动态计算表格高度的方法
const calculateTableHeight = () => {
  // 由于卡片现在有固定高度，表格应该填满可用空间
  // 卡片高度480px - 卡片头部约65px - 卡片内边距等约15px = 约400px可用高度
  return 400;
};

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
  dashboardStore.fetchDashboardSummary();
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

/* === 优化的表格容器样式 === */
.table-container {
  position: relative;
  width: 100%;
  height: 100%; /* 填满卡片body的高度 */
  overflow: hidden;
  border-radius: var(--radius-lg);
  display: flex;
  flex-direction: column;
}

/* === 优化的列表卡片样式 === */
.dashboard-list-card {
  height: 430px; /* 设置固定高度与右侧卡片一致 */
  min-height: 430px;
  max-height: 430px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.dashboard-list-card > :deep(.el-card__body) {
  flex-grow: 1;
  overflow: hidden; /* 确保卡片本身不产生滚动条 */
  padding: 0 !important;
  display: flex;
  flex-direction: column;
}

/* === 图表卡片样式调整 === */
.dashboard-chart-card {
  height: 430px; /* 与左侧卡片保持一致的高度 */
  min-height: 430px;
  max-height: 430px;
  display: flex;
  flex-direction: column;
}

.dashboard-chart-card > :deep(.el-card__body) {
  flex-grow: 1;
  display: flex;
  justify-content: center;
  align-items: center;
  padding: var(--space-3) !important;
  overflow: hidden;
}

/* === 图表容器样式 === */
.chart-container {
  width: 100%;
  height: 100%;
  display: flex;
  justify-content: center;
  align-items: center;
  position: relative;
}

/* === 优化的表格样式 === */
:deep(.el-table) {
  border-radius: var(--radius-lg);
  border: 1px solid var(--color-gray-200);
  overflow: hidden;
  flex-grow: 1;
  height: 100%; /* 填满容器高度 */
}

/* 表头固定样式 */
:deep(.el-table .el-table__header-wrapper) {
  position: sticky;
  top: 0;
  z-index: 10;
  background: linear-gradient(135deg, var(--color-gray-50), var(--color-gray-100));
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
}

:deep(.el-table th.el-table__cell) {
  background: transparent !important; /* 因为父容器已有背景 */
  color: var(--color-gray-700) !important;
  font-weight: var(--font-semibold);
  font-size: var(--text-sm);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  border-bottom: 2px solid var(--color-gray-200);
  padding: var(--space-3) var(--space-2);
  white-space: nowrap;
  position: relative;
}

:deep(.el-table td.el-table__cell) {
  border-bottom: 1px solid var(--color-gray-100);
  font-size: var(--text-sm);
  color: var(--color-gray-700);
  padding: var(--space-3) var(--space-2);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* 表格体滚动样式 */
:deep(.el-table .el-table__body-wrapper) {
  overflow-y: auto;
  overflow-x: hidden;
  max-height: calc(100% - 50px); /* 减去表头高度 */
}

/* 自定义滚动条样式 */
:deep(.el-table .el-table__body-wrapper::-webkit-scrollbar) {
  width: 6px;
}

:deep(.el-table .el-table__body-wrapper::-webkit-scrollbar-track) {
  background: var(--color-gray-100);
  border-radius: 3px;
}

:deep(.el-table .el-table__body-wrapper::-webkit-scrollbar-thumb) {
  background: var(--color-gray-300);
  border-radius: 3px;
  transition: background var(--transition-base);
}

:deep(.el-table .el-table__body-wrapper::-webkit-scrollbar-thumb:hover) {
  background: var(--color-gray-400);
}

/* 表格行悬停效果优化 */
:deep(.el-table .el-table__row:hover > td) {
  background: linear-gradient(135deg, var(--color-primary-50), rgba(59, 130, 246, 0.05)) !important;
  transform: none; /* 移除transform避免影响布局 */
  transition: background-color var(--transition-base);
}

:deep(.el-table__empty-text) {
  color: var(--color-gray-500);
  font-size: var(--text-base);
  font-weight: var(--font-medium);
  padding: var(--space-8) var(--space-4);
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
  white-space: nowrap;
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

/* 无数据状态优化 */
.no-data-table {
  display: flex;
  justify-content: center;
  align-items: center;
  flex-grow: 1; /* 填满剩余空间 */
  padding: var(--space-6);
}

.no-data-table .el-empty {
  padding: 0;
}

.no-data-chart {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100%;
  width: 100%;
  position: absolute;
  top: 0;
  left: 0;
  background: rgba(255, 255, 255, 0.9);
  border-radius: var(--radius-lg);
}

.no-data-chart .el-empty {
  padding: var(--space-4) 0;
}

/* === Enhanced Chart Styles === */
.chart {
  width: 100%;
  height: 100%; /* 填满卡片body的高度 */
  min-height: 350px; /* 设置最小高度确保图表可见 */
  border-radius: var(--radius-lg);
  overflow: hidden;
}

/* === 额外的行间距优化 === */
.stat-row {
  margin-bottom: var(--space-2); /* 减小统计行的底部间距 */
}

/* === 加载状态优化 === */
:deep(.el-loading-mask) {
  border-radius: var(--radius-lg);
  background: rgba(255, 255, 255, 0.8);
  backdrop-filter: blur(4px);
}

/* === 列宽优化 === */
:deep(.el-table .el-table__cell) {
  word-break: normal;
  word-wrap: break-word;
}

/* 特定列的宽度调整 */
:deep(.el-table .cell) {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  line-height: var(--leading-relaxed);
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

  .dashboard-list-card,
  .dashboard-chart-card {
    height: 400px; /* 移动端统一高度 */
    min-height: 400px;
    max-height: 400px;
  }
  
  .table-container {
    overflow-x: auto; /* 移动端允许横向滚动 */
  }
  
  :deep(.el-table) {
    min-width: 600px; /* 确保移动端表格有最小宽度 */
  }
  
  :deep(.el-table td.el-table__cell),
  :deep(.el-table th.el-table__cell) {
    padding: var(--space-2) var(--space-1);
    font-size: var(--text-xs);
  }

  .chart {
    min-height: 280px; /* 移动端图表最小高度 */
  }

  :deep(.el-divider--horizontal) {
    margin: var(--space-3) 0; /* 移动端减小分隔线间距 */
  }
}

@media (max-width: 480px) {
  .dashboard-list-card,
  .dashboard-chart-card {
    height: 350px; /* 小屏幕设备统一高度 */
    min-height: 350px;
    max-height: 350px;
  }
  
  :deep(.el-table) {
    min-width: 550px;
  }

  .chart {
    min-height: 250px;
  }
}
</style>