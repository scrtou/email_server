<template>
    <div class="login-container">
      <el-card class="login-card">
        <template #header>
          <div class="login-header">
            <h2>账号管理系统</h2>
          </div>
        </template>
        
        <el-form
          ref="loginFormRef"
          :model="loginForm"
          :rules="loginRules"
          class="login-form"
          label-width="0"
          @submit.prevent="handleLogin"
        >
          <el-form-item prop="username">
            <el-input
              v-model="loginForm.username"
              placeholder="用户名"
              size="large"
              prefix-icon="User"
              clearable
            />
          </el-form-item>
          
          <el-form-item prop="password">
            <el-input
              v-model="loginForm.password"
              type="password"
              placeholder="密码"
              size="large"
              prefix-icon="Lock"
              show-password
              @keyup.enter="handleLogin"
            />
          </el-form-item>
          
          <el-form-item>
            <el-button
              type="primary"
              size="large"
              style="width: 100%"
              :loading="authStore.isLoading"
              @click="handleLogin"
            >
              登录
            </el-button>
          </el-form-item>
        </el-form>

        <!-- 第三方登录 -->
        <div class="third-party-login">
          <div class="third-party-title">
          </div>

          <div class="third-party-icons">
            <div
              class="oauth-icon-button linuxdo-icon"
              :class="{ 'loading': oauthLoading }"
              @click="handleLinuxDoLogin"
              title="使用 Linux.do 登录"
            >
              <div class="icon-wrapper">
                <div v-if="!oauthLoading" class="linuxdo-logo">
                  <div class="logo-top"></div>
                  <div class="logo-middle"></div>
                  <div class="logo-bottom"></div>
                </div>
                <el-icon v-else class="loading-icon" :size="20">
                  <Loading />
                </el-icon>
              </div>
            </div>

            <div
              class="oauth-icon-button google-icon"
              :class="{ 'loading': googleOauthLoading }"
              @click="handleGoogleLogin"
              title="使用 Google 登录"
            >
              <div class="icon-wrapper">
                <div v-if="!googleOauthLoading" class="google-logo">
                  <svg viewBox="0 0 24 24" width="24" height="24">
                    <path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"/>
                    <path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/>
                    <path fill="#FBBC05" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"/>
                    <path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"/>
                  </svg>
                </div>
                <el-icon v-else class="loading-icon" :size="20">
                  <Loading />
                </el-icon>
              </div>
            </div>

            <div
              class="oauth-icon-button microsoft-icon"
              :class="{ 'loading': microsoftOauthLoading }"
              @click="handleMicrosoftLogin"
              title="使用 Microsoft 登录"
            >
              <div class="icon-wrapper">
                <div v-if="!microsoftOauthLoading" class="microsoft-logo">
                  <svg viewBox="0 0 24 24" width="24" height="24">
                    <path fill="#f25022" d="M1 1h10v10H1z"/>
                    <path fill="#00a4ef" d="M13 1h10v10H13z"/>
                    <path fill="#7fba00" d="M1 13h10v10H1z"/>
                    <path fill="#ffb900" d="M13 13h10v10H13z"/>
                  </svg>
                </div>
                <el-icon v-else class="loading-icon" :size="20">
                  <Loading />
                </el-icon>
              </div>
            </div>
          </div>
        </div>

        <div class="login-footer">
          <p>
            还没有账户？
            <router-link to="/register" class="register-link">立即注册</router-link>
          </p>
        </div>
      </el-card>
    </div>
  </template>
  
  <script>
  import { ref, reactive, onMounted } from 'vue'
  import { useRouter, useRoute } from 'vue-router'
  import { useAuthStore } from '@/stores/auth'
  import { ElMessage } from 'element-plus'
  import { Loading } from '@element-plus/icons-vue'
  import api from '@/utils/api'

  export default {
    name: 'LoginPage',
    components: {
      Loading
    },
    setup() {
      const router = useRouter()
      const route = useRoute()
      const authStore = useAuthStore()
      const loginFormRef = ref()
      const oauthLoading = ref(false)
      const googleOauthLoading = ref(false)
      const microsoftOauthLoading = ref(false)

      const loginForm = reactive({
        username: '',
        password: ''
      })

      const loginRules = {
        username: [
          { required: true, message: '请输入用户名', trigger: 'blur' }
        ],
        password: [
          { required: true, message: '请输入密码', trigger: 'blur' },
          { min: 6, message: '密码长度不能少于6位', trigger: 'blur' }
        ]
      }

      const handleLogin = async () => {
        if (!loginFormRef.value) return

        try {
          const valid = await loginFormRef.value.validate()
          if (!valid) return

          const success = await authStore.login(loginForm)
          if (success) {
            router.push('/')
          }
        } catch (error) {
          console.error('登录表单验证失败:', error)
        }
      }

      const handleLinuxDoLogin = async () => {
        try {
          oauthLoading.value = true
          console.log('开始LinuxDo OAuth2登录...')

          const response = await api.get('/auth/oauth2/linuxdo/login')
          console.log('OAuth2登录响应:', response)

          if (response && response.auth_url) {
            console.log('跳转到授权页面:', response.auth_url)
            // 跳转到LinuxDo授权页面
            window.location.href = response.auth_url
          } else {
            console.error('响应格式错误:', response)
            ElMessage.error('获取授权链接失败')
          }
        } catch (error) {
          console.error('LinuxDo登录失败:', error)
          console.error('错误详情:', error.response?.data || error.message)
          ElMessage.error(`LinuxDo登录失败: ${error.message}`)
        } finally {
          oauthLoading.value = false
        }
      }

      const handleGoogleLogin = async () => {
        try {
          googleOauthLoading.value = true
          console.log('开始Google OAuth2登录...')

          const response = await api.get('/auth/oauth2/google/login')
          console.log('Google OAuth2登录响应:', response)

          if (response && response.auth_url) {
            console.log('跳转到Google授权页面:', response.auth_url)
            // 跳转到Google授权页面
            window.location.href = response.auth_url
          } else {
            console.error('响应格式错误:', response)
            ElMessage.error('获取Google授权链接失败')
          }
        } catch (error) {
          console.error('Google登录失败:', error)
          console.error('错误详情:', error.response?.data || error.message)
          ElMessage.error(`Google登录失败: ${error.message}`)
        } finally {
          googleOauthLoading.value = false
        }
      }

      const handleMicrosoftLogin = async () => {
        try {
          microsoftOauthLoading.value = true
          console.log('开始Microsoft OAuth2登录...')

          const response = await api.get('/auth/oauth2/microsoft/login')
          console.log('Microsoft OAuth2登录响应:', response)

          if (response && response.auth_url) {
            console.log('跳转到Microsoft授权页面:', response.auth_url)
            // 跳转到Microsoft授权页面
            window.location.href = response.auth_url
          } else {
            console.error('响应格式错误:', response)
            ElMessage.error('获取Microsoft授权链接失败')
          }
        } catch (error) {
          console.error('Microsoft登录失败:', error)
          console.error('错误详情:', error.response?.data || error.message)
          ElMessage.error(`Microsoft登录失败: ${error.message}`)
        } finally {
          microsoftOauthLoading.value = false
        }
      }

      // 处理OAuth2回调 - 已移除，现在由专门的OAuth2Callback页面处理

      onMounted(() => {
        // OAuth2回调现在由专门的OAuth2Callback页面处理
        // 检查是否有错误参数
        if (route.query.error) {
          const errorMessages = {
            'invalid_state': 'OAuth2状态验证失败，请重新登录',
            'token_exchange_failed': '获取访问令牌失败，请重试',
            'user_info_failed': '获取用户信息失败，请重试',
            'user_creation_failed': '用户账号创建失败，请重试',
            'token_generation_failed': '生成登录凭证失败，请重试'
          }
          const errorMessage = errorMessages[route.query.error] || '登录过程中发生错误，请重试'
          ElMessage.error(errorMessage)

          // 清除URL中的错误参数
          router.replace('/auth/login')
        }
      })

      return {
        loginFormRef,
        loginForm,
        loginRules,
        authStore,
        oauthLoading,
        googleOauthLoading,
        microsoftOauthLoading,
        handleLogin,
        handleLinuxDoLogin,
        handleGoogleLogin,
        handleMicrosoftLogin
      }
    }
  }
  </script>
  
  <style scoped>
  .login-container {
    min-height: 100vh; /* Changed from 100vh to fill parent */
    width: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    background: linear-gradient(135deg, var(--color-primary-600) 0%, var(--color-primary-800) 50%, var(--color-gray-800) 100%);
    padding: var(--space-6); /* Removed padding to allow background to fill edges */
    position: relative;
    overflow: hidden;
  }

  .login-container::before {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background:
      radial-gradient(circle at 20% 20%, rgba(255, 255, 255, 0.1) 0%, transparent 50%),
      radial-gradient(circle at 80% 80%, rgba(255, 255, 255, 0.05) 0%, transparent 50%);
    animation: float 6s ease-in-out infinite;
  }

  @keyframes float {
    0%, 100% { transform: translateY(0px); }
    50% { transform: translateY(-10px); }
  }

  .login-card {
    width: 100%;
    max-width: 420px;
    border-radius: var(--radius-xl);
    box-shadow: var(--shadow-xl);
    background: rgba(255, 255, 255, 0.95);
    backdrop-filter: blur(20px);
    border: 1px solid rgba(255, 255, 255, 0.2);
    position: relative;
    z-index: 1;
    overflow: hidden;
  }

  .login-card::before {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    height: 4px;
    background: linear-gradient(90deg, var(--color-primary-500), var(--color-success-500), var(--color-warning-500));
  }

  .login-header {
    text-align: center;
    margin-bottom: var(--space-6);
    padding-top: var(--space-4);
  }

  .login-header h2 {
    color: var(--color-gray-800);
    margin-bottom: var(--space-2);
    font-size: var(--text-2xl);
    font-weight: var(--font-bold);
    background: linear-gradient(135deg, var(--color-primary-600), var(--color-primary-500));
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
    background-clip: text;
  }

  .login-header p {
    color: var(--color-gray-600);
    margin: 0;
    font-size: var(--text-sm);
    font-weight: var(--font-medium);
  }

  .login-form {
    padding: 0 var(--space-6);
  }

  /* Enhanced form input styling */
  :deep(.el-form-item) {
    margin-bottom: var(--space-5);
  }

  :deep(.el-input__wrapper) {
    border-radius: var(--radius-lg);
    border: 2px solid var(--color-gray-200);
    background: rgba(255, 255, 255, 0.8);
    backdrop-filter: blur(10px);
    transition: all var(--transition-base);
    box-shadow: var(--shadow-sm);
  }

  :deep(.el-input__wrapper:hover) {
    border-color: var(--color-primary-300);
    background: rgba(255, 255, 255, 0.9);
    box-shadow: var(--shadow-md);
  }

  :deep(.el-input__wrapper.is-focus) {
    border-color: var(--color-primary-500);
    background: rgba(255, 255, 255, 1);
    box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
  }

  :deep(.el-input__inner) {
    font-weight: var(--font-medium);
    color: var(--color-gray-700);
  }

  :deep(.el-input__prefix-inner) {
    color: var(--color-gray-500);
  }

  /* Enhanced button styling */
  :deep(.el-button--primary) {
    background: linear-gradient(135deg, var(--color-primary-500), var(--color-primary-600));
    border: none;
    border-radius: var(--radius-lg);
    font-weight: var(--font-semibold);
    font-size: var(--text-base);
    padding: var(--space-4) var(--space-6);
    box-shadow: var(--shadow-md);
    transition: all var(--transition-base);
  }

  :deep(.el-button--primary:hover) {
    background: linear-gradient(135deg, var(--color-primary-600), var(--color-primary-700));
    box-shadow: var(--shadow-lg);
    transform: translateY(-2px);
  }

  :deep(.el-button--primary:active) {
    transform: translateY(0);
  }

  /* 第三方登录样式 */
  .third-party-login {
    padding: var(--space-6) var(--space-6) var(--space-4);
    text-align: center;
  }

  .third-party-title {
    margin-bottom: var(--space-6);
  }

  .third-party-title span {
    color: var(--color-gray-700);
    font-size: var(--text-base);
    font-weight: var(--font-medium);
  }

  .third-party-icons {
    display: flex;
    justify-content: center;
    align-items: center;
    gap: var(--space-4);
  }

  .oauth-icon-button {
    width: 60px;
    height: 60px;
    border-radius: 50%;
    background: linear-gradient(135deg, #f8f9fa, #e9ecef);
    border: 2px solid var(--color-gray-200);
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    transition: all var(--transition-base);
    box-shadow: var(--shadow-sm);
    position: relative;
    overflow: hidden;
  }

  .oauth-icon-button:hover {
    background: linear-gradient(135deg, #e9ecef, #dee2e6);
    border-color: var(--color-gray-300);
    box-shadow: var(--shadow-md);
    transform: translateY(-2px);
  }

  .oauth-icon-button.loading {
    pointer-events: none;
    opacity: 0.7;
  }

  .icon-wrapper {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 100%;
    height: 100%;
  }

  .linuxdo-logo {
    width: 32px;
    height: 32px;
    border-radius: 50%;
    overflow: hidden;
    position: relative;
    transition: all var(--transition-base);
  }

  .logo-top {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 33.33%;
    background-color: #2c3e50;
  }

  .logo-middle {
    position: absolute;
    top: 33.33%;
    left: 0;
    width: 100%;
    height: 33.34%;
    background-color: #ffffff;
  }

  .logo-bottom {
    position: absolute;
    top: 66.67%;
    left: 0;
    width: 100%;
    height: 33.33%;
    background-color: #ff8c00;
  }

  .linuxdo-icon:hover .linuxdo-logo {
    transform: scale(1.1);
  }

  .google-logo {
    width: 24px;
    height: 24px;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: all var(--transition-base);
  }

  .google-icon:hover .google-logo {
    transform: scale(1.1);
  }

  .microsoft-logo {
    width: 24px;
    height: 24px;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: all var(--transition-base);
  }

  .microsoft-icon:hover .microsoft-logo {
    transform: scale(1.1);
  }

  .loading-icon {
    color: var(--color-blue-500);
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    from {
      transform: rotate(0deg);
    }
    to {
      transform: rotate(360deg);
    }
  }

  .login-footer {
    text-align: center;
    margin-top: var(--space-6);
    padding: var(--space-5) var(--space-6) var(--space-6);
    border-top: 1px solid var(--color-gray-200);
    background: rgba(249, 250, 251, 0.5);
  }

  .login-footer p {
    color: var(--color-gray-600);
    font-size: var(--text-sm);
    margin: 0;
  }

  .register-link {
    color: var(--color-primary-600);
    text-decoration: none;
    font-weight: var(--font-semibold);
    transition: all var(--transition-base);
  }

  .register-link:hover {
    color: var(--color-primary-700);
    text-decoration: underline;
    text-underline-offset: 2px;
  }

  /* Responsive design */
  @media (max-width: 768px) {
    .login-container {
      padding: var(--space-4);
    }

    .login-card {
      max-width: 100%;
    }

    .login-form {
      padding: 0 var(--space-4);
    }

    .login-footer {
      padding: var(--space-4);
    }
  }
  </style>