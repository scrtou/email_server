<template>
    <div class="register-container">
      <el-card class="register-card">
        <template #header>
          <div class="register-header">
            <h2>创建新账户</h2>
            <p>加入邮箱管理系统</p>
          </div>
        </template>
        
        <el-form
          ref="registerFormRef"
          :model="registerForm"
          :rules="registerRules"
          class="register-form"
          label-width="80px"
          @submit.prevent="handleRegister"
        >
          <el-form-item label="用户名" prop="username">
            <el-input
              v-model="registerForm.username"
              placeholder="请输入用户名"
              clearable
            />
          </el-form-item>
          
          <el-form-item label="邮箱" prop="email">
            <el-input
              v-model="registerForm.email"
              type="email"
              placeholder="请输入邮箱地址"
              clearable
            />
          </el-form-item>
          
          <el-form-item label="真实姓名" prop="real_name">
            <el-input
              v-model="registerForm.real_name"
              placeholder="请输入真实姓名"
              clearable
            />
          </el-form-item>
          
          <el-form-item label="手机号" prop="phone">
            <el-input
              v-model="registerForm.phone"
              placeholder="请输入手机号（可选）"
              clearable
            />
          </el-form-item>
          
          <el-form-item label="密码" prop="password">
            <el-input
              v-model="registerForm.password"
              type="password"
              placeholder="请输入密码"
              show-password
            />
          </el-form-item>
          
          <el-form-item label="确认密码" prop="confirmPassword">
            <el-input
              v-model="registerForm.confirmPassword"
              type="password"
              placeholder="请确认密码"
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
        real_name: '',
        phone: '',
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
        real_name: [
          { required: true, message: '请输入真实姓名', trigger: 'blur' }
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
    min-height: 100vh;
    display: flex;
    align-items: center;
    justify-content: center;
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    padding: 20px;
  }
  
  .register-card {
    width: 100%;
    max-width: 500px;
    border-radius: 10px;
    box-shadow: 0 10px 30px rgba(0, 0, 0, 0.2);
  }
  
  .register-header {
    text-align: center;
    margin-bottom: 20px;
  }
  
  .register-header h2 {
    color: #303133;
    margin-bottom: 8px;
  }
  
  .register-header p {
    color: #606266;
    margin: 0;
  }
  
  .register-form {
    padding: 0 10px;
  }
  
  .register-footer {
    text-align: center;
    margin-top: 20px;
    padding-top: 20px;
    border-top: 1px solid #EBEEF5;
  }
  
  .login-link {
    color: #409EFF;
    text-decoration: none;
  }
  
  .login-link:hover {
    text-decoration: underline;
  }
  </style>