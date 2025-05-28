<template>
    <div class="profile-page">
      <el-card>
        <template #header>
          <span class="card-title">个人信息</span>
        </template>
        
        <el-row :gutter="20">
          <!-- 个人信息编辑 -->
          <el-col :span="12">
            <el-form
              ref="profileFormRef"
              :model="profileForm"
              :rules="profileRules"
              label-width="100px"
            >
              <el-form-item label="用户名">
                <el-input v-model="profileForm.username" disabled>
                  <template #prefix>
                    <el-icon><User /></el-icon>
                  </template>
                </el-input>
              </el-form-item>
              
              <el-form-item label="邮箱" prop="email">
                <el-input v-model="profileForm.email">
                  <template #prefix>
                    <el-icon><Message /></el-icon>
                  </template>
                </el-input>
              </el-form-item>
              
              <el-form-item label="真实姓名" prop="real_name">
                <el-input v-model="profileForm.real_name">
                  <template #prefix>
                    <el-icon><User /></el-icon>
                  </template>
                </el-input>
              </el-form-item>
              
              <el-form-item label="手机号" prop="phone">
                <el-input v-model="profileForm.phone">
                  <template #prefix>
                    <el-icon><Phone /></el-icon>
                  </template>
                </el-input>
              </el-form-item>
              
              <el-form-item>
                <el-button type="primary" @click="updateProfile" :loading="isUpdating">
                  更新信息
                </el-button>
              </el-form-item>
            </el-form>
          </el-col>
          
          <!-- 账户信息 -->
          <el-col :span="12">
            <div class="account-info">
              <h3>账户信息</h3>
              <el-descriptions :column="1" border>
                <el-descriptions-item label="角色">
                  <el-tag :type="authStore.userRole === 'admin' ? 'danger' : 'primary'">
                    {{ authStore.userRole === 'admin' ? '管理员' : '普通用户' }}
                  </el-tag>
                </el-descriptions-item>
                <el-descriptions-item label="状态">
                  <el-tag type="success">正常</el-tag>
                </el-descriptions-item>
                <el-descriptions-item label="注册时间">
                  {{ formatDate(authStore.user?.created_at) }}
                </el-descriptions-item>
                <el-descriptions-item label="最后登录">
                  {{ formatDate(authStore.user?.last_login) }}
                </el-descriptions-item>
              </el-descriptions>
            </div>
          </el-col>
        </el-row>
      </el-card>
      
      <!-- 修改密码 -->
      <el-card style="margin-top: 20px;">
        <template #header>
          <span class="card-title">修改密码</span>
        </template>
        
        <el-row>
          <el-col :span="12">
            <el-form
              ref="passwordFormRef"
              :model="passwordForm"
              :rules="passwordRules"
              label-width="100px"
            >
              <el-form-item label="当前密码" prop="old_password">
                <el-input
                  v-model="passwordForm.old_password"
                  type="password"
                  show-password
                />
              </el-form-item>
              
              <el-form-item label="新密码" prop="new_password">
                <el-input
                  v-model="passwordForm.new_password"
                  type="password"
                  show-password
                />
              </el-form-item>
              
              <el-form-item label="确认密码" prop="confirm_password">
                <el-input
                  v-model="passwordForm.confirm_password"
                  type="password"
                  show-password
                />
              </el-form-item>
              
              <el-form-item>
                <el-button type="primary" @click="changePassword" :loading="isChangingPassword">
                  修改密码
                </el-button>
                <el-button @click="resetPasswordForm">重置</el-button>
              </el-form-item>
            </el-form>
          </el-col>
        </el-row>
      </el-card>
    </div>
  </template>
  
  <script>
  import { ref, reactive, onMounted } from 'vue'
  import { useAuthStore } from '@/stores/auth'
  import { User, Message, Phone } from '@element-plus/icons-vue'
  
  export default {
    name: 'ProfilePage',
    components: {
      User,
      Message,
      Phone
    },
    setup() {
      const authStore = useAuthStore()
      const profileFormRef = ref()
      const passwordFormRef = ref()
      const isUpdating = ref(false)
      const isChangingPassword = ref(false)
      
      const profileForm = reactive({
        username: '',
        email: '',
        real_name: '',
        phone: ''
      })
      
      const passwordForm = reactive({
        old_password: '',
        new_password: '',
        confirm_password: ''
      })
      
      const profileRules = {
        email: [
          { required: true, message: '请输入邮箱', trigger: 'blur' },
          { type: 'email', message: '请输入正确的邮箱', trigger: 'blur' }
        ],
        real_name: [
          { required: true, message: '请输入真实姓名', trigger: 'blur' }
        ]
      }
      
      const validateConfirmPassword = (rule, value, callback) => {
        if (value !== passwordForm.new_password) {
          callback(new Error('两次输入密码不一致'))
        } else {
          callback()
        }
      }
      
      const passwordRules = {
        old_password: [
          { required: true, message: '请输入当前密码', trigger: 'blur' }
        ],
        new_password: [
          { required: true, message: '请输入新密码', trigger: 'blur' },
          { min: 6, message: '密码长度不能少于6位', trigger: 'blur' }
        ],
        confirm_password: [
          { required: true, message: '请确认新密码', trigger: 'blur' },
          { validator: validateConfirmPassword, trigger: 'blur' }
        ]
      }
      
      const loadUserProfile = () => {
        if (authStore.user) {
          Object.assign(profileForm, {
            username: authStore.user.username,
            email: authStore.user.email,
            real_name: authStore.user.real_name,
            phone: authStore.user.phone
          })
        }
      }
      
      const updateProfile = async () => {
        if (!profileFormRef.value) return
        
        try {
          const valid = await profileFormRef.value.validate()
          if (!valid) return
          
          isUpdating.value = true
          const success = await authStore.updateProfile({
            email: profileForm.email,
            real_name: profileForm.real_name,
            phone: profileForm.phone
          })
          
          if (success) {
            loadUserProfile()
          }
        } catch (error) {
          console.error('更新个人信息失败:', error)
        } finally {
          isUpdating.value = false
        }
      }
      
      const changePassword = async () => {
        if (!passwordFormRef.value) return
        
        try {
          const valid = await passwordFormRef.value.validate()
          if (!valid) return
          
          isChangingPassword.value = true
          const success = await authStore.changePassword({
            old_password: passwordForm.old_password,
            new_password: passwordForm.new_password
          })
          
          if (success) {
            resetPasswordForm()
          }
        } catch (error) {
          console.error('修改密码失败:', error)
        } finally {
          isChangingPassword.value = false
        }
      }
      
      const resetPasswordForm = () => {
        Object.assign(passwordForm, {
          old_password: '',
          new_password: '',
          confirm_password: ''
        })
        if (passwordFormRef.value) {
          passwordFormRef.value.clearValidate()
        }
      }
      
      const formatDate = (date) => {
        if (!date) return '未知'
        return new Date(date).toLocaleString('zh-CN')
      }
      
      onMounted(() => {
        loadUserProfile()
      })
      
      return {
        authStore,
        profileFormRef,
        passwordFormRef,
        profileForm,
        passwordForm,
        profileRules,
        passwordRules,
        isUpdating,
        isChangingPassword,
        updateProfile,
        changePassword,
        resetPasswordForm,
        formatDate
      }
    }
  }
  </script>
  
  <style scoped>
  .profile-page {
    max-width: 1200px;
    margin: 0 auto;
  }
  
  .card-title {
    font-size: 16px;
    font-weight: 500;
    color: #303133;
  }
  
  .account-info h3 {
    margin-top: 0;
    margin-bottom: 16px;
    color: #303133;
  }
  </style>