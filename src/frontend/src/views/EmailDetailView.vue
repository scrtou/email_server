<template>
  <div class="email-detail-view">
    <!-- 顶部导航栏 -->
    <div class="email-header-bar">
      <div class="header-left">
        <el-button
          :icon="ArrowLeft"
          @click="goBack"
          type="primary"
          plain
          size="default"
        >
          返回收件箱
        </el-button>
      </div>

      <div class="header-center">
        <h2 class="email-title">{{ email ? email.subject || '(无主题)' : '加载中...' }}</h2>
      </div>

      <div class="header-right">
        <el-button-group v-if="email">
          <el-button :icon="email.isRead ? Message : MessageBox" @click="toggleRead" title="标记为已读/未读" />
          <el-button :icon="Star" @click="toggleStar" title="标记星标" />
          <el-button :icon="Delete" @click="deleteEmail" title="删除邮件" />
        </el-button-group>
      </div>
    </div>

    <!-- 邮件内容区域 - 可滚动 -->
    <div class="email-content-container" v-if="email">
      <div class="email-content-wrapper">
        <!-- 邮件元信息卡片 -->
        <el-card class="email-meta-card" shadow="never">
          <div class="email-meta-header">
            <div class="sender-info">
              <div class="sender-avatar">
                {{ getInitials(email.from?.[0]?.name || email.from?.[0]?.address) }}
              </div>
              <div class="sender-details">
                <div class="sender-name">{{ email.from?.[0]?.name || email.from?.[0]?.address }}</div>
                <div class="sender-email" v-if="email.from?.[0]?.name">{{ email.from?.[0]?.address }}</div>
              </div>
            </div>
            <div class="email-date">
              <el-icon><Clock /></el-icon>
              {{ formattedDate }}
            </div>
          </div>

          <!-- 收件人信息 -->
          <div class="recipients-section" v-if="toText || ccText">
            <el-collapse>
              <el-collapse-item name="recipients">
                <template #title>
                  <span class="recipients-title">
                    <el-icon><User /></el-icon>
                    收件人详情
                  </span>
                </template>
                <div class="recipients-content">
                  <div class="recipient-row" v-if="toText">
                    <span class="recipient-label">收件人:</span>
                    <span class="recipient-list">{{ toText }}</span>
                  </div>
                  <div class="recipient-row" v-if="ccText">
                    <span class="recipient-label">抄送:</span>
                    <span class="recipient-list">{{ ccText }}</span>
                  </div>
                </div>
              </el-collapse-item>
            </el-collapse>
          </div>
        </el-card>

        <!-- 邮件正文卡片 -->
        <el-card class="email-body-card" shadow="never">
          <div class="email-body-content">
            <div v-if="email.htmlBody" v-html="sanitizedHtml" class="email-html-body"></div>
            <div v-else class="email-text-body">
              <pre>{{ email.body || '邮件内容为空' }}</pre>
            </div>
          </div>
        </el-card>

        <!-- 附件卡片 -->
        <el-card v-if="email.attachments && email.attachments.length > 0" class="email-attachments-card" shadow="never">
          <template #header>
            <div class="attachments-header">
              <el-icon><Paperclip /></el-icon>
              <span>附件 ({{ email.attachments.length }})</span>
            </div>
          </template>
          <attachment-list :attachments="email.attachments" />
        </el-card>
      </div>
    </div>

    <!-- 加载状态 -->
    <div v-else class="loading-container" v-loading="true">
      <div class="loading-content">
        <el-icon class="loading-icon" :size="48">
          <Loading />
        </el-icon>
        <p class="loading-text">正在加载邮件详情...</p>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, computed, onMounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useInboxStore } from '@/stores/inbox';
import AttachmentList from '@/components/AttachmentList.vue';
import DOMPurify from 'dompurify';
import { format } from 'date-fns';
import { markEmailAsRead } from '@/utils/api';
import {
  ArrowLeft,
  Message,
  MessageBox,
  Star,
  Delete,
  Clock,
  User,
  Paperclip,
  Loading
} from '@element-plus/icons-vue';

