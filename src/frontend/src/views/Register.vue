<template>
    <div class="register-container">
      <el-card class="register-card">
        <template #header>
          <div class="register-header">
            <h2>创建新账户</h2>
          </div>
        </template>
        
        <el-form
          ref="registerFormRef"
          :model="registerForm"
          :rules="registerRules"
          class="register-form"
          label-width="0"
          @submit.prevent="handleRegister"
        >
          <el-form-item prop="username">
            <el-input
              v-model="registerForm.username"
              placeholder="请输入用户名"
              size="large"
              prefix-icon="User"
              clearable
            />
          </el-form-item>
          
          <el-form-item prop="email">
            <el-input
              v-model="registerForm.email"
              type="email"
              placeholder="请输入邮箱地址"
              size="large"
              prefix-icon="Message"
              clearable
            />
          </el-form-item>
          
          <el-form-item prop="password">
            <el-input
              v-model="registerForm.password"
              type="password"
              placeholder="请输入密码"
              size="large"
              prefix-icon="Lock"
              show-password
            />
          </el-form-item>
          
          <el-form-item prop="confirmPassword">
            <el-input
              v-model="registerForm.confirmPassword"
              type="password"
              placeholder="请确认密码"
              size="large"
              prefix-icon="Lock"
              show-password
              @keyup.enter="handleRegister"
            />
          </el-form-item>
          
          <el-form-item>
            <el-button
              type="primary"
              size="large"
              style="width: 100%"
              :loading="authStore.isLoading"
              @click="handleRegister"
            >
              注册
            </el-button>
          </el-form-item>
        </el-form>
        
        <div class="register-footer">
          <p>
            已有账户？
            <router-link to="/login" class="login-link">立即登录</router-link>
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
    name: 'RegisterPage',
    setup() {
      const router = useRouter()
      const authStore = useAuthStore()
      const registerFormRef = ref()
      
      const registerForm = reactive({
        username: '',
        email: '',
        password: '',
        confirmPassword: ''
      })
      
      const validateConfirmPassword = (rule, value, callback) => {
        if (value !== registerForm.password) {
          callback(new Error('两次输入密码不一致'))
        } else {
          callback()
        }
      }
      
      const registerRules = {
        username: [
          { required: true, message: '请输入用户名', trigger: 'blur' },
          { min: 3, max: 20, message: '用户名长度在 3 到 20 个字符', trigger: 'blur' }
        ],
        email: [
          { required: true, message: '请输入邮箱地址', trigger: 'blur' },
          { type: 'email', message: '请输入正确的邮箱地址', trigger: 'blur' }
        ],
        password: [
          { required: true, message: '请输入密码', trigger: 'blur' },
          { min: 6, message: '密码长度不能少于6位', trigger: 'blur' }
        ],
        confirmPassword: [
          { required: true, message: '请确认密码', trigger: 'blur' },
          { validator: validateConfirmPassword, trigger: 'blur' }
        ]
      }
      
      const handleRegister = async () => {
        if (!registerFormRef.value) return
        
        try {
          const valid = await registerFormRef.value.validate()
          if (!valid) return
          
          // eslint-disable-next-line no-unused-vars
          const { confirmPassword, ...registerData } = registerForm
          const success = await authStore.register(registerData)
          if (success) {
            router.push('/')
          }
        } catch (error) {
          console.error('注册表单验证失败:', error)
        }
      }
      
      return {
        registerFormRef,
        registerForm,
        registerRules,
        authStore,
        handleRegister
      }
    }
  }
  </script>
  
  <style scoped>
.register-container {
  min-height: 100vh; /* 改为100vh，占满整个视窗高度 */
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, var(--color-primary-600) 0%, var(--color-primary-800) 50%, var(--color-gray-800) 100%);
  padding: var(--space-4); /* 减少或移除padding */
  position: relative;
  overflow: hidden;
  margin: 0; /* 确保没有外边距 */
}
  
  .register-container::before {
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
  
  .register-card {
    width: 100%;
    max-width: 450px; /* Changed from 420px (Login) and 500px (original Register) */
    border-radius: var(--radius-xl);
    box-shadow: var(--shadow-xl);
    background: rgba(255, 255, 255, 0.95);
    backdrop-filter: blur(20px);
    border: 1px solid rgba(255, 255, 255, 0.2);
    position: relative;
    z-index: 1;
    overflow: hidden;
  }
  
  .register-card::before {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    height: 4px;
    background: linear-gradient(90deg, var(--color-primary-500), var(--color-success-500), var(--color-warning-500));
  }
  
  .register-header {
    text-align: center;
    margin-bottom: var(--space-6);
    padding-top: var(--space-4);
  }
  
  .register-header h2 {
    color: var(--color-gray-800);
    margin-bottom: var(--space-2);
    font-size: var(--text-2xl);
    font-weight: var(--font-bold);
    background: linear-gradient(135deg, var(--color-primary-600), var(--color-primary-500));
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
    background-clip: text;
  }
  
  .register-header p {
    color: var(--color-gray-600);
    margin: 0;
    font-size: var(--text-sm);
    font-weight: var(--font-medium);
  }
  
  .register-form {
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
  
  .register-footer {
    text-align: center;
    margin-top: var(--space-6);
    padding: var(--space-5) var(--space-6) var(--space-6);
    border-top: 1px solid var(--color-gray-200);
    background: rgba(249, 250, 251, 0.5);
  }
  
  .register-footer p {
    color: var(--color-gray-600);
    font-size: var(--text-sm);
    margin: 0;
  }
  
  .login-link { /* This class is for the link to the login page */
    color: var(--color-primary-600);
    text-decoration: none;
    font-weight: var(--font-semibold);
    transition: all var(--transition-base);
  }
  
  .login-link:hover {
    color: var(--color-primary-700);
    text-decoration: underline;
    text-underline-offset: 2px;
  }
  
  /* Responsive design */
  @media (max-width: 768px) {
    .register-container {
      padding: var(--space-4);
    }
  
    .register-card {
      max-width: 100%;
    }
  
    .register-form {
      padding: 0 var(--space-4);
    }
  
    .register-footer {
      padding: var(--space-4);
    }
  }
  </style>