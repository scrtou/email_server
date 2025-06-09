<template>
  <div class="oauth2-callback">
    <div class="callback-container">
      <div v-if="loading" class="loading-state">
        <el-icon class="loading-icon" :size="40">
          <Loading />
        </el-icon>
        <h2>正在处理登录...</h2>
        <p>请稍候，我们正在完成您的LinuxDo登录</p>
      </div>
      
      <div v-else-if="error" class="error-state">
        <el-icon class="error-icon" :size="40">
          <CircleClose />
        </el-icon>
        <h2>登录失败</h2>
        <p>{{ errorMessage }}</p>
        <el-button type="primary" @click="goToLogin">返回登录</el-button>
      </div>
      

    </div>
  </div>
</template>

<script>
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { ElMessage } from 'element-plus'
import { Loading, CircleClose } from '@element-plus/icons-vue'

export default {
  name: 'OAuth2Callback',
  components: {
    Loading,
    CircleClose
  },
  setup() {
    const router = useRouter()
    const route = useRoute()
    const authStore = useAuthStore()

    const loading = ref(true)
    const error = ref(false)
    const errorMessage = ref('')
    const processed = ref(false) // 防止重复处理

    const handleCallback = async () => {
      // 使用全局标志防止重复处理（跨组件实例）
      if (window.oauth2CallbackProcessing) {
        console.log('OAuth2Callback: 全局处理中，跳过重复处理')
        return
      }

      // 防止重复处理
      if (processed.value) {
        console.log('OAuth2Callback: 已处理过，跳过重复处理')
        return
      }

      window.oauth2CallbackProcessing = true
      processed.value = true
      console.log('OAuth2Callback: 开始处理回调')

      try {
        const token = route.query.token
        const errorParam = route.query.error

        if (errorParam) {
          throw new Error(getErrorMessage(errorParam))
        }

        if (!token) {
          throw new Error('未收到登录凭证，请重试')
        }

        console.log('OAuth2Callback: 保存token到store')
        // 保存token到store
        authStore.setToken(token)

        console.log('OAuth2Callback: 获取用户信息')
        // 获取用户信息
        await authStore.fetchUserProfile()

        console.log('OAuth2Callback: 显示成功消息')
        // 使用全局标志防止重复显示
        if (!window.oauth2LoginMessageShown) {
          window.oauth2LoginMessageShown = true
          ElMessage.success('OAuth2登录成功！')
          // 5秒后重置标志
          setTimeout(() => {
            window.oauth2LoginMessageShown = false
          }, 5000)
        }

        console.log('OAuth2Callback: 跳转到主页')
        // 直接跳转到主页
        router.push('/')

      } catch (err) {
        console.error('OAuth2回调处理失败:', err)
        loading.value = false
        error.value = true
        errorMessage.value = err.message || '登录处理失败，请重试'
        ElMessage.error(errorMessage.value)
      } finally {
        // 重置全局标志
        setTimeout(() => {
          window.oauth2CallbackProcessing = false
        }, 2000)
      }
    }
    
    const getErrorMessage = (errorCode) => {
      const errorMessages = {
        'token_generation_failed': '生成登录凭证失败',
        'user_creation_failed': '用户账号创建失败',
        'access_denied': '您拒绝了授权请求',
        'invalid_request': '无效的登录请求',
        'server_error': '服务器错误，请稍后重试'
      }
      return errorMessages[errorCode] || '未知错误，请重试'
    }
    
    const goToLogin = () => {
      router.push('/login')
    }
    
    onMounted(() => {
      handleCallback()
    })
    
    return {
      loading,
      error,
      errorMessage,
      goToLogin
    }
  }
}
</script>

<style scoped>
.oauth2-callback {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: var(--space-4);
}

.callback-container {
  background: white;
  border-radius: var(--radius-xl);
  padding: var(--space-12) var(--space-8);
  box-shadow: var(--shadow-2xl);
  text-align: center;
  max-width: 400px;
  width: 100%;
}

.loading-state,
.error-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-4);
}

.loading-icon {
  color: var(--color-blue-500);
  animation: spin 1s linear infinite;
}

.error-icon {
  color: var(--color-red-500);
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

h2 {
  margin: 0;
  font-size: var(--text-2xl);
  font-weight: var(--font-bold);
  color: var(--color-gray-800);
}

p {
  margin: 0;
  color: var(--color-gray-600);
  font-size: var(--text-base);
  line-height: 1.5;
}

.el-button {
  margin-top: var(--space-4);
}
</style>
