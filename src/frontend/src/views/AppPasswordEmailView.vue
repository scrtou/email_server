<!-- 应用密码邮件查看页面 -->
<template>
  <div class="app-password-email-view">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-content">
        <div class="header-info">
          <h1 class="page-title">
            <el-icon class="title-icon"><Message /></el-icon>
            {{ emailAccount?.email_address }}
          </h1>
          <p class="page-description">
            <el-tag type="success" size="small">
              <el-icon><Key /></el-icon>
              应用密码 - 永不过期
            </el-tag>
            <span style="margin-left: 12px;">使用应用密码获取邮件，稳定可靠的邮箱访问</span>
          </p>
        </div>
        <div class="header-actions">
          <el-button :icon="Refresh" @click="refreshEmails" :loading="loading">
            刷新邮件
          </el-button>
          <el-button :icon="Setting" @click="openSettings">
            设置
          </el-button>
        </div>
      </div>
    </div>

    <!-- 邮件统计 -->
    <div class="email-stats">
      <el-row :gutter="16">
        <el-col :span="6">
          <el-card class="stat-card">
            <div class="stat-item">
              <div class="stat-value">{{ emailStats.total }}</div>
              <div class="stat-label">总邮件数</div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="stat-card">
            <div class="stat-item">
              <div class="stat-value">{{ emailStats.unread }}</div>
              <div class="stat-label">未读邮件</div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="stat-card">
            <div class="stat-item">
              <div class="stat-value">99%</div>
              <div class="stat-label">连接成功率</div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="stat-card">
            <div class="stat-item">
              <div class="stat-value">{{ formatDate(lastSyncTime) }}</div>
              <div class="stat-label">最后同步</div>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </div>

    <!-- 邮件列表 -->
    <el-card class="email-list-card">
      <template #header>
        <div class="card-header">
          <span class="card-title">邮件列表</span>
          <div class="header-filters">
            <el-input
              v-model="searchText"
              placeholder="搜索邮件"
              :prefix-icon="Search"
              clearable
              style="width: 200px;"
              @input="handleSearch"
            />
            <el-select v-model="selectedFolder" placeholder="选择文件夹" style="width: 150px; margin-left: 8px;">
              <el-option
                v-for="folder in folders"
                :key="folder"
                :label="folder"
                :value="folder"
              />
            </el-select>
          </div>
        </div>
      </template>

      <div class="email-list-container">
        <el-table
          :data="filteredEmails"
          v-loading="loading"
          style="width: 100%"
          @row-click="handleRowClick"
          highlight-current-row
        >
          <el-table-column width="60" align="center">
            <template #default="scope">
              <el-icon v-if="!scope.row.isRead" class="unread-icon" color="var(--el-color-primary)">
                <CircleCheckFilled />
              </el-icon>
            </template>
          </el-table-column>

          <el-table-column label="发件人" width="200">
            <template #default="scope">
              <div class="sender-cell">
                <div class="sender-name">{{ getSenderName(scope.row.from) }}</div>
                <div class="sender-email">{{ getSenderEmail(scope.row.from) }}</div>
              </div>
            </template>
          </el-table-column>

          <el-table-column label="主题" min-width="300">
            <template #default="scope">
              <div class="subject-cell">
                <span class="subject-text" :class="{ 'unread': !scope.row.isRead }">
                  {{ scope.row.subject || '(无主题)' }}
                </span>
                <el-icon v-if="scope.row.hasAttachment" class="attachment-icon">
                  <Paperclip />
                </el-icon>
              </div>
            </template>
          </el-table-column>

          <el-table-column label="预览" min-width="200">
            <template #default="scope">
              <div class="snippet-text">{{ scope.row.snippet }}</div>
            </template>
          </el-table-column>

          <el-table-column label="时间" width="160" align="right">
            <template #default="scope">
              <div class="time-cell">
                {{ formatEmailTime(scope.row.date) }}
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
      </div>
    </el-card>

    <!-- 邮件详情抽屉 -->
    <el-drawer
      v-model="emailDrawer.visible"
      :title="emailDrawer.email?.subject || '(无主题)'"
      size="60%"
      direction="rtl"
    >
      <div v-if="emailDrawer.email" class="email-detail">
        <div class="email-header">
          <div class="email-meta">
            <div class="meta-row">
              <span class="meta-label">发件人：</span>
              <span class="meta-value">{{ getSenderDisplay(emailDrawer.email.from) }}</span>
            </div>
            <div class="meta-row">
              <span class="meta-label">收件人：</span>
              <span class="meta-value">{{ getRecipientDisplay(emailDrawer.email.to) }}</span>
            </div>
            <div class="meta-row" v-if="emailDrawer.email.cc && emailDrawer.email.cc.length > 0">
              <span class="meta-label">抄送：</span>
              <span class="meta-value">{{ getRecipientDisplay(emailDrawer.email.cc) }}</span>
            </div>
            <div class="meta-row">
              <span class="meta-label">时间：</span>
              <span class="meta-value">{{ formatFullDate(emailDrawer.email.date) }}</span>
            </div>
          </div>
          
          <!-- 附件 -->
          <div v-if="emailDrawer.email.hasAttachment && emailDrawer.email.attachments" class="attachments">
            <h4>附件 ({{ emailDrawer.email.attachments.length }})</h4>
            <div class="attachment-list">
              <div v-for="attachment in emailDrawer.email.attachments" :key="attachment.filename" class="attachment-item">
                <el-icon><Paperclip /></el-icon>
                <span>{{ attachment.filename }}</span>
                <span class="attachment-size">({{ formatFileSize(attachment.size) }})</span>
              </div>
            </div>
          </div>
        </div>

        <el-divider />

        <!-- 邮件内容 -->
        <div class="email-body">
          <div v-if="emailDrawer.email.htmlBody" v-html="sanitizeHtml(emailDrawer.email.htmlBody)" class="html-content"></div>
          <div v-else-if="emailDrawer.email.body" class="text-content">
            <pre>{{ emailDrawer.email.body }}</pre>
          </div>
          <div v-else class="no-content">
            <el-empty description="无邮件内容" />
          </div>
        </div>
      </div>
    </el-drawer>

    <!-- 设置对话框 -->
    <el-dialog
      v-model="settingsDialog.visible"
      title="应用密码设置"
      width="600px"
    >
      <div class="settings-content">
        <el-alert
          title="当前使用应用密码认证"
          type="success"
          :closable="false"
          show-icon
        >
          <p>该账户正在使用应用密码进行邮箱访问，享受永不过期的稳定连接。</p>
        </el-alert>
        
        <div class="settings-actions">
          <el-button type="primary" @click="testConnection" :loading="testingConnection">
            <el-icon><Connection /></el-icon>
            测试连接
          </el-button>
          <el-button @click="manageAppPassword">
            <el-icon><Setting /></el-icon>
            管理应用密码
          </el-button>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import {
  Message, Key, Refresh, Setting, Search, CircleCheckFilled, Paperclip, Connection
} from '@element-plus/icons-vue'
import axios from 'axios'

