<template>
    <div class="login-container">
      <el-card class="login-card">
        <template #header>
          <div class="login-header">
            <h2>邮箱管理系统</h2>
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
    display: flex;
    align-items: center;
    justify-content: center;
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    padding: 20px;
  }
  
  .login-card {
    width: 100%;
    max-width: 400px;
    border-radius: 10px;
    box-shadow: 0 10px 30px rgba(0, 0, 0, 0.2);
  }
  
  .login-header {
    text-align: center;
    margin-bottom: 20px;
  }
  
  .login-header h2 {
    color: #303133;
    margin-bottom: 8px;
  }
  
  .login-header p {
    color: #606266;
    margin: 0;
  }
  
  .login-form {
    padding: 0 10px;
  }
  
  .login-footer {
    text-align: center;
    margin-top: 20px;
    padding-top: 20px;
    border-top: 1px solid #EBEEF5;
  }
  
  .register-link {
    color: #409EFF;
    text-decoration: none;
  }
  
  .register-link:hover {
    text-decoration: underline;
  }
  </style>