export default {
  name: 'EmailDetailView',
  components: {
    AttachmentList,
  },
  setup() {
    const route = useRoute();
    const router = useRouter();
    const store = useInboxStore();
    const email = ref(null);

    const formatAddresses = (addresses) => {
      if (!addresses || addresses.length === 0) return '';
      return addresses.map(addr => addr.name ? `${addr.name} <${addr.address}>` : addr.address).join(', ');
    };

    const fromText = computed(() => formatAddresses(email.value?.from));
    const toText = computed(() => formatAddresses(email.value?.to));
    const ccText = computed(() => formatAddresses(email.value?.cc));

    const formattedDate = computed(() => {
      if (!email.value?.date) return '';
      return format(new Date(email.value.date), 'yyyy-MM-dd HH:mm:ss');
    });

    const sanitizedHtml = computed(() => {
      if (!email.value?.htmlBody) return '';
      return DOMPurify.sanitize(email.value.htmlBody);
    });

    const goBack = () => {
      router.push({ name: 'Inbox' });
    };

    const getInitials = (name) => {
      if (!name) return '?';
      const words = name.split(' ');
      if (words.length >= 2) {
        return (words[0][0] + words[1][0]).toUpperCase();
      }
      return name.substring(0, 2).toUpperCase();
    };

    const toggleRead = () => {
      // TODO: 实现标记已读/未读功能
      console.log('Toggle read status');
    };

    const toggleStar = () => {
      // TODO: 实现星标功能
      console.log('Toggle star');
    };

    const deleteEmail = () => {
      // TODO: 实现删除功能
      console.log('Delete email');
    };



    onMounted(async () => {
      const emailId = route.params.id;

      try {
        // First try to get from local store
        let fetchedEmail = store.getEmailById(emailId);

        if (!fetchedEmail) {
          // If not found locally, try to fetch emails first
          if (!store.selectedAccountId) {
            // If no account is selected, we can't fetch emails
            console.error('No account selected for fetching emails');
            return;
          }
          await store.fetchEmails();
          fetchedEmail = store.getEmailById(emailId);
        }

        // If we still don't have the email or it doesn't have body content, fetch detail
        if (!fetchedEmail || (!fetchedEmail.body && !fetchedEmail.htmlBody)) {
          fetchedEmail = await store.fetchEmailDetail(emailId);
        }

        email.value = fetchedEmail;

        // 如果邮件未读，自动标记为已读
        if (fetchedEmail && !fetchedEmail.isRead && store.selectedAccountId) {
          try {
            await markEmailAsRead(emailId, { account_id: store.selectedAccountId });
            console.log('Email marked as read automatically');
            // 更新本地状态
            email.value.isRead = true;
            // 刷新邮件列表以更新状态
            if (store.emails.length > 0) {
              const emailIndex = store.emails.findIndex(e => e.messageId === emailId);
              if (emailIndex !== -1) {
                store.emails[emailIndex].isRead = true;
              }
            }
          } catch (error) {
            console.error('Failed to mark email as read:', error);
            // 不显示错误消息，因为这是自动操作
          }
        }
      } catch (error) {
        console.error('Error loading email detail:', error);
        // You might want to show an error message to the user here
      }
    });

    return {
      email,
      fromText,
      toText,
      ccText,
      formattedDate,
      sanitizedHtml,
      goBack,
      getInitials,
      toggleRead,
      toggleStar,
      deleteEmail,
      ArrowLeft,
      Message,
      MessageBox,
      Star,
      Delete,
      Clock,
      User,
      Paperclip,
      Loading,
    };
  },
};
</script>

<style scoped>
.email-detail-view {
  height: 100vh;
  display: flex;
  flex-direction: column;
  background: #f5f7fa;
}

/* 顶部导航栏 */
.email-header-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 24px;
  background: #ffffff;
  border-bottom: 1px solid #e4e7ed;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  flex-shrink: 0;
  z-index: 10;
}

