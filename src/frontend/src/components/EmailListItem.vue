<template>
  <div class="email-list-item" @click="goToEmail" :class="{ 'unread': !email.isRead }">
    <div class="email-content">
      <!-- 左侧状态指示器 -->
      <div class="status-indicator" :class="{ 'unread': !email.isRead }"></div>

      <!-- 主要内容区域 -->
      <div class="email-main">
        <div class="email-header">
          <div class="sender-info">
            <span class="sender-name" :class="{ 'is-read': email.isRead }">{{ senderName }}</span>
            <span class="sender-email" v-if="senderEmail !== senderName">{{ senderEmail }}</span>
          </div>
          <div class="email-meta">
            <span class="date">{{ formattedDate }}</span>
            <el-icon v-if="email.hasAttachment" class="attachment-icon" title="有附件">
              <Paperclip />
            </el-icon>
          </div>
        </div>

        <div class="subject-line" :class="{ 'is-read': email.isRead }">
          {{ email.subject || '(无主题)' }}
        </div>

        <div class="snippet-line" v-if="email.snippet">
          {{ email.snippet }}
        </div>
      </div>
    </div>

    <!-- 悬停时显示的操作按钮 -->
    <div class="email-actions">
      <el-button-group size="small">
        <el-button :icon="View" @click.stop="goToEmail" title="查看详情" />
        <el-button :icon="email.isRead ? Message : MessageBox" @click.stop="toggleRead" title="标记为已读/未读" />
      </el-button-group>
    </div>
  </div>
</template>

<script>
import { computed } from 'vue';
import { useRouter } from 'vue-router';
import { format } from 'date-fns';
import { Paperclip, View, Message, MessageBox } from '@element-plus/icons-vue';

export default {
  name: 'EmailListItem',
  components: {
    Paperclip,
  },
  props: {
    email: {
      type: Object,
      required: true,
    },
  },
  setup(props) {
    const router = useRouter();

    const senderName = computed(() => {
      if (!props.email.from || props.email.from.length === 0) {
        return 'Unknown Sender';
      }
      const firstSender = props.email.from[0];
      return firstSender.name || firstSender.address;
    });

    const senderEmail = computed(() => {
      if (!props.email.from || props.email.from.length === 0) {
        return '';
      }
      return props.email.from[0].address;
    });

    const formattedDate = computed(() => {
      if (!props.email.date) return '';
      const emailDate = new Date(props.email.date);
      const now = new Date();
      const diffInHours = (now - emailDate) / (1000 * 60 * 60);

      if (diffInHours < 24) {
        return format(emailDate, 'HH:mm');
      } else if (diffInHours < 24 * 7) {
        return format(emailDate, 'MM-dd');
      } else {
        return format(emailDate, 'yyyy-MM-dd');
      }
    });

    const goToEmail = () => {
      router.push({ name: 'EmailDetail', params: { id: props.email.messageId } });
    };

    const toggleRead = () => {
      // TODO: 实现标记已读/未读功能
      console.log('Toggle read status for email:', props.email.messageId);
    };

    return {
      senderName,
      senderEmail,
      formattedDate,
      goToEmail,
      toggleRead,
      Paperclip,
      View,
      Message,
      MessageBox,
    };
  },
};
</script>

<style scoped>
.email-list-item {
  position: relative;
  background: #ffffff;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  margin-bottom: 8px;
  cursor: pointer;
  transition: all 0.3s ease;
  overflow: hidden;
}

.email-list-item:hover {
  border-color: #409eff;
  box-shadow: 0 4px 12px rgba(64, 158, 255, 0.15);
  transform: translateY(-1px);
}

.email-list-item.unread {
  background: linear-gradient(135deg, #f8f9ff 0%, #ffffff 100%);
  border-left: 4px solid #409eff;
}

.email-content {
  display: flex;
  align-items: flex-start;
  padding: 16px;
  gap: 12px;
}

.status-indicator {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #dcdfe6;
  margin-top: 6px;
  flex-shrink: 0;
}

.status-indicator.unread {
  background: #409eff;
  box-shadow: 0 0 8px rgba(64, 158, 255, 0.4);
}

.email-main {
  flex: 1;
  min-width: 0;
}

.email-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 8px;
  gap: 12px;
}

.sender-info {
  flex: 1;
  min-width: 0;
}

.sender-name {
  font-weight: 600;
  font-size: 14px;
  color: #303133;
  display: block;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.sender-name.is-read {
  font-weight: 500;
  color: #606266;
}

.sender-email {
  font-size: 12px;
  color: #909399;
  display: block;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  margin-top: 2px;
}

.email-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.date {
  font-size: 12px;
  color: #909399;
  white-space: nowrap;
}

.attachment-icon {
  color: #909399;
  font-size: 14px;
}

.subject-line {
  font-weight: 600;
  font-size: 15px;
  color: #303133;
  margin-bottom: 6px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  line-height: 1.4;
}

.subject-line.is-read {
  font-weight: 500;
  color: #606266;
}

.snippet-line {
  font-size: 13px;
  color: #909399;
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  text-overflow: ellipsis;
}

.email-actions {
  position: absolute;
  top: 50%;
  right: 16px;
  transform: translateY(-50%);
  opacity: 0;
  transition: opacity 0.3s ease;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(4px);
  border-radius: 6px;
  padding: 4px;
}

.email-list-item:hover .email-actions {
  opacity: 1;
}

.email-actions .el-button {
  border: none;
  background: transparent;
  color: #606266;
  padding: 6px;
}

.email-actions .el-button:hover {
  background: #f5f7fa;
  color: #409eff;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .email-content {
    padding: 12px;
  }

  .email-header {
    flex-direction: column;
    gap: 4px;
  }

  .email-meta {
    align-self: flex-end;
  }

  .email-actions {
    position: static;
    opacity: 1;
    transform: none;
    margin-top: 8px;
    background: transparent;
    backdrop-filter: none;
  }
}
</style>