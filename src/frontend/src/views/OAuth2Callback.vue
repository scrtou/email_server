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
    
    const handleCallback = async () => {
      try {
        const token = route.query.token
        const errorParam = route.query.error
        
        if (errorParam) {
          throw new Error(getErrorMessage(errorParam))
        }
        
        if (!token) {
          throw new Error('未收到登录凭证，请重试')
        }
        
        // 保存token到store
        authStore.setToken(token)
        
        // 获取用户信息
        await authStore.fetchUserProfile()

        ElMessage.success('LinuxDo登录成功！')

        // 直接跳转到主页
        router.push('/')
        
      } catch (err) {
        console.error('OAuth2回调处理失败:', err)
        loading.value = false
        error.value = true
        errorMessage.value = err.message || '登录处理失败，请重试'
        ElMessage.error(errorMessage.value)
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