.header-left {
  flex: 0 0 auto;
}

.header-center {
  flex: 1;
  text-align: center;
  margin: 0 24px;
}

.email-title {
  font-size: 18px;
  font-weight: 600;
  color: #303133;
  margin: 0;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.header-right {
  flex: 0 0 auto;
}

/* 邮件内容容器 - 可滚动 */
.email-content-container {
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
  padding: 0;
}

.email-content-wrapper {
  max-width: 900px;
  margin: 0 auto;
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

/* 邮件元信息卡片 */
.email-meta-card {
  border-radius: 12px;
  border: 1px solid #e4e7ed;
}

.email-meta-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.sender-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.sender-avatar {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  background: linear-gradient(135deg, #409eff, #67c23a);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  font-size: 16px;
  flex-shrink: 0;
}

.sender-details {
  flex: 1;
  min-width: 0;
}

.sender-name {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 2px;
}

.sender-email {
  font-size: 14px;
  color: #909399;
}

.email-date {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 14px;
  color: #606266;
  flex-shrink: 0;
}

/* 收件人信息 */
.recipients-section {
  margin-top: 16px;
}

.recipients-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  color: #606266;
}

.recipients-content {
  padding: 12px 0;
}

.recipient-row {
  display: flex;
  margin-bottom: 8px;
  gap: 12px;
}

.recipient-label {
  font-weight: 600;
  color: #606266;
  min-width: 60px;
  flex-shrink: 0;
}

.recipient-list {
  color: #303133;
  word-break: break-all;
}

/* 邮件正文卡片 */
.email-body-card {
  border-radius: 12px;
  border: 1px solid #e4e7ed;
}

.email-body-content {
  line-height: 1.6;
}

.email-html-body {
  color: #303133;
  word-wrap: break-word;
  overflow-wrap: break-word;
}

.email-html-body :deep(img) {
  max-width: 100%;
  height: auto;
  border-radius: 8px;
}

.email-html-body :deep(a) {
  color: #409eff;
  text-decoration: none;
}

.email-html-body :deep(a:hover) {
  text-decoration: underline;
}

.email-text-body {
  color: #303133;
}

.email-text-body pre {
  white-space: pre-wrap;
  word-wrap: break-word;
  font-family: inherit;
  margin: 0;
  background: #f8f9fa;
  padding: 16px;
  border-radius: 8px;
  border: 1px solid #e9ecef;
}

/* 附件卡片 */
.email-attachments-card {
  border-radius: 12px;
  border: 1px solid #e4e7ed;
}

.attachments-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
  color: #303133;
}

/* 加载状态 */
.loading-container {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f5f7fa;
}

.loading-content {
  text-align: center;
}

.loading-icon {
  color: #409eff;
  margin-bottom: 16px;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.loading-text {
  font-size: 16px;
  color: #606266;
  margin: 0;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .email-header-bar {
    padding: 12px 16px;
    flex-wrap: wrap;
    gap: 12px;
  }

  .header-center {
    order: 3;
    flex-basis: 100%;
    margin: 8px 0 0 0;
    text-align: left;
  }

  .email-title {
    font-size: 16px;
  }

  .email-content-wrapper {
    padding: 16px;
  }

  .email-meta-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }

  .sender-avatar {
    width: 40px;
    height: 40px;
    font-size: 14px;
  }

  .recipient-row {
    flex-direction: column;
    gap: 4px;
  }

  .recipient-label {
    min-width: auto;
  }
}

/* 滚动条样式 */
.email-content-container::-webkit-scrollbar {
  width: 6px;
}

.email-content-container::-webkit-scrollbar-track {
  background: #f1f1f1;
  border-radius: 3px;
}

.email-content-container::-webkit-scrollbar-thumb {
  background: #c1c1c1;
  border-radius: 3px;
}

.email-content-container::-webkit-scrollbar-thumb:hover {
  background: #a8a8a8;
}
</style>