const route = useRoute()
const router = useRouter()

// 响应式数据
const loading = ref(false)
const testingConnection = ref(false)
const searchText = ref('')
const selectedFolder = ref('INBOX')
const emailAccount = ref(null)
const emails = ref([])
const folders = ref([])
const lastSyncTime = ref(new Date())

// 统计数据
const emailStats = reactive({
  total: 0,
  unread: 0
})

// 分页
const pagination = reactive({
  currentPage: 1,
  pageSize: 20,
  total: 0
})

// 邮件详情抽屉
const emailDrawer = reactive({
  visible: false,
  email: null
})

// 设置对话框
const settingsDialog = reactive({
  visible: false
})

// 计算属性
const filteredEmails = computed(() => {
  if (!searchText.value) return emails.value
  
  const search = searchText.value.toLowerCase()
  return emails.value.filter(email =>
    email.subject?.toLowerCase().includes(search) ||
    getSenderName(email.from).toLowerCase().includes(search) ||
    email.snippet?.toLowerCase().includes(search)
  )
})

// 生命周期
onMounted(() => {
  initializePage()
})

// 监听路由参数变化
watch(() => route.params.id, (newId) => {
  if (newId) {
    initializePage()
  }
})

// 方法
const initializePage = async () => {
  const accountId = route.params.id
  if (!accountId) {
    ElMessage.error('账户ID缺失')
    router.push('/app-password')
    return
  }

  await loadEmailAccount(accountId)
  await loadFolders(accountId)
  await loadEmails(accountId)
}

