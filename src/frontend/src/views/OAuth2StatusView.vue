<template>
  <div class="oauth2-status-view">
    <div class="page-header">
      <h1>OAuth2令牌状态管理</h1>
      <p class="description">
        管理和监控您的Google、Microsoft等OAuth2邮箱账户的授权状态
      </p>
    </div>

    <!-- 令牌状态组件 -->
    <OAuth2TokenStatus />

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

        <el-collapse-item title="为什么令牌会过期？" name="2">
          <p>令牌过期是正常的安全机制，可能的原因包括：</p>
          <ul>
            <li>访问令牌自然过期（通常1小时）</li>
            <li>用户更改了邮箱密码</li>
            <li>用户在邮箱提供商处撤销了应用授权</li>
            <li>长期未使用导致刷新令牌过期</li>
          </ul>
        </el-collapse-item>

        <el-collapse-item title="如何处理令牌过期？" name="3">
          <p>系统会自动处理大部分情况：</p>
          <ul>
            <li><strong>自动刷新</strong>：访问令牌过期时自动使用刷新令牌获取新令牌</li>
            <li><strong>重新授权</strong>：当刷新令牌也失效时，需要手动重新授权</li>
            <li><strong>状态监控</strong>：定期检查令牌状态，及时发现问题</li>
          </ul>
        </el-collapse-item>

        <el-collapse-item title="令牌状态说明" name="4">
          <div class="status-explanation">
            <div class="status-item">
              <el-tag type="success" size="small">正常</el-tag>
              <span>令牌有效，可以正常访问邮件</span>
            </div>
            <div class="status-item">
              <el-tag type="danger" size="small">已过期</el-tag>
              <span>刷新令牌失效，需要重新授权</span>
            </div>
            <div class="status-item">
              <el-tag type="warning" size="small">错误</el-tag>
              <span>令牌验证失败，可能需要重新授权</span>
            </div>
          </div>
        </el-collapse-item>

        <el-collapse-item title="安全建议" name="5">
          <ul>
            <li>定期检查令牌状态，确保邮箱连接正常</li>
            <li>如果长期不使用某个邮箱，建议及时移除授权</li>
            <li>发现异常访问时，立即撤销授权并重新连接</li>
            <li>不要在不安全的网络环境下进行OAuth2授权</li>
          </ul>
        </el-collapse-item>
      </el-collapse>
    </el-card>

    <!-- 操作按钮 -->
    <div class="action-buttons">
      <el-button type="primary" @click="goToEmailAccounts">
        <el-icon><Setting /></el-icon>
        管理邮箱账户
      </el-button>
      <el-button @click="goToInbox">
        <el-icon><Message /></el-icon>
        返回收件箱
      </el-button>
    </div>
  </div>
</template>

<script setup>
import { useRouter } from 'vue-router'
import { QuestionFilled, Setting, Message } from '@element-plus/icons-vue'
import OAuth2TokenStatus from '@/components/OAuth2TokenStatus.vue'

const router = useRouter()

// 导航到邮箱账户管理页面
const goToEmailAccounts = () => {
  router.push('/email-accounts')
}

// 导航到收件箱
const goToInbox = () => {
  router.push('/inbox')
}
</script>

<style scoped>
.oauth2-status-view {
  max-width: 1000px;
  margin: 0 auto;
  padding: 20px;
}

.page-header {
  text-align: center;
  margin-bottom: 30px;
}

.page-header h1 {
  color: #303133;
  margin-bottom: 10px;
}

.description {
  color: #606266;
  font-size: 16px;
  margin: 0;
}

.help-card {
  margin: 30px 0;
}

.help-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
}

.help-card ul {
  margin: 10px 0;
  padding-left: 20px;
}

.help-card li {
  margin: 5px 0;
  line-height: 1.6;
}

.status-explanation {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.status-explanation .status-item {
  display: flex;
  align-items: center;
  gap: 12px;
}

.action-buttons {
  display: flex;
  justify-content: center;
  gap: 16px;
  margin-top: 30px;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .oauth2-status-view {
    padding: 10px;
  }
  
  .action-buttons {
    flex-direction: column;
    align-items: center;
  }
  
  .action-buttons .el-button {
    width: 200px;
  }
}
</style>
