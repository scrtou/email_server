<template>
    <div class="login-container">
      <el-card class="login-card">
        <template #header>
          <div class="login-header">
            <h2>账号管理系统</h2>
            <p>请登录您的账户</p>
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
  import { ref, reactive } from 'vue'
  import { useRouter } from 'vue-router'
  import { useAuthStore } from '@/stores/auth'
  
  export default {
    name: 'LoginPage',
    setup() {
      const router = useRouter()
      const authStore = useAuthStore()
      const loginFormRef = ref()
      
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
      
      return {
        loginFormRef,
        loginForm,
        loginRules,
        authStore,
        handleLogin
      }
    }
  }
  </script>
  
  <style scoped>
  .login-container {
    min-height: 100vh;
    width: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    background: linear-gradient(135deg, var(--color-primary-600) 0%, var(--color-primary-800) 50%, var(--color-gray-800) 100%);
    padding: var(--space-6);
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