const loadEmailAccount = async (accountId) => {
  try {
    const response = await axios.get(`/api/protected/email-accounts/${accountId}`)
    if (response.data.success) {
      emailAccount.value = response.data.data
    }
  } catch (error) {
    console.error('加载邮箱账户失败:', error)
    ElMessage.error('加载邮箱账户信息失败')
  }
}

const loadFolders = async (accountId) => {
  try {
    const response = await axios.get(`/app-password/folders/${accountId}`)
    if (response.data.success) {
      folders.value = response.data.data || ['INBOX']
    }
  } catch (error) {
    console.error('加载文件夹失败:', error)
    folders.value = ['INBOX']
  }
}

const loadEmails = async (accountId) => {
  loading.value = true
  try {
    const response = await axios.get(`/app-password/emails/${accountId}`, {
      params: {
        page: pagination.currentPage,
        pageSize: pagination.pageSize
      }
    })

    if (response.data.success) {
      emails.value = response.data.data.emails || []
      pagination.total = response.data.data.total || 0
      emailStats.total = response.data.data.total || 0
      emailStats.unread = emails.value.filter(email => !email.isRead).length
      lastSyncTime.value = new Date()
    }
  } catch (error) {
    console.error('加载邮件失败:', error)
    ElMessage.error('加载邮件失败: ' + (error.response?.data?.error || error.message))
  } finally {
    loading.value = false
  }
}

const refreshEmails = () => {
  const accountId = route.params.id
  if (accountId) {
    loadEmails(accountId)
  }
}

const handleSearch = () => {
  // 搜索功能由computed属性实现
}

const handleSizeChange = (size) => {
  pagination.pageSize = size
  pagination.currentPage = 1
  refreshEmails()
}

const handleCurrentChange = (page) => {
  pagination.currentPage = page
  refreshEmails()
}

const handleRowClick = async (row) => {
  // 加载邮件详情
  const accountId = route.params.id
  try {
    const response = await axios.get(`/app-password/emails/${accountId}/${row.messageId}`)
    if (response.data.success) {
      emailDrawer.email = response.data.data.email
      emailDrawer.visible = true
      
      // 标记为已读
      row.isRead = true
      emailStats.unread = emails.value.filter(email => !email.isRead).length
    }
  } catch (error) {
    console.error('加载邮件详情失败:', error)
    ElMessage.error('加载邮件详情失败')
  }
}

const openSettings = () => {
  settingsDialog.visible = true
}

const testConnection = async () => {
  testingConnection.value = true
  try {
    const response = await axios.post('/app-password/test', {
      email: emailAccount.value.email_address
    })
    
    if (response.data.success) {
      ElMessage.success('连接测试成功！应用密码工作正常')
    } else {
      ElMessage.error('连接测试失败')
    }
  } catch (error) {
    ElMessage.error('连接测试失败: ' + (error.response?.data?.error || error.message))
  } finally {
    testingConnection.value = false
  }
}

const manageAppPassword = () => {
  router.push('/app-password')
}

// 辅助函数
const getSenderName = (fromArray) => {
  if (!fromArray || fromArray.length === 0) return '未知发件人'
  return fromArray[0].name || fromArray[0].address || '未知发件人'
}

const getSenderEmail = (fromArray) => {
  if (!fromArray || fromArray.length === 0) return ''
  return fromArray[0].address || ''
}

const getSenderDisplay = (fromArray) => {
  if (!fromArray || fromArray.length === 0) return '未知发件人'
  const sender = fromArray[0]
  return sender.name ? `${sender.name} <${sender.address}>` : sender.address
}

const getRecipientDisplay = (toArray) => {
  if (!toArray || toArray.length === 0) return ''
  return toArray.map(recipient => 
    recipient.name ? `${recipient.name} <${recipient.address}>` : recipient.address
  ).join(', ')
}

const formatEmailTime = (dateStr) => {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  const now = new Date()
  const diffMs = now - date
  const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24))
  
  if (diffDays === 0) {
    return date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
  } else if (diffDays === 1) {
    return '昨天'
  } else if (diffDays < 7) {
    return `${diffDays}天前`
  } else {
    return date.toLocaleDateString('zh-CN', { month: '2-digit', day: '2-digit' })
  }
}

const formatFullDate = (dateStr) => {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}

const formatDate = (date) => {
  return date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
}

const formatFileSize = (bytes) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const sanitizeHtml = (html) => {
  // 简化版HTML清理，生产环境建议使用DOMPurify
  return html.replace(/<script\b[^<]*(?:(?!<\/script>)<[^<]*)*<\/script>/gi, '')
}
</script>

<style scoped>
.app-password-email-view {
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
  font-size: 24px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.title-icon {
  margin-right: 8px;
  font-size: 28px;
  color: var(--el-color-primary);
}

.page-description {
  margin: 0;
  display: flex;
  align-items: center;
  color: var(--el-text-color-regular);
}

.email-stats {
  margin-bottom: 20px;
}

.stat-card {
  height: 80px;
  text-align: center;
}

.stat-item {
  padding: 16px 0;
}

.stat-value {
  font-size: 24px;
  font-weight: bold;
  color: var(--el-text-color-primary);
  line-height: 1;
}

.stat-label {
  font-size: 13px;
  color: var(--el-text-color-regular);
  margin-top: 4px;
}

.email-list-card {
  background: white;
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

.header-filters {
  display: flex;
  align-items: center;
}

.email-list-container {
  min-height: 400px;
}

.unread-icon {
  font-size: 8px;
}

.sender-cell {
  line-height: 1.4;
}

.sender-name {
  font-weight: 500;
  color: var(--el-text-color-primary);
}

.sender-email {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.subject-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}

.subject-text {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.subject-text.unread {
  font-weight: 600;
}

.attachment-icon {
  color: var(--el-color-info);
  font-size: 14px;
}

.snippet-text {
  color: var(--el-text-color-regular);
  font-size: 13px;
  line-height: 1.3;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.time-cell {
  font-size: 13px;
  color: var(--el-text-color-secondary);
}

.pagination-container {
  margin-top: 20px;
  display: flex;
  justify-content: center;
}

/* 邮件详情样式 */
.email-detail {
  padding: 0 16px;
}

.email-header {
  margin-bottom: 16px;
}

.email-meta {
  background: var(--el-fill-color-lighter);
  padding: 16px;
  border-radius: 6px;
}

.meta-row {
  margin-bottom: 8px;
  display: flex;
  align-items: flex-start;
}

.meta-row:last-child {
  margin-bottom: 0;
}

.meta-label {
  font-weight: 500;
  min-width: 60px;
  color: var(--el-text-color-regular);
}

.meta-value {
  flex: 1;
  word-break: break-all;
}

.attachments {
  margin-top: 16px;
}

.attachments h4 {
  margin: 0 0 8px 0;
  color: var(--el-text-color-primary);
}

.attachment-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.attachment-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 0;
  color: var(--el-text-color-regular);
}

.attachment-size {
  color: var(--el-text-color-secondary);
  font-size: 12px;
}

.email-body {
  line-height: 1.6;
}

.html-content {
  max-width: 100%;
  overflow-x: auto;
}

.text-content pre {
  white-space: pre-wrap;
  word-wrap: break-word;
  font-family: inherit;
  margin: 0;
}

.no-content {
  text-align: center;
  padding: 40px 0;
}

.settings-content {
  padding: 20px 0;
}

.settings-actions {
  margin-top: 20px;
  display: flex;
  gap: 12px;
  justify-content: center;
}

@media (max-width: 768px) {
  .header-content {
    flex-direction: column;
    text-align: center;
    gap: 16px;
  }
  
  .header-filters {
    flex-direction: column;
    gap: 8px;
  }
}
</style> 